package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/iancoleman/strcase"
	"go.einride.tech/aip/filtering"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	fieldmask_utils "github.com/mennanov/fieldmask-utils"

	"github.com/instill-ai/mgmt-backend/internal/resource"
	"github.com/instill-ai/mgmt-backend/pkg/constant"
	"github.com/instill-ai/mgmt-backend/pkg/datamodel"
	"github.com/instill-ai/mgmt-backend/pkg/logger"
	"github.com/instill-ai/mgmt-backend/pkg/middleware"
	"github.com/instill-ai/mgmt-backend/pkg/service"
	"github.com/instill-ai/mgmt-backend/pkg/usage"
	"github.com/instill-ai/x/sterr"

	custom_otel "github.com/instill-ai/mgmt-backend/pkg/logger/otel"
	healthcheckPB "github.com/instill-ai/protogen-go/common/healthcheck/v1alpha"
	mgmtPB "github.com/instill-ai/protogen-go/core/mgmt/v1alpha"
	checkfield "github.com/instill-ai/x/checkfield"
)

// TODO: Validate mask based on the field behavior. Currently, the fields are hard-coded.
// We stipulate that the ID of the user is IMMUTABLE
var createRequiredFields = []string{"id", "email", "newsletter_subscription"}
var outputOnlyFields = []string{"name", "type", "create_time", "update_time", "customer_id"}
var immutableFields = []string{"uid", "id"}

var createRequiredFieldsForToken = []string{"id"}
var outputOnlyFieldsForToken = []string{"name", "uid", "state", "token_type", "access_token", "create_time", "update_time"}

type PublicHandler struct {
	mgmtPB.UnimplementedMgmtPublicServiceServer
	Service      service.Service
	Usg          usage.Usage
	usageEnabled bool
}

// NewPublicHandler initiates a public handler instance
func NewPublicHandler(s service.Service, u usage.Usage, usageEnabled bool) mgmtPB.MgmtPublicServiceServer {
	return &PublicHandler{
		Service:      s,
		Usg:          u,
		usageEnabled: usageEnabled,
	}
}

var tracer = otel.Tracer("mgmt-backend.public-handler.tracer")

// Liveness checks the liveness of the server
func (h *PublicHandler) Liveness(ctx context.Context, in *mgmtPB.LivenessRequest) (*mgmtPB.LivenessResponse, error) {
	return &mgmtPB.LivenessResponse{
		HealthCheckResponse: &healthcheckPB.HealthCheckResponse{
			Status: healthcheckPB.HealthCheckResponse_SERVING_STATUS_SERVING,
		},
	}, nil
}

// Readiness checks the readiness of the server
func (h *PublicHandler) Readiness(ctx context.Context, in *mgmtPB.ReadinessRequest) (*mgmtPB.ReadinessResponse, error) {
	return &mgmtPB.ReadinessResponse{
		HealthCheckResponse: &healthcheckPB.HealthCheckResponse{
			Status: healthcheckPB.HealthCheckResponse_SERVING_STATUS_SERVING,
		},
	}, nil
}

// AuthTokenIssuer
func (h *PublicHandler) AuthTokenIssuer(ctx context.Context, in *mgmtPB.AuthTokenIssuerRequest) (*mgmtPB.AuthTokenIssuerResponse, error) {

	user, err := h.Service.GetUserByID(ctx, in.Username)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Unauthenticated request")
	}

	passwordHash, _, err := h.Service.GetUserPasswordHash(ctx, user.UID)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Unauthenticated request")
	}
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(in.Password))
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Unauthenticated request")
	}

	jti, _ := uuid.NewV4()
	exp := int32(time.Now().Unix()) + constant.DefaultJwtExpiration
	return &mgmtPB.AuthTokenIssuerResponse{
		AccessToken: &mgmtPB.AuthTokenIssuerResponse_UnsignedAccessToken{
			Aud: constant.DefaultJwtAudience,
			Sub: user.UID.String(),
			Iss: constant.DefaultJwtIssuer,
			Jti: jti.String(),
			Exp: exp,
		},
	}, nil
}

func (h *PublicHandler) AuthChangePassword(ctx context.Context, in *mgmtPB.AuthChangePasswordRequest) (*mgmtPB.AuthChangePasswordResponse, error) {
	userId, _, err := h.Service.GetCtxUser(ctx)

	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Unauthenticated request")
	}

	user, err := h.Service.GetUserByID(ctx, userId)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Unauthenticated request")
	}

	passwordHash, _, err := h.Service.GetUserPasswordHash(ctx, user.UID)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Unauthenticated request")
	}
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(in.OldPassword))
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Unauthenticated request")
	}

	passwordBytes, err := bcrypt.GenerateFromPassword([]byte(in.NewPassword), 10)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Update Password Failed")
	}

	err = h.Service.UpdateUserPasswordHash(ctx, user.UID, string(passwordBytes))
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Update Password Failed")
	}

	return &mgmtPB.AuthChangePasswordResponse{}, nil
}

