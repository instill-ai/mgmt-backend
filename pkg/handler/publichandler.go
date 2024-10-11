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
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	fieldmask_utils "github.com/mennanov/fieldmask-utils"

	"github.com/instill-ai/mgmt-backend/internal/resource"
	"github.com/instill-ai/mgmt-backend/pkg/constant"
	"github.com/instill-ai/mgmt-backend/pkg/logger"
	"github.com/instill-ai/mgmt-backend/pkg/service"
	"github.com/instill-ai/mgmt-backend/pkg/usage"

	custom_otel "github.com/instill-ai/mgmt-backend/pkg/logger/otel"
	healthcheckPB "github.com/instill-ai/protogen-go/common/healthcheck/v1beta"
	mgmtPB "github.com/instill-ai/protogen-go/core/mgmt/v1beta"
	checkfield "github.com/instill-ai/x/checkfield"
)

// TODO: Validate mask based on the field behavior. Currently, the fields are hard-coded.
// We stipulate that the ID of the user is IMMUTABLE
var outputOnlyFields = []string{"name", "type", "create_time", "update_time", "customer_id"}
var immutableFields = []string{"uid", "id"}

var createRequiredFieldsForToken = []string{"id"}
var outputOnlyFieldsForToken = []string{"name", "uid", "state", "token_type", "access_token", "create_time", "update_time"}

var createRequiredFieldsForOrganization = []string{"id"}
var outputOnlyFieldsForOrganization = []string{"name", "uid", "create_time", "update_time"}

var requiredFieldsForOrganizationMembership = []string{"role"}
var outputOnlyFieldsForOrganizationMembership = []string{"name", "state", "user", "organization"}

var requiredFieldsForUserMembership = []string{"state"}
var outputOnlyFieldsForUserMembership = []string{"name", "role", "user", "organization"}

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

	user, err := h.Service.GetUserAdmin(ctx, in.Username)
	if err != nil {
		return nil, err
	}

	err = h.Service.CheckUserPassword(ctx, uuid.FromStringOrNil(*user.Uid), in.Password)
	if err != nil {
		return nil, err
	}

	jti, _ := uuid.NewV4()
	exp := int32(time.Now().Unix()) + constant.DefaultJwtExpiration
	return &mgmtPB.AuthTokenIssuerResponse{
		AccessToken: &mgmtPB.AuthTokenIssuerResponse_UnsignedAccessToken{
			Aud: constant.DefaultJwtAudience,
			Sub: *user.Uid,
			Iss: constant.DefaultJwtIssuer,
			Jti: jti.String(),
			Exp: exp,
		},
	}, nil
}

func (h *PublicHandler) AuthChangePassword(ctx context.Context, in *mgmtPB.AuthChangePasswordRequest) (*mgmtPB.AuthChangePasswordResponse, error) {

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		return nil, err
	}
	user, err := h.Service.GetUserByUIDAdmin(ctx, ctxUserUID)
	if err != nil {
		return nil, err
	}

	err = h.Service.CheckUserPassword(ctx, uuid.FromStringOrNil(*user.Uid), in.OldPassword)
	if err != nil {
		return nil, err
	}

	err = h.Service.UpdateUserPassword(ctx, uuid.FromStringOrNil(*user.Uid), in.NewPassword)
	if err != nil {
		return nil, err
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

func (h *PublicHandler) ListUsers(ctx context.Context, req *mgmtPB.ListUsersRequest) (*mgmtPB.ListUsersResponse, error) {

	eventName := "ListUsers"

	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, true)
	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, err
	}

	pbUsers, totalSize, nextPageToken, err := h.Service.ListUsers(ctx, ctxUserUID, int(req.GetPageSize()), req.GetPageToken(), filtering.Filter{})
	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, err
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		ctxUserUID,
		eventName,
	)))

	resp := mgmtPB.ListUsersResponse{
		Users:         pbUsers,
		NextPageToken: nextPageToken,
		TotalSize:     int32(totalSize),
	}

	return &resp, nil
}

