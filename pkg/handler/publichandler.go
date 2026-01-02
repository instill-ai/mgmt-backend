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
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	fieldmask_utils "github.com/mennanov/fieldmask-utils"

	"github.com/instill-ai/mgmt-backend/internal/resource"
	"github.com/instill-ai/mgmt-backend/pkg/constant"
	"github.com/instill-ai/mgmt-backend/pkg/service"
	"github.com/instill-ai/mgmt-backend/pkg/usage"
	"github.com/instill-ai/x/checkfield"

	healthcheckpb "github.com/instill-ai/protogen-go/common/healthcheck/v1beta"
	mgmtpb "github.com/instill-ai/protogen-go/core/mgmt/v1beta"
	errorsx "github.com/instill-ai/x/errors"
)

// TODO: Validate mask based on the field behavior. Currently, the fields are hard-coded.
// We stipulate that the ID of the user is IMMUTABLE
var outputOnlyFields = []string{"name", "create_time", "update_time", "customer_id"}
var immutableFields = []string{"uid", "id"}

var createRequiredFieldsForToken = []string{"id"}
var outputOnlyFieldsForToken = []string{"name", "uid", "state", "token_type", "access_token", "create_time", "update_time"}

var requiredFieldsForOrganizationMembership = []string{"role"}
var outputOnlyFieldsForOrganizationMembership = []string{"name", "state", "user", "organization"}

var requiredFieldsForUserMembership = []string{"state"}
var outputOnlyFieldsForUserMembership = []string{"name", "role", "user", "organization"}

// PublicHandler is the handler for the public endpoints.
type PublicHandler struct {
	mgmtpb.UnimplementedMgmtPublicServiceServer
	Service      service.Service
	Usg          usage.Usage
	usageEnabled bool
}

// NewPublicHandler initiates a public handler instance
func NewPublicHandler(s service.Service, u usage.Usage, usageEnabled bool) mgmtpb.MgmtPublicServiceServer {
	return &PublicHandler{
		Service:      s,
		Usg:          u,
		usageEnabled: usageEnabled,
	}
}

// Liveness checks the liveness of the server
func (h *PublicHandler) Liveness(ctx context.Context, in *mgmtpb.LivenessRequest) (*mgmtpb.LivenessResponse, error) {
	return &mgmtpb.LivenessResponse{
		HealthCheckResponse: &healthcheckpb.HealthCheckResponse{
			Status: healthcheckpb.HealthCheckResponse_SERVING_STATUS_SERVING,
		},
	}, nil
}

// Readiness checks the readiness of the server
func (h *PublicHandler) Readiness(ctx context.Context, in *mgmtpb.ReadinessRequest) (*mgmtpb.ReadinessResponse, error) {
	return &mgmtpb.ReadinessResponse{
		HealthCheckResponse: &healthcheckpb.HealthCheckResponse{
			Status: healthcheckpb.HealthCheckResponse_SERVING_STATUS_SERVING,
		},
	}, nil
}

// AuthTokenIssuer issues a token for the user.
func (h *PublicHandler) AuthTokenIssuer(ctx context.Context, in *mgmtpb.AuthTokenIssuerRequest) (*mgmtpb.AuthTokenIssuerResponse, error) {

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
	return &mgmtpb.AuthTokenIssuerResponse{
		AccessToken: &mgmtpb.AuthTokenIssuerResponse_UnsignedAccessToken{
			Aud: constant.DefaultJwtAudience,
			Sub: *user.Uid,
			Iss: constant.DefaultJwtIssuer,
			Jti: jti.String(),
			Exp: exp,
		},
	}, nil
}

// AuthChangePassword changes the password of the user.
func (h *PublicHandler) AuthChangePassword(ctx context.Context, in *mgmtpb.AuthChangePasswordRequest) (*mgmtpb.AuthChangePasswordResponse, error) {

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

	return &mgmtpb.AuthChangePasswordResponse{}, nil
}

// AuthLogout logs out the user.
func (h *PublicHandler) AuthLogout(ctx context.Context, in *mgmtpb.AuthLogoutRequest) (*mgmtpb.AuthLogoutResponse, error) {
	// TODO: implement this
	return &mgmtpb.AuthLogoutResponse{}, nil
}

// AuthLogin logs in the user.
func (h *PublicHandler) AuthLogin(ctx context.Context, in *mgmtpb.AuthLoginRequest) (*mgmtpb.AuthLoginResponse, error) {
	// This endpoint will be handled by KrakenD. We don't need to implement here
	return &mgmtpb.AuthLoginResponse{}, nil
}