func (h *PublicHandler) AuthLogout(ctx context.Context, in *mgmtPB.AuthLogoutRequest) (*mgmtPB.AuthLogoutResponse, error) {
	// TODO: implement this
	return &mgmtPB.AuthLogoutResponse{}, nil
}

func (h *PublicHandler) AuthLogin(ctx context.Context, in *mgmtPB.AuthLoginRequest) (*mgmtPB.AuthLoginResponse, error) {
	// This endpoint will be handled by KrakenD. We don't need to implement here
	return &mgmtPB.AuthLoginResponse{}, nil
}

func (h *PublicHandler) AuthValidateAccessToken(ctx context.Context, in *mgmtPB.AuthValidateAccessTokenRequest) (*mgmtPB.AuthValidateAccessTokenResponse, error) {
	// This endpoint will be handled by KrakenD. We don't need to implement here
	return &mgmtPB.AuthValidateAccessTokenResponse{}, nil
}

// GetUser returns the authenticated user
func (h *PublicHandler) GetUser(ctx context.Context) (*mgmtPB.User, error) {

	eventName := "GetUser"
	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	var dbUser *datamodel.User
	var err error

	// Verify if "jwt-sub" is in the header
	headerUserUId := middleware.GetRequestSingleHeader(ctx, constant.HeaderUserUIDKey)
	if headerUserUId != "" {
		uid, err := uuid.FromString(headerUserUId)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "Unauthenticated request")
		}
		dbUser, err = h.Service.GetUser(ctx, uid)
		if err != nil {
			sta := status.Convert(err)
			switch sta.Code() {
			case codes.InvalidArgument:
				st, e := sterr.CreateErrorBadRequest(
					"get user error", []*errdetails.BadRequest_FieldViolation{
						{
							Field:       "GetAuthenticatedUser",
							Description: sta.Message(),
						},
					})
				if e != nil {
					logger.Error(e.Error())
				}
				return nil, st.Err()
			default:
				st, e := sterr.CreateErrorResourceInfo(
					sta.Code(),
					"get user error",
					"user",
					fmt.Sprintf("uid %s", headerUserUId),
					"",
					sta.Message(),
				)
				if e != nil {
					logger.Error(e.Error())
				}
				return nil, st.Err()
			}
		}
	} else {
		// Verify "user-id" in the header if there is no "jwt-sub"
		headerUserId := middleware.GetRequestSingleHeader(ctx, constant.HeaderUserIDKey)

		dbUser, err = h.Service.GetUserByID(ctx, headerUserId)
		if err != nil {
			sta := status.Convert(err)
			switch sta.Code() {
			case codes.InvalidArgument:
				st, e := sterr.CreateErrorBadRequest(
					"get user error", []*errdetails.BadRequest_FieldViolation{
						{
							Field:       "GetAuthenticatedUser",
							Description: sta.Message(),
						},
					})
				if e != nil {
					logger.Error(e.Error())
				}
				return nil, st.Err()
			default:
				st, e := sterr.CreateErrorResourceInfo(
					sta.Code(),
					"get user error",
					"user",
					fmt.Sprintf("id %s", headerUserId),
					"",
					sta.Message(),
				)
				if e != nil {
					logger.Error(e.Error())
				}
				return nil, st.Err()
			}

		}
	}

	pbUser, err := datamodel.DBUser2PBUser(dbUser)
	if err != nil {
		logger.Error(err.Error())
		st, e := sterr.CreateErrorResourceInfo(
			codes.Internal,
			"get user error",
			"user",
			fmt.Sprintf("id %s", dbUser.ID),
			"",
			err.Error(),
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return nil, st.Err()
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		pbUser,
		eventName,
		custom_otel.SetEventResource(dbUser),
	)))

	return pbUser, nil
}

// QueryAuthenticatedUser gets the authenticated user.
// Note: this endpoint assumes the ID of the authenticated user is the default user.
func (h *PublicHandler) QueryAuthenticatedUser(ctx context.Context, req *mgmtPB.QueryAuthenticatedUserRequest) (*mgmtPB.QueryAuthenticatedUserResponse, error) {

	eventName := "QueryAuthenticatedUser"
	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	pbUser, err := h.GetUser(ctx)
	if err != nil {
		return &mgmtPB.QueryAuthenticatedUserResponse{}, err
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		pbUser,
		eventName,
		custom_otel.SetEventResource(pbUser),
	)))

	resp := mgmtPB.QueryAuthenticatedUserResponse{
		User: pbUser,
	}
	return &resp, nil
}