// GetUser gets the user.
func (h *PublicHandler) GetUser(ctx context.Context, req *mgmtPB.GetUserRequest) (*mgmtPB.GetUserResponse, error) {

	eventName := "GetUser"
	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, true)
	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, err
	}

	pbUser, err := h.Service.GetUser(ctx, ctxUserUID, req.UserId)

	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		ctxUserUID,
		eventName,
	)))

	resp := mgmtPB.GetUserResponse{
		User: pbUser,
	}
	return &resp, nil
}

// GetAuthenticatedUser gets the authenticated user.
func (h *PublicHandler) GetAuthenticatedUser(ctx context.Context, req *mgmtPB.GetAuthenticatedUserRequest) (*mgmtPB.GetAuthenticatedUserResponse, error) {

	eventName := "GetAuthenticatedUser"
	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, err
	}

	pbUser, err := h.Service.GetAuthenticatedUser(ctx, ctxUserUID)

	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		ctxUserUID,
		eventName,
	)))

	resp := mgmtPB.GetAuthenticatedUserResponse{
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

	pbUpdateMask := req.GetUpdateMask()

	// Validate the field mask
	if !pbUpdateMask.IsValid(reqUser) {
		return nil, fmt.Errorf("err")
	}

	reqFieldMask, err := checkfield.CheckUpdateOutputOnlyFields(req.GetUpdateMask(), outputOnlyFields)
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	mask, err := fieldmask_utils.MaskFromProtoFieldMask(reqFieldMask, strcase.ToCamel)
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, err
	}

	pbUserToUpdate, err := h.Service.GetAuthenticatedUser(ctx, ctxUserUID)
	if err != nil {
		return nil, err
	}

	if mask.IsEmpty() {
		// return the un-changed user `pbUserToUpdate`
		resp := mgmtPB.PatchAuthenticatedUserResponse{
			User: pbUserToUpdate,
		}
		return &resp, nil
	}

	// Handle immutable fields from the update mask
	err = checkfield.CheckUpdateImmutableFields(reqUser, pbUserToUpdate, immutableFields)
	if err != nil {
		return nil, ErrCheckUpdateImmutableFields
	}

	// Only the fields mentioned in the field mask will be copied to `pbUserToUpdate`, other fields are left intact
	err = fieldmask_utils.StructToStruct(mask, reqUser, pbUserToUpdate)
	if err != nil {
		return nil, ErrFieldMask
	}

	pbUserUpdated, err := h.Service.UpdateAuthenticatedUser(ctx, ctxUserUID, pbUserToUpdate)
	if err != nil {
		return nil, err
	}

	resp := mgmtPB.PatchAuthenticatedUserResponse{
		User: pbUserUpdated,
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		ctxUserUID,
		eventName,
		custom_otel.SetEventResource(pbUserUpdated),
	)))

	// Trigger single reporter right after user updated
	if h.usageEnabled && h.Usg != nil {
		h.Usg.TriggerSingleReporter(context.Background())
	}

	return &resp, nil
}

func (h *PublicHandler) CheckNamespace(ctx context.Context, req *mgmtPB.CheckNamespaceRequest) (*mgmtPB.CheckNamespaceResponse, error) {

	eventName := "CheckNamespace"
	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	err := checkfield.CheckResourceID(req.GetId())
	if err != nil {
		return nil, ErrResourceID
	}

	_, err = h.Service.GetUserAdmin(ctx, req.GetId())
	if err == nil {
		return &mgmtPB.CheckNamespaceResponse{
			Type: mgmtPB.CheckNamespaceResponse_NAMESPACE_USER,
		}, nil
	}
	_, err = h.Service.GetOrganizationAdmin(ctx, req.GetId())
	if err == nil {
		return &mgmtPB.CheckNamespaceResponse{
			Type: mgmtPB.CheckNamespaceResponse_NAMESPACE_ORGANIZATION,
		}, nil
	}

	return &mgmtPB.CheckNamespaceResponse{
		Type: mgmtPB.CheckNamespaceResponse_NAMESPACE_AVAILABLE,
	}, nil
}