// AuthValidateAccessToken validates the access token.
func (h *PublicHandler) AuthValidateAccessToken(ctx context.Context, in *mgmtpb.AuthValidateAccessTokenRequest) (*mgmtpb.AuthValidateAccessTokenResponse, error) {
	// This endpoint will be handled by KrakenD. We don't need to implement here
	return &mgmtpb.AuthValidateAccessTokenResponse{}, nil
}

// ListUsers lists the users.
func (h *PublicHandler) ListUsers(ctx context.Context, req *mgmtpb.ListUsersRequest) (*mgmtpb.ListUsersResponse, error) {

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, true)
	if err != nil {
		return nil, err
	}

	declarations, err := filtering.NewDeclarations([]filtering.DeclarationOption{
		filtering.DeclareStandardFunctions(),
		filtering.DeclareIdent(constant.Email, filtering.TypeString),
	}...)
	if err != nil {
		return nil, err
	}
	filter, err := filtering.ParseFilter(req, declarations)
	if err != nil {
		return nil, err
	}

	pbUsers, totalSize, nextPageToken, err := h.Service.ListUsers(ctx, ctxUserUID, int(req.GetPageSize()), req.GetPageToken(), filter)
	if err != nil {
		return nil, err
	}

	resp := mgmtpb.ListUsersResponse{
		Users:         pbUsers,
		NextPageToken: nextPageToken,
		TotalSize:     int32(totalSize),
	}

	return &resp, nil
}

// GetUser gets the user.
func (h *PublicHandler) GetUser(ctx context.Context, req *mgmtpb.GetUserRequest) (*mgmtpb.GetUserResponse, error) {

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, true)
	if err != nil {
		return nil, err
	}

	pbUser, err := h.Service.GetUser(ctx, ctxUserUID, req.UserId)

	if err != nil {
		return nil, err
	}

	resp := mgmtpb.GetUserResponse{
		User: pbUser,
	}
	return &resp, nil
}

// GetAuthenticatedUser gets the authenticated user.
func (h *PublicHandler) GetAuthenticatedUser(ctx context.Context, req *mgmtpb.GetAuthenticatedUserRequest) (*mgmtpb.GetAuthenticatedUserResponse, error) {

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		return nil, err
	}

	pbUser, err := h.Service.GetAuthenticatedUser(ctx, ctxUserUID)

	if err != nil {
		return nil, err
	}

	resp := mgmtpb.GetAuthenticatedUserResponse{
		User: pbUser,
	}
	return &resp, nil
}

// PatchAuthenticatedUser updates the authenticated user.
// Note: this endpoint assumes the ID of the authenticated user is the default user.
func (h *PublicHandler) PatchAuthenticatedUser(ctx context.Context, req *mgmtpb.PatchAuthenticatedUserRequest) (*mgmtpb.PatchAuthenticatedUserResponse, error) {

	reqUser := req.GetUser()

	pbUpdateMask := req.GetUpdateMask()

	// Validate the field mask
	if !pbUpdateMask.IsValid(reqUser) {
		return nil, fmt.Errorf("err")
	}

	reqFieldMask, err := checkfield.CheckUpdateOutputOnlyFields(req.GetUpdateMask(), outputOnlyFields)
	if err != nil {
		return nil, err
	}

	mask, err := fieldmask_utils.MaskFromProtoFieldMask(reqFieldMask, strcase.ToCamel)
	if err != nil {
		return nil, err
	}

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		return nil, err
	}

	pbUserToUpdate, err := h.Service.GetAuthenticatedUser(ctx, ctxUserUID)
	if err != nil {
		return nil, err
	}

	if mask.IsEmpty() {
		// return the un-changed user `pbUserToUpdate`
		resp := mgmtpb.PatchAuthenticatedUserResponse{
			User: pbUserToUpdate,
		}
		return &resp, nil
	}

	// Handle immutable fields from the update mask
	err = checkfield.CheckUpdateImmutableFields(reqUser, pbUserToUpdate, immutableFields)
	if err != nil {
		return nil, errorsx.ErrCheckUpdateImmutableFields
	}

	// Only the fields mentioned in the field mask will be copied to `pbUserToUpdate`, other fields are left intact
	err = fieldmask_utils.StructToStruct(mask, reqUser, pbUserToUpdate)
	if err != nil {
		return nil, errorsx.ErrFieldMask
	}

	pbUserUpdated, err := h.Service.UpdateAuthenticatedUser(ctx, ctxUserUID, pbUserToUpdate)
	if err != nil {
		return nil, err
	}

	resp := mgmtpb.PatchAuthenticatedUserResponse{
		User: pbUserUpdated,
	}

	// Trigger single reporter right after user updated
	if h.usageEnabled && h.Usg != nil {
		h.Usg.TriggerSingleReporter(context.Background())
	}

	return &resp, nil
}

