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
var outputOnlyFields = []string{"Name", "Uid", "Type", "CreateTime", "UpdateTime"}

const defaultPageSize = int32(10)
const maxPageSize = int32(100)

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

	pageSize := req.GetPageSize()
	if pageSize == 0 {
		pageSize = defaultPageSize
	} else if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	dbUsers, nextPageToken, totalSize, err := h.service.ListUser(int(pageSize), req.GetPageToken())

	if err != nil {
		return &mgmtPB.ListUserResponse{}, err
	}

	pbUsers := []*mgmtPB.User{}
	for _, dbUser := range dbUsers {
		pbUser, err := DBUser2PBUser(&dbUser)
		if err != nil {
			return &mgmtPB.ListUserResponse{}, err
		}
		pbUsers = append(pbUsers, pbUser)
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
	// TODO: validate the user id conforms to RFC-1034, which restricts to letters, numbers,
	// and hyphen, with the first character a letter, the last a letter or a
	// number, and a 63 character maximum.
	return &mgmtPB.CreateUserResponse{User: &mgmtPB.User{}}, status.Error(codes.Unimplemented, "this endpoint is not supported")
}

// GetUser gets a user
func (h *handler) GetUser(ctx context.Context, req *mgmtPB.GetUserRequest) (*mgmtPB.GetUserResponse, error) {
	id := strings.TrimPrefix(req.GetName(), "users/")

	dbUser, err := h.service.GetUserByID(id)
	if err != nil {
		return &mgmtPB.GetUserResponse{}, err
	}
	pbUser, err := DBUser2PBUser(dbUser)
	if err != nil {
		return &mgmtPB.GetUserResponse{}, err
	}

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
		return &mgmtPB.UpdateUserResponse{}, status.Error(codes.InvalidArgument, "`update_mask` is invalid")
	}

	mask, err := fieldmask_utils.MaskFromProtoFieldMask(reqFieldMask, strcase.ToCamel)
	if err != nil {
		return &mgmtPB.UpdateUserResponse{}, err
	}

	// Remove OUTPUT_ONLY fields from the update mask
	for _, field := range outputOnlyFields {
		_, ok := mask.Filter(field)
		if ok {
			delete(mask, field)
		}
	}

	// Get current user
	GResp, err := h.GetUser(ctx, &mgmtPB.GetUserRequest{Name: reqUser.GetName()})
	if err != nil {
		return &mgmtPB.UpdateUserResponse{}, err
	}
	pbUserToUpdate := GResp.GetUser()

	if mask.IsEmpty() {
		// return the un-changed user `pbUserToUpdate`
		resp := mgmtPB.UpdateUserResponse{
			User: pbUserToUpdate,
		}
		return &resp, nil
	}

	// the current user `pbUserToUpdate`: a struct to copy to
	uid, err := uuid.FromString(pbUserToUpdate.GetUid())
	if err != nil {
		return &mgmtPB.UpdateUserResponse{}, err
	}

	// Handle IMMUTABLE fields from the update mask:
	// TODO: we hard coded the IMMUTABLE field "Id" here
	_, ok := mask.Filter("Id")
	if ok {
		if reqUser.GetId() != pbUserToUpdate.GetId() {
			return &mgmtPB.UpdateUserResponse{}, status.Error(codes.InvalidArgument, "`id` is not allowed to be updated")
		}
	}

	// Only the fields mentioned in the field mask will be copied to `pbUserToUpdate`, other fields are left intact
	err = fieldmask_utils.StructToStruct(mask, reqUser, pbUserToUpdate)
	if err != nil {
		return &mgmtPB.UpdateUserResponse{}, err
	}

	dbUserToUpd, err := PBUser2DBUser(pbUserToUpdate)
	if err != nil {
		return &mgmtPB.UpdateUserResponse{}, err
	}

	dbUserUpdated, err := h.service.UpdateUser(uid, dbUserToUpd)
	if err != nil {
		return &mgmtPB.UpdateUserResponse{}, err
	}

	pbUserUpdated, err := DBUser2PBUser(dbUserUpdated)
	if err != nil {
		return &mgmtPB.UpdateUserResponse{}, err
	}
	resp := mgmtPB.UpdateUserResponse{
		User: pbUserUpdated,
	}
	return &resp, nil
}

// DeleteUser deletes a user. This endpoint is not supported.
func (h *handler) DeleteUser(ctx context.Context, req *mgmtPB.DeleteUserRequest) (*mgmtPB.DeleteUserResponse, error) {
	return &mgmtPB.DeleteUserResponse{}, status.Error(codes.Unimplemented, "this endpoint is not supported")
}