// PatchAuthenticatedUser updates the authenticated user.
// Note: this endpoint assumes the ID of the authenticated user is the default user.
func (h *PublicHandler) PatchAuthenticatedUser(ctx context.Context, req *mgmtPB.PatchAuthenticatedUserRequest) (*mgmtPB.PatchAuthenticatedUserResponse, error) {

	eventName := "PatchAuthenticatedUser"
	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	reqUser := req.GetUser()

	// Validate the field mask
	if !req.GetUpdateMask().IsValid(reqUser) {
		st, e := sterr.CreateErrorBadRequest(
			"update user invalid fieldmask error", []*errdetails.BadRequest_FieldViolation{
				{
					Field:       "PatchAuthenticatedUserRequest.update_mask",
					Description: "invalid",
				},
			},
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return &mgmtPB.PatchAuthenticatedUserResponse{}, st.Err()
	}

	reqFieldMask, err := checkfield.CheckUpdateOutputOnlyFields(req.GetUpdateMask(), outputOnlyFields)
	if err != nil {
		st, e := sterr.CreateErrorBadRequest(
			"update user update OUTPUT_ONLY fields error", []*errdetails.BadRequest_FieldViolation{
				{
					Field:       "PatchAuthenticatedUserRequest OUTPUT_ONLY fields",
					Description: err.Error(),
				},
			},
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return &mgmtPB.PatchAuthenticatedUserResponse{}, st.Err()
	}

	mask, err := fieldmask_utils.MaskFromProtoFieldMask(reqFieldMask, strcase.ToCamel)
	if err != nil {
		logger.Error(err.Error())
		st, e := sterr.CreateErrorBadRequest(
			"update user update mask error", []*errdetails.BadRequest_FieldViolation{
				{
					Field:       "PatchAuthenticatedUserRequest.update_mask",
					Description: err.Error(),
				},
			},
		)
		if e != nil {
			logger.Error(e.Error())
		}

		return &mgmtPB.PatchAuthenticatedUserResponse{}, st.Err()
	}

	// Get current authenticated user
	GResp, err := h.QueryAuthenticatedUser(ctx, &mgmtPB.QueryAuthenticatedUserRequest{})

	if err != nil {
		return &mgmtPB.PatchAuthenticatedUserResponse{}, err
	}
	pbUserToUpdate := GResp.GetUser()

	if mask.IsEmpty() {
		// return the un-changed user `pbUserToUpdate`
		resp := mgmtPB.PatchAuthenticatedUserResponse{
			User: pbUserToUpdate,
		}
		return &resp, nil
	}

	// the current user `pbUserToUpdate`: a struct to copy to
	uid, err := uuid.FromString(pbUserToUpdate.GetUid())
	if err != nil {
		logger.Error(err.Error())
		st, e := sterr.CreateErrorResourceInfo(
			codes.Internal,
			"update authenticated user error",
			"user",
			fmt.Sprintf("user %v", pbUserToUpdate),
			"",
			err.Error(),
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return &mgmtPB.PatchAuthenticatedUserResponse{}, st.Err()
	}

	// Handle immutable fields from the update mask
	err = checkfield.CheckUpdateImmutableFields(reqUser, pbUserToUpdate, immutableFields)
	if err != nil {
		st, e := sterr.CreateErrorBadRequest(
			"update authenticated user update IMMUTABLE fields error", []*errdetails.BadRequest_FieldViolation{
				{
					Field:       "PatchAuthenticatedUserRequest IMMUTABLE fields",
					Description: err.Error(),
				},
			},
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return &mgmtPB.PatchAuthenticatedUserResponse{}, st.Err()
	}

	// Only the fields mentioned in the field mask will be copied to `pbUserToUpdate`, other fields are left intact
	err = fieldmask_utils.StructToStruct(mask, reqUser, pbUserToUpdate)
	if err != nil {
		logger.Error(err.Error())
		st, e := sterr.CreateErrorResourceInfo(
			codes.Internal,
			"update authenticated user error", "user", fmt.Sprintf("uid %s", *reqUser.Uid),
			"",
			err.Error(),
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return &mgmtPB.PatchAuthenticatedUserResponse{}, st.Err()
	}

	dbUserToUpd, err := datamodel.PBUser2DBUser(pbUserToUpdate)
	if err != nil {
		logger.Error(err.Error())
		st, e := sterr.CreateErrorResourceInfo(
			codes.Internal,
			"update authenticated user error",
			"user",
			fmt.Sprintf("id %s", pbUserToUpdate.GetId()),
			"",
			err.Error(),
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return &mgmtPB.PatchAuthenticatedUserResponse{}, st.Err()
	}

	dbUserUpdated, err := h.Service.UpdateUser(ctx, uid, dbUserToUpd)
	if err != nil {
		sta := status.Convert(err)
		switch sta.Code() {
		case codes.InvalidArgument:
			st, e := sterr.CreateErrorBadRequest(
				"update authenticated user error", []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "PatchAuthenticatedUserRequest",
						Description: sta.Message(),
					},
				})
			if e != nil {
				logger.Error(e.Error())
			}
			return &mgmtPB.PatchAuthenticatedUserResponse{}, st.Err()
		default:
			st, e := sterr.CreateErrorResourceInfo(
				sta.Code(),
				"update authenticated user error",
				"user",
				fmt.Sprintf("uid %s", uid.String()),
				"",
				sta.Message(),
			)
			if e != nil {
				logger.Error(e.Error())
			}
			return &mgmtPB.PatchAuthenticatedUserResponse{}, st.Err()
		}
	}

	pbUserUpdated, err := datamodel.DBUser2PBUser(dbUserUpdated)
	if err != nil {
		logger.Error(err.Error())
		st, e := sterr.CreateErrorResourceInfo(
			codes.Internal,
			"get authenticated user error",
			"user",
			fmt.Sprintf("uid %s", dbUserUpdated.UID),
			"",
			err.Error(),
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return &mgmtPB.PatchAuthenticatedUserResponse{}, st.Err()
	}
	resp := mgmtPB.PatchAuthenticatedUserResponse{
		User: pbUserUpdated,
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		pbUserUpdated,
		eventName,
		custom_otel.SetEventResource(dbUserUpdated),
	)))

	// Trigger single reporter right after user updated
	if h.usageEnabled && h.Usg != nil {
		h.Usg.TriggerSingleReporter(context.Background())
	}

	return &resp, nil
}

// ExistUsername verifies if a username (ID) has been occupied
func (h *PublicHandler) ExistUsername(ctx context.Context, req *mgmtPB.ExistUsernameRequest) (*mgmtPB.ExistUsernameResponse, error) {

	eventName := "ExistUsername"
	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	id := strings.TrimPrefix(req.GetName(), "users/")

	// Validate the user id conforms to RFC-1034, which restricts to letters, numbers,
	// and hyphen, with the first character a letter, the last a letter or a
	// number, and a 63 character maximum.
	err := checkfield.CheckResourceID(id)
	if err != nil {
		st, e := sterr.CreateErrorBadRequest(
			"verify whether username is occupied bad request error", []*errdetails.BadRequest_FieldViolation{
				{
					Field:       "id",
					Description: err.Error(),
				},
			},
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return &mgmtPB.ExistUsernameResponse{}, st.Err()
	}

	dbUser, err := h.Service.GetUserByID(ctx, id)
	if err != nil {
		sta := status.Convert(err)
		switch sta.Code() {
		// user not exist - username not occupied
		case codes.NotFound:
			resp := mgmtPB.ExistUsernameResponse{
				Exists: false,
			}
			return &resp, nil
		// invalid username
		case codes.InvalidArgument:
			st, e := sterr.CreateErrorBadRequest(
				"verify whether username is occupied error", []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "ExistUsernameRequest.name",
						Description: sta.Message(),
					},
				})
			if e != nil {
				logger.Error(e.Error())
			}
			return &mgmtPB.ExistUsernameResponse{}, st.Err()
		default:
			st, e := sterr.CreateErrorResourceInfo(
				sta.Code(),
				"verify whether username is occupied error",
				"user",
				fmt.Sprintf("id %s", id),
				"",
				sta.Message(),
			)
			if e != nil {
				logger.Error(e.Error())
			}
			return &mgmtPB.ExistUsernameResponse{}, st.Err()
		}
	}

	pbUser, err := datamodel.DBUser2PBUser(dbUser)
	if err != nil {
		logger.Error(err.Error())
		st, e := sterr.CreateErrorResourceInfo(
			codes.Internal,
			"get user error",
			"user",
			fmt.Sprintf("id %s", dbUser.ID),
			"",
			err.Error(),
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return nil, st.Err()
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		pbUser,
		eventName,
		custom_otel.SetEventResource(dbUser),
	)))

	resp := mgmtPB.ExistUsernameResponse{
		Exists: true,
	}
	return &resp, nil
}

// CreateToken creates an API token for triggering pipelines. This endpoint is not supported yet.
func (h *PublicHandler) CreateToken(ctx context.Context, req *mgmtPB.CreateTokenRequest) (*mgmtPB.CreateTokenResponse, error) {

	eventName := "CreateToken"
	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	// Set all OUTPUT_ONLY fields to zero value on the requested payload token resource
	if err := checkfield.CheckCreateOutputOnlyFields(req.Token, outputOnlyFieldsForToken); err != nil {
		return &mgmtPB.CreateTokenResponse{}, status.Errorf(codes.InvalidArgument, err.Error())
	}

	// Return error if REQUIRED fields are not provided in the requested payload token resource
	if err := checkfield.CheckRequiredFields(req.Token, createRequiredFieldsForToken); err != nil {
		return &mgmtPB.CreateTokenResponse{}, status.Errorf(codes.InvalidArgument, err.Error())
	}

	// Return error if resource ID does not follow RFC-1034
	if err := checkfield.CheckResourceID(req.Token.GetId()); err != nil {
		return &mgmtPB.CreateTokenResponse{}, status.Errorf(codes.InvalidArgument, err.Error())
	}

	// Return error if expiration is not provided
	if req.Token.GetExpiration() == nil {
		return &mgmtPB.CreateTokenResponse{}, status.Errorf(codes.InvalidArgument, "no expiration info")
	}

	owner, err := h.GetUser(ctx)
	if err != nil {
		return &mgmtPB.CreateTokenResponse{}, status.Errorf(codes.Unauthenticated, "Unauthorized")
	}
	ownerPermalink := "users/" + owner.GetUid()

	_, err = h.Service.GetToken(ctx, req.Token.Id, ownerPermalink)
	if err == nil {
		return &mgmtPB.CreateTokenResponse{}, status.Errorf(codes.AlreadyExists, "Token ID already existed")
	}

	dbToken := datamodel.PBToken2DBToken(ctx, req.Token)
	dbToken.AccessToken = datamodel.GenerateToken()
	dbToken.Owner = ownerPermalink
	curTime := time.Now()
	dbToken.CreateTime = curTime
	dbToken.UpdateTime = curTime
	dbToken.State = datamodel.TokenState(mgmtPB.ApiToken_STATE_ACTIVE)

	switch req.Token.GetExpiration().(type) {
	case *mgmtPB.ApiToken_Ttl:
		if req.Token.GetTtl() >= 0 {
			dbToken.ExpireTime = curTime.Add(time.Second * time.Duration(req.Token.GetTtl()))
		} else if req.Token.GetTtl() == -1 {
			dbToken.ExpireTime = time.Date(2099, 12, 31, 0, 0, 0, 0, time.Now().UTC().Location())
		} else {
			return &mgmtPB.CreateTokenResponse{}, status.Errorf(codes.InvalidArgument, "ttl should >= -1")
		}
	case *mgmtPB.ApiToken_ExpireTime:
		dbToken.ExpireTime = req.Token.GetExpireTime().AsTime()
	}

	dbToken.TokenType = constant.DefaultTokenType
	createErr := h.Service.CreateToken(ctx, dbToken)
	if createErr != nil {
		return &mgmtPB.CreateTokenResponse{}, status.Errorf(codes.AlreadyExists, createErr.Error())
	}

	dbCreatedToken, err := h.Service.GetToken(ctx, req.Token.Id, ownerPermalink)
	if createErr != nil {
		return &mgmtPB.CreateTokenResponse{}, status.Errorf(codes.AlreadyExists, err.Error())
	}

	resp := &mgmtPB.CreateTokenResponse{
		Token: datamodel.DBToken2PBToken(dbCreatedToken),
	}

	// Manually set the custom header to have a StatusCreated http response for REST endpoint
	if err := grpc.SetHeader(ctx, metadata.Pairs("x-http-code", strconv.Itoa(http.StatusCreated))); err != nil {
		return nil, err
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		owner,
		eventName,
		custom_otel.SetEventResult(fmt.Sprintf("Total records retrieved: %v", dbToken)),
	)))

	return resp, nil
}