// CheckNamespace checks if the namespace is available.
func (h *PublicHandler) CheckNamespace(ctx context.Context, req *mgmtpb.CheckNamespaceRequest) (*mgmtpb.CheckNamespaceResponse, error) {

	err := checkfield.CheckResourceID(req.GetId())
	if err != nil {
		return nil, errorsx.ErrResourceID
	}

	_, err = h.Service.GetUserAdmin(ctx, req.GetId())
	if err == nil {
		return &mgmtpb.CheckNamespaceResponse{
			Type: mgmtpb.CheckNamespaceResponse_NAMESPACE_USER,
		}, nil
	}
	_, err = h.Service.GetOrganizationAdmin(ctx, req.GetId())
	if err == nil {
		return &mgmtpb.CheckNamespaceResponse{
			Type: mgmtpb.CheckNamespaceResponse_NAMESPACE_ORGANIZATION,
		}, nil
	}

	return &mgmtpb.CheckNamespaceResponse{
		Type: mgmtpb.CheckNamespaceResponse_NAMESPACE_AVAILABLE,
	}, nil
}

// CreateOrganization creates an organization with the authenticated user as
// the owner.
func (h *PublicHandler) CreateOrganization(ctx context.Context, req *mgmtpb.CreateOrganizationRequest) (*mgmtpb.CreateOrganizationResponse, error) {
	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		return nil, err
	}

	pbCreatedOrg, createErr := h.Service.CreateOrganization(ctx, ctxUserUID, req.Organization)
	if createErr != nil {
		return nil, createErr
	}

	resp := &mgmtpb.CreateOrganizationResponse{
		Organization: pbCreatedOrg,
	}

	// Manually set the custom header to have a StatusCreated http response for REST endpoint
	if err := grpc.SetHeader(ctx, metadata.Pairs("x-http-code", strconv.Itoa(http.StatusCreated))); err != nil {
		return nil, err
	}

	return resp, nil
}

// ListOrganizations lists organizations.
func (h *PublicHandler) ListOrganizations(ctx context.Context, req *mgmtpb.ListOrganizationsRequest) (*mgmtpb.ListOrganizationsResponse, error) {

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, true)
	if err != nil {
		return nil, err
	}

	pbOrgs, totalSize, nextPageToken, err := h.Service.ListOrganizations(ctx, ctxUserUID, int(req.GetPageSize()), req.GetPageToken(), filtering.Filter{})
	if err != nil {
		return nil, err
	}

	resp := &mgmtpb.ListOrganizationsResponse{
		Organizations: pbOrgs,
		NextPageToken: nextPageToken,
		TotalSize:     int32(totalSize),
	}
	return resp, nil
}

// GetOrganization gets an organization.
func (h *PublicHandler) GetOrganization(ctx context.Context, req *mgmtpb.GetOrganizationRequest) (*mgmtpb.GetOrganizationResponse, error) {

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, true)
	if err != nil {
		return nil, err
	}

	pbOrg, err := h.Service.GetOrganization(ctx, ctxUserUID, req.OrganizationId)
	if err != nil {
		return nil, err
	}

	resp := &mgmtpb.GetOrganizationResponse{
		Organization: pbOrg,
	}

	return resp, nil
}