func (h *PublicHandler) CreateOrganization(ctx context.Context, req *mgmtPB.CreateOrganizationRequest) (*mgmtPB.CreateOrganizationResponse, error) {

	eventName := "CreateOrganization"
	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	// Set all OUTPUT_ONLY fields to zero value on the requested payload organization resource
	if err := checkfield.CheckCreateOutputOnlyFields(req.Organization, outputOnlyFieldsForOrganization); err != nil {
		return nil, ErrCheckOutputOnlyFields
	}

	// Return error if REQUIRED fields are not provided in the requested payload organization resource
	if err := checkfield.CheckRequiredFields(req.Organization, createRequiredFieldsForOrganization); err != nil {
		return nil, ErrCheckRequiredFields
	}

	// Return error if resource ID does not follow RFC-1034
	if err := checkfield.CheckResourceID(req.Organization.GetId()); err != nil {
		return nil, ErrResourceID
	}

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, err
	}

	pbCreatedOrg, createErr := h.Service.CreateOrganization(ctx, ctxUserUID, req.Organization)
	if createErr != nil {
		return nil, createErr
	}

	resp := &mgmtPB.CreateOrganizationResponse{
		Organization: pbCreatedOrg,
	}

	// Manually set the custom header to have a StatusCreated http response for REST endpoint
	if err := grpc.SetHeader(ctx, metadata.Pairs("x-http-code", strconv.Itoa(http.StatusCreated))); err != nil {
		return nil, err
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		ctxUserUID,
		eventName,
		custom_otel.SetEventResult(fmt.Sprintf("Total records retrieved: %v", pbCreatedOrg)),
	)))

	return resp, nil
}

func (h *PublicHandler) ListOrganizations(ctx context.Context, req *mgmtPB.ListOrganizationsRequest) (*mgmtPB.ListOrganizationsResponse, error) {

	eventName := "ListOrganizations"
	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, true)
	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, err
	}

	pbOrgs, totalSize, nextPageToken, err := h.Service.ListOrganizations(ctx, ctxUserUID, int(req.GetPageSize()), req.GetPageToken(), filtering.Filter{})
	if err != nil {
		return nil, err
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		ctxUserUID,
		eventName,
	)))

	resp := &mgmtPB.ListOrganizationsResponse{
		Organizations: pbOrgs,
		NextPageToken: nextPageToken,
		TotalSize:     int32(totalSize),
	}
	return resp, nil
}

func (h *PublicHandler) GetOrganization(ctx context.Context, req *mgmtPB.GetOrganizationRequest) (*mgmtPB.GetOrganizationResponse, error) {

	eventName := "GetOrganization"
	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, true)
	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, err
	}

	pbOrg, err := h.Service.GetOrganization(ctx, ctxUserUID, req.OrganizationId)
	if err != nil {
		return nil, err
	}

	resp := &mgmtPB.GetOrganizationResponse{
		Organization: pbOrg,
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		ctxUserUID,
		eventName,
	)))

	return resp, nil
}