// ListTokens lists all the API tokens of the authenticated user. This endpoint is not supported yet.
func (h *PublicHandler) ListTokens(ctx context.Context, req *mgmtPB.ListTokensRequest) (*mgmtPB.ListTokensResponse, error) {

	eventName := "ListTokens"
	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	owner, err := h.GetUser(ctx)
	if err != nil {
		return &mgmtPB.ListTokensResponse{}, err
	}

	ownerPermalink := "users/" + owner.GetUid()

	dbTokens, totalSize, nextPageToken, err := h.Service.ListTokens(ctx, int64(req.GetPageSize()), req.GetPageToken(), ownerPermalink)
	if err != nil {
		return &mgmtPB.ListTokensResponse{}, err
	}

	pbTokens := []*mgmtPB.ApiToken{}
	for _, dbToken := range dbTokens {
		pbTokens = append(pbTokens, datamodel.DBToken2PBToken(&dbToken))
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		owner,
		eventName,
	)))

	resp := &mgmtPB.ListTokensResponse{
		Tokens:        pbTokens,
		NextPageToken: nextPageToken,
		TotalSize:     int32(totalSize),
	}
	return resp, nil
}

// GetToken gets an API token of the authenticated user. This endpoint is not supported yet.
func (h *PublicHandler) GetToken(ctx context.Context, req *mgmtPB.GetTokenRequest) (*mgmtPB.GetTokenResponse, error) {

	eventName := "GetToken"
	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	owner, err := h.GetUser(ctx)
	if err != nil {
		return &mgmtPB.GetTokenResponse{}, err
	}

	ownerPermalink := "users/" + owner.GetUid()
	id, err := resource.GetRscNameID(req.GetName())
	if err != nil {
		return &mgmtPB.GetTokenResponse{}, err
	}

	dbToken, err := h.Service.GetToken(ctx, id, ownerPermalink)
	if err != nil {
		return &mgmtPB.GetTokenResponse{}, err
	}

	pbToken := datamodel.DBToken2PBToken(dbToken)

	resp := &mgmtPB.GetTokenResponse{
		Token: pbToken,
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		owner,
		eventName,
		custom_otel.SetEventResource(dbToken),
	)))

	return resp, nil
}