// UpdateOrganization updates an organization.
func (h *PublicHandler) UpdateOrganization(ctx context.Context, req *mgmtpb.UpdateOrganizationRequest) (*mgmtpb.UpdateOrganizationResponse, error) {

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		return nil, err
	}

	pbOrgReq := req.GetOrganization()
	pbUpdateMask := req.GetUpdateMask()

	// Validate the field mask
	if !pbUpdateMask.IsValid(pbOrgReq) {
		return nil, errorsx.ErrUpdateMask
	}

	getResp, err := h.GetOrganization(ctx, &mgmtpb.GetOrganizationRequest{OrganizationId: req.OrganizationId})
	if err != nil {
		return nil, err
	}

	mask, err := fieldmask_utils.MaskFromProtoFieldMask(pbUpdateMask, strcase.ToCamel)
	if err != nil {
		return nil, errorsx.ErrFieldMask
	}

	if mask.IsEmpty() {
		return &mgmtpb.UpdateOrganizationResponse{
			Organization: getResp.GetOrganization(),
		}, nil
	}

	pbOrgToUpdate := getResp.GetOrganization()

	// Return error if IMMUTABLE fields are intentionally changed
	if err := checkfield.CheckUpdateImmutableFields(pbOrgReq, pbOrgToUpdate, immutableFields); err != nil {
		return nil, errorsx.ErrCheckUpdateImmutableFields
	}

	// Only the fields mentioned in the field mask will be copied to `pbPipelineToUpdate`, other fields are left intact
	err = fieldmask_utils.StructToStruct(mask, pbOrgReq, pbOrgToUpdate)
	if err != nil {
		return nil, errorsx.ErrFieldMask
	}

	pbOrg, err := h.Service.UpdateOrganization(ctx, ctxUserUID, req.OrganizationId, pbOrgToUpdate)

	if err != nil {
		return nil, err
	}

	resp := &mgmtpb.UpdateOrganizationResponse{
		Organization: pbOrg,
	}

	return resp, nil
}

// DeleteOrganization deletes an organization.
func (h *PublicHandler) DeleteOrganization(ctx context.Context, req *mgmtpb.DeleteOrganizationRequest) (*mgmtpb.DeleteOrganizationResponse, error) {

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		return nil, err
	}

	_, err = h.GetOrganization(ctx, &mgmtpb.GetOrganizationRequest{OrganizationId: req.OrganizationId})
	if err != nil {
		return nil, err
	}

	if err := h.Service.DeleteOrganization(ctx, ctxUserUID, req.OrganizationId); err != nil {
		return nil, err
	}

	// We need to manually set the custom header to have a StatusCreated http response for REST endpoint
	if err := grpc.SetHeader(ctx, metadata.Pairs("x-http-code", strconv.Itoa(http.StatusNoContent))); err != nil {
		return nil, err
	}

	return &mgmtpb.DeleteOrganizationResponse{}, nil
}

// CreateToken creates an API token for triggering pipelines. This endpoint is not supported yet.
func (h *PublicHandler) CreateToken(ctx context.Context, req *mgmtpb.CreateTokenRequest) (*mgmtpb.CreateTokenResponse, error) {

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		return nil, err
	}

	// Set all OUTPUT_ONLY fields to zero value on the requested payload token resource
	if err := checkfield.CheckCreateOutputOnlyFields(req.Token, outputOnlyFieldsForToken); err != nil {
		return &mgmtpb.CreateTokenResponse{}, errorsx.ErrCheckOutputOnlyFields
	}

	// Return error if REQUIRED fields are not provided in the requested payload token resource
	if err := checkfield.CheckRequiredFields(req.Token, createRequiredFieldsForToken); err != nil {
		return &mgmtpb.CreateTokenResponse{}, errorsx.ErrCheckRequiredFields
	}

	// Return error if resource ID does not follow RFC-1034
	if err := checkfield.CheckResourceID(req.Token.GetId()); err != nil {
		return &mgmtpb.CreateTokenResponse{}, errorsx.ErrResourceID
	}

	// Return error if expiration is not provided
	if req.Token.GetExpiration() == nil {
		return &mgmtpb.CreateTokenResponse{}, errorsx.ErrCheckRequiredFields
	}

	err = h.Service.CreateToken(ctx, ctxUserUID, req.Token)
	if err != nil {
		return nil, err
	}

	pbCreatedToken, err := h.Service.GetToken(ctx, ctxUserUID, req.Token.Id)
	if err != nil {
		return nil, err
	}

	resp := &mgmtpb.CreateTokenResponse{
		Token: pbCreatedToken,
	}

	// Manually set the custom header to have a StatusCreated http response for REST endpoint
	if err := grpc.SetHeader(ctx, metadata.Pairs("x-http-code", strconv.Itoa(http.StatusCreated))); err != nil {
		return nil, err
	}

	return resp, nil
}

