package handler

import (
	"context"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/iancoleman/strcase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	fieldmask_utils "github.com/mennanov/fieldmask-utils"

	"github.com/instill-ai/mgmt-backend/pkg/service"

	mgmtPB "github.com/instill-ai/protogen-go/mgmt/v1alpha"
)

// TODO: Validate mask based on the field behavior.
// Currently, the OUTPUT_ONLY fields are hard-coded.
var outputOnlyFields = []string{"Name", "Id", "Type", "CreateTime", "UpdateTime"}

type handler struct {
	mgmtPB.UnimplementedUserServiceServer
	service service.Service
}

// NewHandler initiates a handler instance
func NewHandler(s service.Service) mgmtPB.UserServiceServer {
	return &handler{
		service: s,
	}
}

// Liveness checks the liveness of the server
func (h *handler) Liveness(ctx context.Context, in *mgmtPB.LivenessRequest) (*mgmtPB.LivenessResponse, error) {
	return &mgmtPB.LivenessResponse{
		HealthCheckResponse: &mgmtPB.HealthCheckResponse{
			Status: mgmtPB.HealthCheckResponse_SERVING_STATUS_SERVING,
		},
	}, nil
}

// Readiness checks the readiness of the server
func (h *handler) Readiness(ctx context.Context, in *mgmtPB.ReadinessRequest) (*mgmtPB.ReadinessResponse, error) {
	return &mgmtPB.ReadinessResponse{
		HealthCheckResponse: &mgmtPB.HealthCheckResponse{
			Status: mgmtPB.HealthCheckResponse_SERVING_STATUS_SERVING,
		},
	}, nil
}

// ListUser lists all users
func (h *handler) ListUser(ctx context.Context, req *mgmtPB.ListUserRequest) (*mgmtPB.ListUserResponse, error) {

	dbUsers, nextPageToken, totalSize, err := h.service.ListUser(int(req.GetPageSize()), req.GetPageToken())

	if err != nil {
		return &mgmtPB.ListUserResponse{}, err
	}

	pbUsers := []*mgmtPB.User{}
	for _, dbUser := range dbUsers {
		pbUsers = append(pbUsers, DBUser2PBUser(&dbUser))
	}

	resp := mgmtPB.ListUserResponse{
		Users:         pbUsers,
		NextPageToken: nextPageToken,
		TotalSize:     int32(totalSize),
	}
	return &resp, nil
}

// CreateUser creates a user. This endpoint is not supported.
func (h *handler) CreateUser(ctx context.Context, req *mgmtPB.CreateUserRequest) (*mgmtPB.CreateUserResponse, error) {

	return &mgmtPB.CreateUserResponse{User: &mgmtPB.User{}}, status.Error(codes.Unimplemented, "This endpoint is not supported")
}

// GetUser gets a user
func (h *handler) GetUser(ctx context.Context, req *mgmtPB.GetUserRequest) (*mgmtPB.GetUserResponse, error) {
	login := strings.TrimPrefix(req.GetName(), "users/")

	dbUser, err := h.service.GetUserByLogin(login)
	if err != nil {
		return &mgmtPB.GetUserResponse{}, err
	}
	pbUser := DBUser2PBUser(dbUser)

	resp := mgmtPB.GetUserResponse{
		User: pbUser,
	}
	return &resp, nil
}

// UpdateUser updates an existing user
func (h *handler) UpdateUser(ctx context.Context, req *mgmtPB.UpdateUserRequest) (*mgmtPB.UpdateUserResponse, error) {
	reqUser := req.GetUser()
	reqFieldMask := req.GetUpdateMask()

	// Validate the field mask
	if !reqFieldMask.IsValid(reqUser) {
		return &mgmtPB.UpdateUserResponse{}, status.Error(codes.FailedPrecondition, "The `update_mask` is invalid")
	}

	mask, err := fieldmask_utils.MaskFromProtoFieldMask(reqFieldMask, strcase.ToCamel)
	if err != nil {
		return &mgmtPB.UpdateUserResponse{}, err
	}

	for _, field := range outputOnlyFields {
		_, ok := mask.Filter(field)
		if ok {
			delete(mask, field)
		}
	}

	if mask.IsEmpty() {
		// return the un-changed user `pbUserToUpd`
		GResp, err := h.GetUser(ctx, &mgmtPB.GetUserRequest{Name: reqUser.GetName()})
		if err != nil {
			return &mgmtPB.UpdateUserResponse{}, err
		}

		resp := mgmtPB.UpdateUserResponse{
			User: GResp.GetUser(),
		}
		return &resp, nil
	}

	// the current user `pbUserToUpd`: a struct to copy to
	GResp, err := h.GetUser(ctx, &mgmtPB.GetUserRequest{Name: reqUser.GetName()})
	if err != nil {
		return &mgmtPB.UpdateUserResponse{}, err
	}
	pbUserToUpd := GResp.GetUser()
	id, err := uuid.FromString(pbUserToUpd.GetId())
	if err != nil {
		return &mgmtPB.UpdateUserResponse{}, err
	}

	// Only the fields mentioned in the field mask will be copied to `pbUserToUpd`, other fields are left intact
	err = fieldmask_utils.StructToStruct(mask, reqUser, pbUserToUpd)
	if err != nil {
		return &mgmtPB.UpdateUserResponse{}, err
	}

	dbUserToUpd, err := PBUser2DBUser(pbUserToUpd)
	if err != nil {
		return &mgmtPB.UpdateUserResponse{}, err
	}

	dbUserUpdated, err := h.service.UpdateUser(id, dbUserToUpd)
	if err != nil {
		return &mgmtPB.UpdateUserResponse{}, err
	}

	pbUserUpdated := DBUser2PBUser(dbUserUpdated)
	resp := mgmtPB.UpdateUserResponse{
		User: pbUserUpdated,
	}
	return &resp, nil
}

// DeleteUser deletes a user. This endpoint is not supported.
func (h *handler) DeleteUser(ctx context.Context, req *mgmtPB.DeleteUserRequest) (*mgmtPB.DeleteUserResponse, error) {
	return &mgmtPB.DeleteUserResponse{}, status.Error(codes.Unimplemented, "This endpoint is not supported")
}