// DeleteToken deletes an API token of the authenticated user. This endpoint is not supported yet.
func (h *PublicHandler) DeleteToken(ctx context.Context, req *mgmtPB.DeleteTokenRequest) (*mgmtPB.DeleteTokenResponse, error) {

	eventName := "DeleteToken"
	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	owner, err := h.GetUser(ctx)
	if err != nil {
		return &mgmtPB.DeleteTokenResponse{}, err
	}
	ownerPermalink := "users/" + owner.GetUid()

	existToken, err := h.GetToken(ctx, &mgmtPB.GetTokenRequest{Name: req.GetName()})
	if err != nil {
		return &mgmtPB.DeleteTokenResponse{}, err
	}

	if err := h.Service.DeleteToken(ctx, existToken.Token.GetId(), ownerPermalink); err != nil {
		return &mgmtPB.DeleteTokenResponse{}, err
	}

	// We need to manually set the custom header to have a StatusCreated http response for REST endpoint
	if err := grpc.SetHeader(ctx, metadata.Pairs("x-http-code", strconv.Itoa(http.StatusNoContent))); err != nil {
		return &mgmtPB.DeleteTokenResponse{}, err
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		owner,
		eventName,
		custom_otel.SetEventResource(existToken.GetToken()),
	)))

	return &mgmtPB.DeleteTokenResponse{}, nil
}

