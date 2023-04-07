package handler

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/iancoleman/strcase"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	fieldmask_utils "github.com/mennanov/fieldmask-utils"

	"github.com/instill-ai/mgmt-backend/config"
	"github.com/instill-ai/mgmt-backend/pkg/constant"
	"github.com/instill-ai/mgmt-backend/pkg/datamodel"
	"github.com/instill-ai/mgmt-backend/pkg/logger"
	"github.com/instill-ai/mgmt-backend/pkg/middleware"
	"github.com/instill-ai/mgmt-backend/pkg/service"
	"github.com/instill-ai/mgmt-backend/pkg/usage"
	"github.com/instill-ai/x/sterr"

	healthcheckPB "github.com/instill-ai/protogen-go/vdp/healthcheck/v1alpha"
	mgmtPB "github.com/instill-ai/protogen-go/vdp/mgmt/v1alpha"
	checkfield "github.com/instill-ai/x/checkfield"
)

// TODO: Validate mask based on the field behavior. Currently, the fields are hard-coded.
// We stipulate that the ID of the user is IMMUTABLE
var createRequiredFields = []string{"id", "email", "newsletter_subscription"}
var outputOnlyFields = []string{"name", "type", "create_time", "update_time", "customer_id"}
var immutableFields = []string{"uid", "id"}

type PublicHandler struct {
	mgmtPB.UnimplementedMgmtPublicServiceServer
	service service.Service
	usg     usage.Usage
}

// NewPublicHandler initiates a public handler instance
func NewPublicHandler(s service.Service, u usage.Usage) mgmtPB.MgmtPublicServiceServer {
	return &PublicHandler{
		service: s,
		usg:     u,
	}
}

// GetService returns the service
func (h *PublicHandler) GetService() service.Service {
	return h.service
}

// SetService sets the service
func (h *PublicHandler) SetService(s service.Service) {
	h.service = s
}

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

// GetUser returns the authenticated user
func (h *PublicHandler) GetUser(ctx context.Context) (*mgmtPB.User, error) {
	logger, _ := logger.GetZapLogger()

	var dbUser *datamodel.User
	var err error

	// Verify if "jwt-sub" is in the header
	headerUserUId := middleware.GetRequestSingleHeader(ctx, constant.HeaderUserUIDKey)
	if headerUserUId != "" {
		uid, err := uuid.FromString(headerUserUId)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "Unauthenticated request")
		}
		dbUser, err = h.service.GetUser(uid)
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
		if headerUserId != constant.DefaultUserID {
			return nil, status.Error(codes.Unauthenticated, "Unauthenticated request")
		} else {
			dbUser, err = h.service.GetUserByID(headerUserId)
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
	return pbUser, nil
}

// QueryAuthenticatedUser gets the authenticated user.
// Note: this endpoint assumes the ID of the authenticated user is the default user.
func (h *PublicHandler) QueryAuthenticatedUser(ctx context.Context, req *mgmtPB.QueryAuthenticatedUserRequest) (*mgmtPB.QueryAuthenticatedUserResponse, error) {
	pbUser, err := h.GetUser(ctx)
	if err != nil {
		return &mgmtPB.QueryAuthenticatedUserResponse{}, err
	}
	resp := mgmtPB.QueryAuthenticatedUserResponse{
		User: pbUser,
	}
	return &resp, nil
}

// PatchAuthenticatedUser updates the authenticated user.
// Note: this endpoint assumes the ID of the authenticated user is the default user.
func (h *PublicHandler) PatchAuthenticatedUser(ctx context.Context, req *mgmtPB.PatchAuthenticatedUserRequest) (*mgmtPB.PatchAuthenticatedUserResponse, error) {
	logger, _ := logger.GetZapLogger()

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

	dbUserUpdated, err := h.service.UpdateUser(uid, dbUserToUpd)
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

	// Trigger single reporter right after user updated
	if !config.Config.Server.DisableUsage && h.usg != nil {
		h.usg.TriggerSingleReporter(context.Background())
	}

	return &resp, nil
}

// ExistUsername verifies if a username (ID) has been occupied
func (h *PublicHandler) ExistUsername(ctx context.Context, req *mgmtPB.ExistUsernameRequest) (*mgmtPB.ExistUsernameResponse, error) {
	logger, _ := logger.GetZapLogger()

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

	_, err = h.service.GetUserByID(id)
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

	resp := mgmtPB.ExistUsernameResponse{
		Exists: true,
	}
	return &resp, nil
}

// CreateToken creates an API token for triggering pipelines. This endpoint is not supported yet.
func (h *PublicHandler) CreateToken(ctx context.Context, req *mgmtPB.CreateTokenRequest) (*mgmtPB.CreateTokenResponse, error) {
	logger, _ := logger.GetZapLogger()

	st, err := sterr.CreateErrorResourceInfo(
		codes.Unimplemented,
		"create token not implemented error",
		"endpoint",
		"/tokens",
		"",
		"not implemented",
	)
	if err != nil {
		logger.Error(err.Error())
	}
	return &mgmtPB.CreateTokenResponse{}, st.Err()
}

// ListTokens lists all the API tokens of the authenticated user. This endpoint is not supported yet.
func (h *PublicHandler) ListTokens(ctx context.Context, req *mgmtPB.ListTokensRequest) (*mgmtPB.ListTokensResponse, error) {
	logger, _ := logger.GetZapLogger()

	st, err := sterr.CreateErrorResourceInfo(
		codes.Unimplemented,
		"list tokens not implemented error",
		"endpoint",
		"/tokens",
		"",
		"not implemented",
	)
	if err != nil {
		logger.Error(err.Error())
	}
	return &mgmtPB.ListTokensResponse{}, st.Err()
}

// GetToken gets an API token of the authenticated user. This endpoint is not supported yet.
func (h *PublicHandler) GetToken(ctx context.Context, req *mgmtPB.GetTokenRequest) (*mgmtPB.GetTokenResponse, error) {
	logger, _ := logger.GetZapLogger()

	st, err := sterr.CreateErrorResourceInfo(
		codes.Unimplemented,
		"get token not implemented error",
		"endpoint",
		"/tokens/{token}",
		"",
		"not implemented",
	)
	if err != nil {
		logger.Error(err.Error())
	}
	return &mgmtPB.GetTokenResponse{}, st.Err()
}

// DeleteToken deletes an API token of the authenticated user. This endpoint is not supported yet.
func (h *PublicHandler) DeleteToken(ctx context.Context, req *mgmtPB.DeleteTokenRequest) (*mgmtPB.DeleteTokenResponse, error) {
	logger, _ := logger.GetZapLogger()

	st, err := sterr.CreateErrorResourceInfo(
		codes.Unimplemented,
		"delete token not implemented error",
		"endpoint",
		"/tokens/{token}",
		"",
		"not implemented",
	)
	if err != nil {
		logger.Error(err.Error())
	}
	return &mgmtPB.DeleteTokenResponse{}, st.Err()
}