// ListTokens lists all the API tokens of the authenticated user.
func (h *PublicHandler) ListTokens(ctx context.Context, req *mgmtpb.ListTokensRequest) (*mgmtpb.ListTokensResponse, error) {

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		return nil, err
	}

	pbTokens, totalSize, nextPageToken, err := h.Service.ListTokens(ctx, ctxUserUID, int64(req.GetPageSize()), req.GetPageToken())
	if err != nil {
		return &mgmtpb.ListTokensResponse{}, err
	}

	resp := &mgmtpb.ListTokensResponse{
		Tokens:        pbTokens,
		NextPageToken: nextPageToken,
		TotalSize:     int32(totalSize),
	}
	return resp, nil
}

// GetToken gets an API token of the authenticated user. This endpoint is not supported yet.
func (h *PublicHandler) GetToken(ctx context.Context, req *mgmtpb.GetTokenRequest) (*mgmtpb.GetTokenResponse, error) {

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		return nil, err
	}

	pbToken, err := h.Service.GetToken(ctx, ctxUserUID, req.TokenId)
	if err != nil {
		return nil, err
	}

	resp := &mgmtpb.GetTokenResponse{
		Token: pbToken,
	}

	return resp, nil
}

// DeleteToken deletes an API token of the authenticated user. This endpoint is not supported yet.
func (h *PublicHandler) DeleteToken(ctx context.Context, req *mgmtpb.DeleteTokenRequest) (*mgmtpb.DeleteTokenResponse, error) {

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		return nil, err
	}

	existToken, err := h.GetToken(ctx, &mgmtpb.GetTokenRequest{TokenId: req.TokenId})
	if err != nil {
		return nil, err
	}

	if err := h.Service.DeleteToken(ctx, ctxUserUID, existToken.Token.GetId()); err != nil {
		return nil, err
	}

	// We need to manually set the custom header to have a StatusCreated http response for REST endpoint
	if err := grpc.SetHeader(ctx, metadata.Pairs("x-http-code", strconv.Itoa(http.StatusNoContent))); err != nil {
		return &mgmtpb.DeleteTokenResponse{}, err
	}

	return &mgmtpb.DeleteTokenResponse{}, nil
}

// ValidateToken validate the token
func (h *PublicHandler) ValidateToken(ctx context.Context, req *mgmtpb.ValidateTokenRequest) (*mgmtpb.ValidateTokenResponse, error) {

	authorization := resource.GetRequestSingleHeader(ctx, constant.HeaderAuthorization)
	apiToken := strings.Replace(authorization, "Bearer ", "", 1)

	userUID, err := h.Service.ValidateToken(ctx, apiToken)

	if err != nil {
		return nil, err
	}

	err = h.Service.UpdateTokenLastUseTime(ctx, apiToken)

	if err != nil {
		return nil, err
	}

	return &mgmtpb.ValidateTokenResponse{UserUid: userUID}, nil
}

// GetPipelineTriggerCount returns the pipeline trigger count of a given
// requester within a timespan.  Results are grouped by trigger status.
func (h *PublicHandler) GetPipelineTriggerCount(ctx context.Context, req *mgmtpb.GetPipelineTriggerCountRequest) (*mgmtpb.GetPipelineTriggerCountResponse, error) {

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		return nil, err
	}

	resp, err := h.Service.GetPipelineTriggerCount(ctx, req, ctxUserUID)
	if err != nil {
		return nil, fmt.Errorf("fetching pipeline trigger count: %w", err)
	}

	return resp, nil
}

// GetModelTriggerCount returns the model trigger count of a given
// requester within a timespan. Results are grouped by trigger status.
func (h *PublicHandler) GetModelTriggerCount(ctx context.Context, req *mgmtpb.GetModelTriggerCountRequest) (*mgmtpb.GetModelTriggerCountResponse, error) {

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		return nil, err
	}

	resp, err := h.Service.GetModelTriggerCount(ctx, req, ctxUserUID)
	if err != nil {
		return nil, fmt.Errorf("fetching model trigger count: %w", err)
	}

	return resp, nil
}