// return pbPipeline, nil
func (h *PublicHandler) UpdateOrganization(ctx context.Context, req *mgmtPB.UpdateOrganizationRequest) (*mgmtPB.UpdateOrganizationResponse, error) {

	eventName := "UpdateOrganization"
	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, err
	}

	pbOrgReq := req.GetOrganization()
	pbUpdateMask := req.GetUpdateMask()

	// Validate the field mask
	if !pbUpdateMask.IsValid(pbOrgReq) {
		return nil, ErrUpdateMask
	}

	getResp, err := h.GetOrganization(ctx, &mgmtPB.GetOrganizationRequest{OrganizationId: req.OrganizationId})
	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, err
	}

	mask, err := fieldmask_utils.MaskFromProtoFieldMask(pbUpdateMask, strcase.ToCamel)
	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, ErrFieldMask
	}

	if mask.IsEmpty() {
		return &mgmtPB.UpdateOrganizationResponse{
			Organization: getResp.GetOrganization(),
		}, nil
	}

	pbOrgToUpdate := getResp.GetOrganization()

	// Return error if IMMUTABLE fields are intentionally changed
	if err := checkfield.CheckUpdateImmutableFields(pbOrgReq, pbOrgToUpdate, immutableFields); err != nil {
		span.SetStatus(1, err.Error())
		return nil, ErrCheckUpdateImmutableFields
	}

	// Only the fields mentioned in the field mask will be copied to `pbPipelineToUpdate`, other fields are left intact
	err = fieldmask_utils.StructToStruct(mask, pbOrgReq, pbOrgToUpdate)
	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, ErrFieldMask
	}

	pbOrg, err := h.Service.UpdateOrganization(ctx, ctxUserUID, req.OrganizationId, pbOrgToUpdate)

	if err != nil {
		return nil, err
	}

	resp := &mgmtPB.UpdateOrganizationResponse{
		Organization: pbOrg,
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		ctxUserUID,
		eventName,
		custom_otel.SetEventResource(pbOrg),
	)))

	return resp, nil
}
func (h *PublicHandler) DeleteOrganization(ctx context.Context, req *mgmtPB.DeleteOrganizationRequest) (*mgmtPB.DeleteOrganizationResponse, error) {

	eventName := "DeleteOrganization"

	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, err
	}
	existOrg, err := h.GetOrganization(ctx, &mgmtPB.GetOrganizationRequest{OrganizationId: req.OrganizationId})
	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, err
	}

	if err := h.Service.DeleteOrganization(ctx, ctxUserUID, req.OrganizationId); err != nil {
		span.SetStatus(1, err.Error())
		return nil, err
	}

	// We need to manually set the custom header to have a StatusCreated http response for REST endpoint
	if err := grpc.SetHeader(ctx, metadata.Pairs("x-http-code", strconv.Itoa(http.StatusNoContent))); err != nil {
		span.SetStatus(1, err.Error())
		return nil, err
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		ctxUserUID,
		eventName,
		custom_otel.SetEventResource(existOrg.GetOrganization()),
	)))

	return &mgmtPB.DeleteOrganizationResponse{}, nil
}

// CreateToken creates an API token for triggering pipelines. This endpoint is not supported yet.
func (h *PublicHandler) CreateToken(ctx context.Context, req *mgmtPB.CreateTokenRequest) (*mgmtPB.CreateTokenResponse, error) {

	eventName := "CreateToken"
	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, err
	}

	// Set all OUTPUT_ONLY fields to zero value on the requested payload token resource
	if err := checkfield.CheckCreateOutputOnlyFields(req.Token, outputOnlyFieldsForToken); err != nil {
		return &mgmtPB.CreateTokenResponse{}, ErrCheckOutputOnlyFields
	}

	// Return error if REQUIRED fields are not provided in the requested payload token resource
	if err := checkfield.CheckRequiredFields(req.Token, createRequiredFieldsForToken); err != nil {
		return &mgmtPB.CreateTokenResponse{}, ErrCheckRequiredFields
	}

	// Return error if resource ID does not follow RFC-1034
	if err := checkfield.CheckResourceID(req.Token.GetId()); err != nil {
		return &mgmtPB.CreateTokenResponse{}, ErrResourceID
	}

	// Return error if expiration is not provided
	if req.Token.GetExpiration() == nil {
		return &mgmtPB.CreateTokenResponse{}, ErrCheckRequiredFields
	}

	err = h.Service.CreateToken(ctx, ctxUserUID, req.Token)
	if err != nil {
		return nil, err
	}

	pbCreatedToken, err := h.Service.GetToken(ctx, ctxUserUID, req.Token.Id)
	if err != nil {
		return nil, err
	}

	resp := &mgmtPB.CreateTokenResponse{
		Token: pbCreatedToken,
	}

	// Manually set the custom header to have a StatusCreated http response for REST endpoint
	if err := grpc.SetHeader(ctx, metadata.Pairs("x-http-code", strconv.Itoa(http.StatusCreated))); err != nil {
		return nil, err
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		ctxUserUID,
		eventName,
		custom_otel.SetEventResult(fmt.Sprintf("Total records retrieved: %v", pbCreatedToken)),
	)))

	return resp, nil
}