// ValidateToken validate the token
func (h *PublicHandler) ValidateToken(ctx context.Context, req *mgmtPB.ValidateTokenRequest) (*mgmtPB.ValidateTokenResponse, error) {

	eventName := "ValidateToken"
	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	authorization := resource.GetRequestSingleHeader(ctx, constant.HeaderAuthorization)
	apiToken := strings.Replace(authorization, "Bearer ", "", 1)

	userUid, err := h.Service.ValidateToken(apiToken)

	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, err
	}

	return &mgmtPB.ValidateTokenResponse{UserUid: userUid}, nil
}

func (h *PublicHandler) ListPipelineTriggerRecords(ctx context.Context, req *mgmtPB.ListPipelineTriggerRecordsRequest) (*mgmtPB.ListPipelineTriggerRecordsResponse, error) {

	eventName := "ListPipelineTriggerRecords"
	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	pbUser, err := h.GetUser(ctx)
	if err != nil {
		span.SetStatus(1, err.Error())
		return &mgmtPB.ListPipelineTriggerRecordsResponse{}, err
	}

	var mode mgmtPB.Mode
	var status mgmtPB.Status

	declarations, err := filtering.NewDeclarations([]filtering.DeclarationOption{
		filtering.DeclareStandardFunctions(),
		filtering.DeclareIdent(constant.Start, filtering.TypeTimestamp),
		filtering.DeclareIdent(constant.Stop, filtering.TypeTimestamp),
		filtering.DeclareIdent(constant.PipelineID, filtering.TypeString),
		filtering.DeclareIdent(constant.PipelineUID, filtering.TypeString),
		filtering.DeclareIdent(constant.PipelineReleaseID, filtering.TypeString),
		filtering.DeclareIdent(constant.PipelineReleaseUID, filtering.TypeString),
		filtering.DeclareEnumIdent(constant.TriggerMode, mode.Type()),
		filtering.DeclareEnumIdent(constant.Status, status.Type()),
	}...)
	if err != nil {
		span.SetStatus(1, err.Error())
		return &mgmtPB.ListPipelineTriggerRecordsResponse{}, err
	}

	filter, err := filtering.ParseFilter(req, declarations)
	if err != nil {
		span.SetStatus(1, err.Error())
		return &mgmtPB.ListPipelineTriggerRecordsResponse{}, err
	}

	pipelineTriggerRecords, totalSize, nextPageToken, err := h.Service.ListPipelineTriggerRecords(ctx, pbUser, int64(req.GetPageSize()), req.GetPageToken(), filter)
	if err != nil {
		span.SetStatus(1, err.Error())
		return &mgmtPB.ListPipelineTriggerRecordsResponse{}, err
	}

	resp := mgmtPB.ListPipelineTriggerRecordsResponse{
		PipelineTriggerRecords: pipelineTriggerRecords,
		NextPageToken:          nextPageToken,
		TotalSize:              int32(totalSize),
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		pbUser,
		eventName,
		custom_otel.SetEventResult(fmt.Sprintf("Total records retrieved: %v", totalSize)),
	)))

	return &resp, nil
}

func (h *PublicHandler) ListPipelineTriggerTableRecords(ctx context.Context, req *mgmtPB.ListPipelineTriggerTableRecordsRequest) (*mgmtPB.ListPipelineTriggerTableRecordsResponse, error) {

	eventName := "ListPipelineTriggerTableRecords"
	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	pbUser, err := h.GetUser(ctx)
	if err != nil {
		span.SetStatus(1, err.Error())
		return &mgmtPB.ListPipelineTriggerTableRecordsResponse{}, err
	}

	declarations, err := filtering.NewDeclarations([]filtering.DeclarationOption{
		filtering.DeclareStandardFunctions(),
		filtering.DeclareIdent(constant.Start, filtering.TypeTimestamp),
		filtering.DeclareIdent(constant.Stop, filtering.TypeTimestamp),
		filtering.DeclareIdent(constant.PipelineID, filtering.TypeString),
		filtering.DeclareIdent(constant.PipelineUID, filtering.TypeString),
		filtering.DeclareIdent(constant.PipelineReleaseID, filtering.TypeString),
		filtering.DeclareIdent(constant.PipelineReleaseUID, filtering.TypeString),
	}...)
	if err != nil {
		span.SetStatus(1, err.Error())
		return &mgmtPB.ListPipelineTriggerTableRecordsResponse{}, err
	}

	filter, err := filtering.ParseFilter(req, declarations)
	if err != nil {
		span.SetStatus(1, err.Error())
		return &mgmtPB.ListPipelineTriggerTableRecordsResponse{}, err
	}

	pipelineTriggerTableRecords, totalSize, nextPageToken, err := h.Service.ListPipelineTriggerTableRecords(ctx, pbUser, int64(req.GetPageSize()), req.GetPageToken(), filter)
	if err != nil {
		span.SetStatus(1, err.Error())
		return &mgmtPB.ListPipelineTriggerTableRecordsResponse{}, err
	}

	resp := mgmtPB.ListPipelineTriggerTableRecordsResponse{
		PipelineTriggerTableRecords: pipelineTriggerTableRecords,
		NextPageToken:               nextPageToken,
		TotalSize:                   int32(totalSize),
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		pbUser,
		eventName,
		custom_otel.SetEventResult(fmt.Sprintf("Total records retrieved: %v", totalSize)),
	)))

	return &resp, nil
}