// ListPipelineTriggerRecords lists pipeline trigger records.
func (h *PublicHandler) ListPipelineTriggerRecords(ctx context.Context, req *mgmtpb.ListPipelineTriggerRecordsRequest) (*mgmtpb.ListPipelineTriggerRecordsResponse, error) {

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		return nil, err
	}
	pbUser, err := h.Service.GetUserByUIDAdmin(ctx, ctxUserUID)
	if err != nil {
		return nil, err
	}
	var mode mgmtpb.Mode
	var status mgmtpb.Status
	declarations, err := filtering.NewDeclarations([]filtering.DeclarationOption{
		filtering.DeclareStandardFunctions(),
		filtering.DeclareIdent(constant.Start, filtering.TypeTimestamp),
		filtering.DeclareIdent(constant.Stop, filtering.TypeTimestamp),
		filtering.DeclareIdent(strcase.ToLowerCamel(constant.OwnerName), filtering.TypeString),
		filtering.DeclareIdent(strcase.ToLowerCamel(constant.PipelineID), filtering.TypeString),
		filtering.DeclareIdent(strcase.ToLowerCamel(constant.PipelineUID), filtering.TypeString),
		filtering.DeclareIdent(strcase.ToLowerCamel(constant.PipelineReleaseID), filtering.TypeString),
		filtering.DeclareIdent(strcase.ToLowerCamel(constant.PipelineReleaseUID), filtering.TypeString),
		filtering.DeclareEnumIdent(strcase.ToLowerCamel(constant.TriggerMode), mode.Type()),
		filtering.DeclareEnumIdent(constant.Status, status.Type()),
	}...)
	if err != nil {
		return nil, err
	}
	filter, err := filtering.ParseFilter(req, declarations)
	if err != nil {
		return nil, err
	}
	pipelineTriggerRecords, totalSize, nextPageToken, err := h.Service.ListPipelineTriggerRecords(ctx, pbUser, int64(req.GetPageSize()), req.GetPageToken(), filter)
	if err != nil {
		return nil, err
	}
	resp := mgmtpb.ListPipelineTriggerRecordsResponse{
		PipelineTriggerRecords: pipelineTriggerRecords,
		NextPageToken:          nextPageToken,
		TotalSize:              int32(totalSize),
	}

	return &resp, nil
}

// ListPipelineTriggerTableRecords lists pipeline trigger table records.
func (h *PublicHandler) ListPipelineTriggerTableRecords(ctx context.Context, req *mgmtpb.ListPipelineTriggerTableRecordsRequest) (*mgmtpb.ListPipelineTriggerTableRecordsResponse, error) {

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		return nil, err
	}
	pbUser, err := h.Service.GetUserByUIDAdmin(ctx, ctxUserUID)
	if err != nil {
		return nil, err
	}

	declarations, err := filtering.NewDeclarations([]filtering.DeclarationOption{
		filtering.DeclareStandardFunctions(),
		filtering.DeclareIdent(constant.Start, filtering.TypeTimestamp),
		filtering.DeclareIdent(constant.Stop, filtering.TypeTimestamp),
		filtering.DeclareIdent(strcase.ToLowerCamel(constant.OwnerName), filtering.TypeString),
		filtering.DeclareIdent(strcase.ToLowerCamel(constant.PipelineID), filtering.TypeString),
		filtering.DeclareIdent(strcase.ToLowerCamel(constant.PipelineUID), filtering.TypeString),
		filtering.DeclareIdent(strcase.ToLowerCamel(constant.PipelineReleaseID), filtering.TypeString),
		filtering.DeclareIdent(strcase.ToLowerCamel(constant.PipelineReleaseUID), filtering.TypeString),
	}...)
	if err != nil {
		return nil, err
	}

	filter, err := filtering.ParseFilter(req, declarations)
	if err != nil {
		return nil, err
	}

	pipelineTriggerTableRecords, totalSize, nextPageToken, err := h.Service.ListPipelineTriggerTableRecords(ctx, pbUser, int64(req.GetPageSize()), req.GetPageToken(), filter)
	if err != nil {
		return nil, err
	}

	resp := mgmtpb.ListPipelineTriggerTableRecordsResponse{
		PipelineTriggerTableRecords: pipelineTriggerTableRecords,
		NextPageToken:               nextPageToken,
		TotalSize:                   int32(totalSize),
	}

	return &resp, nil
}

// ListPipelineTriggerChartRecords returns a timeline of a requester's pipeline
// trigger count.
func (h *PublicHandler) ListPipelineTriggerChartRecords(ctx context.Context, req *mgmtpb.ListPipelineTriggerChartRecordsRequest) (*mgmtpb.ListPipelineTriggerChartRecordsResponse, error) {
	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		return nil, err
	}

	resp, err := h.Service.ListPipelineTriggerChartRecords(ctx, req, ctxUserUID)
	if err != nil {
		return nil, fmt.Errorf("fetching credit chart records: %w", err)
	}

	return resp, nil
}