// ListTokens lists all the API tokens of the authenticated user.
func (h *PublicHandler) ListTokens(ctx context.Context, req *mgmtPB.ListTokensRequest) (*mgmtPB.ListTokensResponse, error) {

	eventName := "ListTokens"
	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, err
	}

	pbTokens, totalSize, nextPageToken, err := h.Service.ListTokens(ctx, ctxUserUID, int64(req.GetPageSize()), req.GetPageToken())
	if err != nil {
		return &mgmtPB.ListTokensResponse{}, err
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		ctxUserUID,
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

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, err
	}

	pbToken, err := h.Service.GetToken(ctx, ctxUserUID, req.TokenId)
	if err != nil {
		return nil, err
	}

	resp := &mgmtPB.GetTokenResponse{
		Token: pbToken,
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		ctxUserUID,
		eventName,
		custom_otel.SetEventResource(pbToken),
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

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, err
	}

	existToken, err := h.GetToken(ctx, &mgmtPB.GetTokenRequest{TokenId: req.TokenId})
	if err != nil {
		return nil, err
	}

	if err := h.Service.DeleteToken(ctx, ctxUserUID, existToken.Token.GetId()); err != nil {
		return nil, err
	}

	// We need to manually set the custom header to have a StatusCreated http response for REST endpoint
	if err := grpc.SetHeader(ctx, metadata.Pairs("x-http-code", strconv.Itoa(http.StatusNoContent))); err != nil {
		return &mgmtPB.DeleteTokenResponse{}, err
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		ctxUserUID,
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

	userUID, err := h.Service.ValidateToken(ctx, apiToken)

	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, err
	}

	err = h.Service.UpdateTokenLastUseTime(ctx, apiToken)

	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, err
	}

	return &mgmtPB.ValidateTokenResponse{UserUid: userUID}, nil
}

// GetPipelineTriggerCount returns the pipeline trigger count of a given
// requester within a timespan.  Results are grouped by trigger status.
func (h *PublicHandler) GetPipelineTriggerCount(ctx context.Context, req *mgmtPB.GetPipelineTriggerCountRequest) (*mgmtPB.GetPipelineTriggerCountResponse, error) {
	eventName := "GetPipelineTriggerCount"
	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()
	logger, _ := logger.GetZapLogger(ctx)

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, err
	}

	resp, err := h.Service.GetPipelineTriggerCount(ctx, req, ctxUserUID)
	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, fmt.Errorf("fetching credit chart records: %w", err)
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		ctxUserUID,
		eventName,
	)))

	return resp, nil
}

// ListPipelineTriggerChartRecords returns a timeline of a requester's pipeline
// trigger count.
func (h *PublicHandler) ListPipelineTriggerChartRecords(ctx context.Context, req *mgmtPB.ListPipelineTriggerChartRecordsRequest) (*mgmtPB.ListPipelineTriggerChartRecordsResponse, error) {
	eventName := "ListPipelineTriggerChartRecords"
	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()
	logger, _ := logger.GetZapLogger(ctx)

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, err
	}

	resp, err := h.Service.ListPipelineTriggerChartRecords(ctx, req, ctxUserUID)
	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, fmt.Errorf("fetching credit chart records: %w", err)
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		ctxUserUID,
		eventName,
	)))

	return resp, nil
}

func (h *PublicHandler) ListUserMemberships(ctx context.Context, req *mgmtPB.ListUserMembershipsRequest) (*mgmtPB.ListUserMembershipsResponse, error) {

	eventName := "ListUserMemberships"

	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, err
	}

	pbMemberships, err := h.Service.ListUserMemberships(ctx, ctxUserUID, req.UserId)
	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, err
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		ctxUserUID,
		eventName,
	)))

	resp := mgmtPB.ListUserMembershipsResponse{
		Memberships: pbMemberships,
	}

	return &resp, nil
}

func (h *PublicHandler) GetUserMembership(ctx context.Context, req *mgmtPB.GetUserMembershipRequest) (*mgmtPB.GetUserMembershipResponse, error) {

	eventName := "GetUserMembership"

	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, err
	}

	pbMembership, err := h.Service.GetUserMembership(ctx, ctxUserUID, req.UserId, req.OrganizationId)
	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, err
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		ctxUserUID,
		eventName,
	)))

	resp := mgmtPB.GetUserMembershipResponse{
		Membership: pbMembership,
	}

	return &resp, nil
}