func (h *PublicHandler) ListPipelineTriggerChartRecords(ctx context.Context, req *mgmtPB.ListPipelineTriggerChartRecordsRequest) (*mgmtPB.ListPipelineTriggerChartRecordsResponse, error) {

	eventName := "ListPipelineTriggerChartRecords"
	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	pbUser, err := h.GetUser(ctx)
	if err != nil {
		span.SetStatus(1, err.Error())
		return &mgmtPB.ListPipelineTriggerChartRecordsResponse{}, err
	}

	var mode mgmtPB.Mode
	var status mgmtPB.Status

	declarations, err := filtering.NewDeclarations([]filtering.DeclarationOption{
		filtering.DeclareStandardFunctions(),
		filtering.DeclareIdent(constant.Start, filtering.TypeTimestamp),
		filtering.DeclareIdent(constant.Stop, filtering.TypeTimestamp),
		filtering.DeclareIdent(constant.PipelineID, filtering.TypeString),
		filtering.DeclareIdent(constant.PipelineUID, filtering.TypeString),
		filtering.DeclareIdent(constant.PipelineReleaseID, filtering.TypeString),
		filtering.DeclareIdent(constant.PipelineReleaseUID, filtering.TypeString),
		filtering.DeclareEnumIdent(constant.TriggerMode, mode.Type()),
		filtering.DeclareEnumIdent(constant.Status, status.Type()),
	}...)
	if err != nil {
		span.SetStatus(1, err.Error())
		return &mgmtPB.ListPipelineTriggerChartRecordsResponse{}, err
	}

	filter, err := filtering.ParseFilter(req, declarations)
	if err != nil {
		span.SetStatus(1, err.Error())
		return &mgmtPB.ListPipelineTriggerChartRecordsResponse{}, err
	}

	pipelineTriggerChartRecords, err := h.Service.ListPipelineTriggerChartRecords(ctx, pbUser, int64(req.GetAggregationWindow()), filter)
	if err != nil {
		span.SetStatus(1, err.Error())
		return &mgmtPB.ListPipelineTriggerChartRecordsResponse{}, err
	}

	resp := mgmtPB.ListPipelineTriggerChartRecordsResponse{
		PipelineTriggerChartRecords: pipelineTriggerChartRecords,
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		pbUser,
		eventName,
	)))

	return &resp, nil
}

func (h *PublicHandler) ListConnectorExecuteRecords(ctx context.Context, req *mgmtPB.ListConnectorExecuteRecordsRequest) (*mgmtPB.ListConnectorExecuteRecordsResponse, error) {

	eventName := "ListConnectorExecuteRecords"
	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	pbUser, err := h.GetUser(ctx)
	if err != nil {
		span.SetStatus(1, err.Error())
		return &mgmtPB.ListConnectorExecuteRecordsResponse{}, err
	}

	var status mgmtPB.Status

	declarations, err := filtering.NewDeclarations([]filtering.DeclarationOption{
		filtering.DeclareStandardFunctions(),
		filtering.DeclareIdent(constant.Start, filtering.TypeTimestamp),
		filtering.DeclareIdent(constant.Stop, filtering.TypeTimestamp),
		filtering.DeclareIdent(constant.ConnectorID, filtering.TypeString),
		filtering.DeclareIdent(constant.ConnectorUID, filtering.TypeString),
		filtering.DeclareIdent(constant.PipelineID, filtering.TypeString),
		filtering.DeclareIdent(constant.PipelineUID, filtering.TypeString),
		filtering.DeclareIdent(constant.PipelineReleaseID, filtering.TypeString),
		filtering.DeclareIdent(constant.PipelineReleaseUID, filtering.TypeString),
		filtering.DeclareEnumIdent(constant.Status, status.Type()),
	}...)
	if err != nil {
		span.SetStatus(1, err.Error())
		return &mgmtPB.ListConnectorExecuteRecordsResponse{}, err
	}

	filter, err := filtering.ParseFilter(req, declarations)
	if err != nil {
		span.SetStatus(1, err.Error())
		return &mgmtPB.ListConnectorExecuteRecordsResponse{}, err
	}

	connectorExecuteRecords, totalSize, nextPageToken, err := h.Service.ListConnectorExecuteRecords(ctx, pbUser, int64(req.GetPageSize()), req.GetPageToken(), filter)
	if err != nil {
		span.SetStatus(1, err.Error())
		return &mgmtPB.ListConnectorExecuteRecordsResponse{}, err
	}

	resp := mgmtPB.ListConnectorExecuteRecordsResponse{
		ConnectorExecuteRecords: connectorExecuteRecords,
		NextPageToken:           nextPageToken,
		TotalSize:               int32(totalSize),
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		pbUser,
		eventName,
		custom_otel.SetEventResult(fmt.Sprintf("Total records retrieved: %v", totalSize)),
	)))

	return &resp, nil
}