// ListPipelineTriggerChartRecordsV0 returns a timeline of a requester's pipeline
// trigger count.
func (h *PublicHandler) ListPipelineTriggerChartRecordsV0(ctx context.Context, req *mgmtpb.ListPipelineTriggerChartRecordsV0Request) (*mgmtpb.ListPipelineTriggerChartRecordsV0Response, error) {

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		return nil, err
	}
	pbUser, err := h.Service.GetUserByUIDAdmin(ctx, ctxUserUID)
	if err != nil {
		return nil, err
	}

	var mode mgmtpb.Mode
	var status mgmtpb.Status

	declarations, err := filtering.NewDeclarations([]filtering.DeclarationOption{
		filtering.DeclareStandardFunctions(),
		filtering.DeclareIdent(constant.Start, filtering.TypeTimestamp),
		filtering.DeclareIdent(constant.Stop, filtering.TypeTimestamp),
		filtering.DeclareIdent(strcase.ToLowerCamel(constant.OwnerName), filtering.TypeString),
		filtering.DeclareIdent(strcase.ToLowerCamel(constant.PipelineID), filtering.TypeString),
		filtering.DeclareIdent(strcase.ToLowerCamel(constant.PipelineUID), filtering.TypeString),
		filtering.DeclareIdent(strcase.ToLowerCamel(constant.PipelineReleaseID), filtering.TypeString),
		filtering.DeclareIdent(strcase.ToLowerCamel(constant.PipelineReleaseUID), filtering.TypeString),
		filtering.DeclareEnumIdent(strcase.ToLowerCamel(constant.TriggerMode), mode.Type()),
		filtering.DeclareEnumIdent(constant.Status, status.Type()),
	}...)
	if err != nil {
		return nil, err
	}

	filter, err := filtering.ParseFilter(req, declarations)
	if err != nil {
		return nil, err
	}

	pipelineTriggerChartRecords, err := h.Service.ListPipelineTriggerChartRecordsV0(ctx, pbUser, int64(req.GetAggregationWindow()), filter)
	if err != nil {
		return nil, err
	}

	resp := mgmtpb.ListPipelineTriggerChartRecordsV0Response{
		PipelineTriggerChartRecords: pipelineTriggerChartRecords,
	}

	return &resp, nil
}

// ListModelTriggerChartRecords returns a timeline of model trigger counts for a given requester. The
// response will contain one set of records (datapoints), representing the amount of triggers in a time bucket.
func (h *PublicHandler) ListModelTriggerChartRecords(ctx context.Context, req *mgmtpb.ListModelTriggerChartRecordsRequest) (*mgmtpb.ListModelTriggerChartRecordsResponse, error) {

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		return nil, err
	}

	resp, err := h.Service.ListModelTriggerChartRecords(ctx, req, ctxUserUID)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// ListUserMemberships lists user memberships.
func (h *PublicHandler) ListUserMemberships(ctx context.Context, req *mgmtpb.ListUserMembershipsRequest) (*mgmtpb.ListUserMembershipsResponse, error) {

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		return nil, err
	}

	pbMemberships, err := h.Service.ListUserMemberships(ctx, ctxUserUID, req.UserId)
	if err != nil {
		return nil, err
	}

	resp := mgmtpb.ListUserMembershipsResponse{
		Memberships: pbMemberships,
	}

	return &resp, nil
}

// GetUserMembership gets a user membership.
func (h *PublicHandler) GetUserMembership(ctx context.Context, req *mgmtpb.GetUserMembershipRequest) (*mgmtpb.GetUserMembershipResponse, error) {

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		return nil, err
	}

	pbMembership, err := h.Service.GetUserMembership(ctx, ctxUserUID, req.UserId, req.OrganizationId)
	if err != nil {
		return nil, err
	}

	resp := mgmtpb.GetUserMembershipResponse{
		Membership: pbMembership,
	}

	return &resp, nil
}