func (h *PublicHandler) UpdateUserMembership(ctx context.Context, req *mgmtPB.UpdateUserMembershipRequest) (*mgmtPB.UpdateUserMembershipResponse, error) {

	eventName := "UpdateUserMembership"

	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, err
	}

	if err := checkfield.CheckRequiredFields(req.Membership, requiredFieldsForUserMembership); err != nil {
		return nil, ErrCheckRequiredFields
	}

	if err := checkfield.CheckCreateOutputOnlyFields(req.Membership, outputOnlyFieldsForUserMembership); err != nil {
		return nil, ErrCheckOutputOnlyFields
	}

	pbMembership, err := h.Service.UpdateUserMembership(ctx, ctxUserUID, req.UserId, req.OrganizationId, req.Membership)
	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, err
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		ctxUserUID,
		eventName,
	)))

	resp := mgmtPB.UpdateUserMembershipResponse{
		Membership: pbMembership,
	}

	return &resp, nil
}

func (h *PublicHandler) DeleteUserMembership(ctx context.Context, req *mgmtPB.DeleteUserMembershipRequest) (*mgmtPB.DeleteUserMembershipResponse, error) {

	eventName := "DeleteUserMembership"

	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, err
	}

	err = h.Service.DeleteUserMembership(ctx, ctxUserUID, req.UserId, req.OrganizationId)
	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, err
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		ctxUserUID,
		eventName,
	)))

	resp := mgmtPB.DeleteUserMembershipResponse{}

	return &resp, nil
}

func (h *PublicHandler) ListOrganizationMemberships(ctx context.Context, req *mgmtPB.ListOrganizationMembershipsRequest) (*mgmtPB.ListOrganizationMembershipsResponse, error) {

	eventName := "ListOrganizationMemberships"

	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, err
	}

	pbMemberships, err := h.Service.ListOrganizationMemberships(ctx, ctxUserUID, req.OrganizationId)
	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, err
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		ctxUserUID,
		eventName,
	)))

	resp := mgmtPB.ListOrganizationMembershipsResponse{
		Memberships: pbMemberships,
	}

	return &resp, nil
}

func (h *PublicHandler) GetOrganizationMembership(ctx context.Context, req *mgmtPB.GetOrganizationMembershipRequest) (*mgmtPB.GetOrganizationMembershipResponse, error) {

	eventName := "GetOrganizationMembership"

	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, err
	}

	pbMembership, err := h.Service.GetOrganizationMembership(ctx, ctxUserUID, req.OrganizationId, req.UserId)
	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, err
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		ctxUserUID,
		eventName,
	)))

	resp := mgmtPB.GetOrganizationMembershipResponse{
		Membership: pbMembership,
	}

	return &resp, nil
}

func (h *PublicHandler) UpdateOrganizationMembership(ctx context.Context, req *mgmtPB.UpdateOrganizationMembershipRequest) (*mgmtPB.UpdateOrganizationMembershipResponse, error) {

	eventName := "UpdateOrganizationMembership"

	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, err
	}

	if err := checkfield.CheckRequiredFields(req.Membership, requiredFieldsForOrganizationMembership); err != nil {
		return nil, ErrCheckRequiredFields
	}

	if err := checkfield.CheckCreateOutputOnlyFields(req.Membership, outputOnlyFieldsForOrganizationMembership); err != nil {
		return nil, ErrCheckOutputOnlyFields
	}

	pbMembership, err := h.Service.UpdateOrganizationMembership(ctx, ctxUserUID, req.OrganizationId, req.UserId, req.Membership)
	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, err
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		ctxUserUID,
		eventName,
	)))

	resp := mgmtPB.UpdateOrganizationMembershipResponse{
		Membership: pbMembership,
	}

	return &resp, nil
}

func (h *PublicHandler) DeleteOrganizationMembership(ctx context.Context, req *mgmtPB.DeleteOrganizationMembershipRequest) (*mgmtPB.DeleteOrganizationMembershipResponse, error) {

	eventName := "DeleteOrganizationMembership"

	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, err
	}

	err = h.Service.DeleteOrganizationMembership(ctx, ctxUserUID, req.OrganizationId, req.UserId)
	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, err
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		ctxUserUID,
		eventName,
	)))

	return &mgmtPB.DeleteOrganizationMembershipResponse{}, nil
}