func (h *PublicHandler) ListConnectorExecuteTableRecords(ctx context.Context, req *mgmtPB.ListConnectorExecuteTableRecordsRequest) (*mgmtPB.ListConnectorExecuteTableRecordsResponse, error) {

	eventName := "ListConnectorExecuteTableRecords"
	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	pbUser, err := h.GetUser(ctx)
	if err != nil {
		span.SetStatus(1, err.Error())
		return &mgmtPB.ListConnectorExecuteTableRecordsResponse{}, err
	}

	declarations, err := filtering.NewDeclarations([]filtering.DeclarationOption{
		filtering.DeclareStandardFunctions(),
		filtering.DeclareIdent(constant.Start, filtering.TypeTimestamp),
		filtering.DeclareIdent(constant.Stop, filtering.TypeTimestamp),
		filtering.DeclareIdent(constant.ConnectorID, filtering.TypeString),
		filtering.DeclareIdent(constant.ConnectorUID, filtering.TypeString),
	}...)
	if err != nil {
		span.SetStatus(1, err.Error())
		return &mgmtPB.ListConnectorExecuteTableRecordsResponse{}, err
	}

	filter, err := filtering.ParseFilter(req, declarations)
	if err != nil {
		span.SetStatus(1, err.Error())
		return &mgmtPB.ListConnectorExecuteTableRecordsResponse{}, err
	}

	connectorExecuteTableRecords, totalSize, nextPageToken, err := h.Service.ListConnectorExecuteTableRecords(ctx, pbUser, int64(req.GetPageSize()), req.GetPageToken(), filter)
	if err != nil {
		span.SetStatus(1, err.Error())
		return &mgmtPB.ListConnectorExecuteTableRecordsResponse{}, err
	}

	resp := mgmtPB.ListConnectorExecuteTableRecordsResponse{
		ConnectorExecuteTableRecords: connectorExecuteTableRecords,
		NextPageToken:                nextPageToken,
		TotalSize:                    int32(totalSize),
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		pbUser,
		eventName,
		custom_otel.SetEventResult(fmt.Sprintf("Total records retrieved: %v", totalSize)),
	)))

	return &resp, nil
}

func (h *PublicHandler) ListConnectorExecuteChartRecords(ctx context.Context, req *mgmtPB.ListConnectorExecuteChartRecordsRequest) (*mgmtPB.ListConnectorExecuteChartRecordsResponse, error) {

	eventName := "ListConnectorExecuteChartRecords"
	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	pbUser, err := h.GetUser(ctx)
	if err != nil {
		span.SetStatus(1, err.Error())
		return &mgmtPB.ListConnectorExecuteChartRecordsResponse{}, err
	}

	var status mgmtPB.Status

	declarations, err := filtering.NewDeclarations([]filtering.DeclarationOption{
		filtering.DeclareStandardFunctions(),
		filtering.DeclareIdent(constant.Start, filtering.TypeTimestamp),
		filtering.DeclareIdent(constant.Stop, filtering.TypeTimestamp),
		filtering.DeclareIdent(constant.ConnectorID, filtering.TypeString),
		filtering.DeclareIdent(constant.ConnectorUID, filtering.TypeString),
		filtering.DeclareIdent(constant.PipelineID, filtering.TypeString),
		filtering.DeclareIdent(constant.PipelineUID, filtering.TypeString),
		filtering.DeclareIdent(constant.PipelineReleaseID, filtering.TypeString),
		filtering.DeclareIdent(constant.PipelineReleaseUID, filtering.TypeString),
		filtering.DeclareEnumIdent(constant.Status, status.Type()),
	}...)
	if err != nil {
		span.SetStatus(1, err.Error())
		return &mgmtPB.ListConnectorExecuteChartRecordsResponse{}, err
	}

	filter, err := filtering.ParseFilter(req, declarations)
	if err != nil {
		span.SetStatus(1, err.Error())
		return &mgmtPB.ListConnectorExecuteChartRecordsResponse{}, err
	}

	connectorExecuteChartRecords, err := h.Service.ListConnectorExecuteChartRecords(ctx, pbUser, int64(req.GetAggregationWindow()), filter)
	if err != nil {
		span.SetStatus(1, err.Error())
		return &mgmtPB.ListConnectorExecuteChartRecordsResponse{}, err
	}

	resp := mgmtPB.ListConnectorExecuteChartRecordsResponse{
		ConnectorExecuteChartRecords: connectorExecuteChartRecords,
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		pbUser,
		eventName,
	)))

	return &resp, nil
}