// UpdateUserMembership updates a user membership.
func (h *PublicHandler) UpdateUserMembership(ctx context.Context, req *mgmtpb.UpdateUserMembershipRequest) (*mgmtpb.UpdateUserMembershipResponse, error) {

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		return nil, err
	}

	if req.UpdateMask == nil || len(req.UpdateMask.Paths) == 0 {
		return nil, errorsx.ErrFieldMask
	}

	if err := checkfield.CheckRequiredFields(req.Membership, requiredFieldsForUserMembership); err != nil {
		return nil, errorsx.ErrCheckRequiredFields
	}

	if err := checkfield.CheckCreateOutputOnlyFields(req.Membership, outputOnlyFieldsForUserMembership); err != nil {
		return nil, errorsx.ErrCheckOutputOnlyFields
	}

	pbMembership, err := h.Service.UpdateUserMembership(ctx, ctxUserUID, req.UserId, req.OrganizationId, req.Membership, req.UpdateMask)
	if err != nil {
		return nil, err
	}

	resp := mgmtpb.UpdateUserMembershipResponse{
		Membership: pbMembership,
	}

	return &resp, nil
}

// DeleteUserMembership deletes a user membership.
func (h *PublicHandler) DeleteUserMembership(ctx context.Context, req *mgmtpb.DeleteUserMembershipRequest) (*mgmtpb.DeleteUserMembershipResponse, error) {

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		return nil, err
	}

	err = h.Service.DeleteUserMembership(ctx, ctxUserUID, req.UserId, req.OrganizationId)
	if err != nil {
		return nil, err
	}

	resp := mgmtpb.DeleteUserMembershipResponse{}

	return &resp, nil
}

// ListOrganizationMemberships lists organization memberships.
func (h *PublicHandler) ListOrganizationMemberships(ctx context.Context, req *mgmtpb.ListOrganizationMembershipsRequest) (*mgmtpb.ListOrganizationMembershipsResponse, error) {

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		return nil, err
	}

	pbMemberships, err := h.Service.ListOrganizationMemberships(ctx, ctxUserUID, req.OrganizationId)
	if err != nil {
		return nil, err
	}

	resp := mgmtpb.ListOrganizationMembershipsResponse{
		Memberships: pbMemberships,
	}

	return &resp, nil
}

// GetOrganizationMembership gets an organization membership.
func (h *PublicHandler) GetOrganizationMembership(ctx context.Context, req *mgmtpb.GetOrganizationMembershipRequest) (*mgmtpb.GetOrganizationMembershipResponse, error) {

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		return nil, err
	}

	pbMembership, err := h.Service.GetOrganizationMembership(ctx, ctxUserUID, req.OrganizationId, req.UserId)
	if err != nil {
		return nil, err
	}

	resp := mgmtpb.GetOrganizationMembershipResponse{
		Membership: pbMembership,
	}

	return &resp, nil
}

// UpdateOrganizationMembership updates an organization membership.
func (h *PublicHandler) UpdateOrganizationMembership(ctx context.Context, req *mgmtpb.UpdateOrganizationMembershipRequest) (*mgmtpb.UpdateOrganizationMembershipResponse, error) {

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		return nil, err
	}

	if req.UpdateMask == nil || len(req.UpdateMask.Paths) == 0 {
		return nil, errorsx.ErrFieldMask
	}

	if err := checkfield.CheckRequiredFields(req.Membership, requiredFieldsForOrganizationMembership); err != nil {
		return nil, errorsx.ErrCheckRequiredFields
	}

	if err := checkfield.CheckCreateOutputOnlyFields(req.Membership, outputOnlyFieldsForOrganizationMembership); err != nil {
		return nil, errorsx.ErrCheckOutputOnlyFields
	}

	pbMembership, err := h.Service.UpdateOrganizationMembership(ctx, ctxUserUID, req.OrganizationId, req.UserId, req.Membership, req.UpdateMask)
	if err != nil {
		return nil, err
	}

	resp := mgmtpb.UpdateOrganizationMembershipResponse{
		Membership: pbMembership,
	}

	return &resp, nil
}

// DeleteOrganizationMembership deletes an organization membership.
func (h *PublicHandler) DeleteOrganizationMembership(ctx context.Context, req *mgmtpb.DeleteOrganizationMembershipRequest) (*mgmtpb.DeleteOrganizationMembershipResponse, error) {

	ctxUserUID, err := h.Service.ExtractCtxUser(ctx, false)
	if err != nil {
		return nil, err
	}

	err = h.Service.DeleteOrganizationMembership(ctx, ctxUserUID, req.OrganizationId, req.UserId)
	if err != nil {
		return nil, err
	}

	return &mgmtpb.DeleteOrganizationMembershipResponse{}, nil
}
