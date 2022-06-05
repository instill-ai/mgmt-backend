package handler

import (
	"context"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/iancoleman/strcase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	fieldmask_utils "github.com/mennanov/fieldmask-utils"

	checkfield "github.com/instill-ai/x/checkfield"

	"github.com/instill-ai/mgmt-backend/pkg/service"

	healthcheckPB "github.com/instill-ai/protogen-go/vdp/healthcheck/v1alpha"
	mgmtv1alpha "github.com/instill-ai/protogen-go/vdp/mgmt/v1alpha"
)

// TODO: Validate mask based on the field behavior.
// Currently, the OUTPUT_ONLY fields are hard-coded.
var outputOnlyFields = []string{"name", "uid", "type", "create_time", "update_time"}
var immutableFields = []string{"id"}

const defaultPageSize = int64(10)
const maxPageSize = int64(100)

type handler struct {
	mgmtv1alpha.UnimplementedUserServiceServer
	service service.Service
}

// NewHandler initiates a handler instance
func NewHandler(s service.Service) mgmtv1alpha.UserServiceServer {
	return &handler{
		service: s,
	}
}

// Liveness checks the liveness of the server
func (h *handler) Liveness(ctx context.Context, in *mgmtv1alpha.LivenessRequest) (*mgmtv1alpha.LivenessResponse, error) {
	return &mgmtv1alpha.LivenessResponse{
		HealthCheckResponse: &healthcheckPB.HealthCheckResponse{
			Status: healthcheckPB.HealthCheckResponse_SERVING_STATUS_SERVING,
		},
	}, nil
}

// Readiness checks the readiness of the server
func (h *handler) Readiness(ctx context.Context, in *mgmtv1alpha.ReadinessRequest) (*mgmtv1alpha.ReadinessResponse, error) {
	return &mgmtv1alpha.ReadinessResponse{
		HealthCheckResponse: &healthcheckPB.HealthCheckResponse{
			Status: healthcheckPB.HealthCheckResponse_SERVING_STATUS_SERVING,
		},
	}, nil
}

// ListUser lists all users
func (h *handler) ListUser(ctx context.Context, req *mgmtv1alpha.ListUserRequest) (*mgmtv1alpha.ListUserResponse, error) {

	pageSize := req.GetPageSize()
	if pageSize == 0 {
		pageSize = defaultPageSize
	} else if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	dbUsers, nextPageToken, totalSize, err := h.service.ListUser(int(pageSize), req.GetPageToken())

	if err != nil {
		return &mgmtv1alpha.ListUserResponse{}, err
	}

	pbUsers := []*mgmtv1alpha.User{}
	for _, dbUser := range dbUsers {
		pbUser, err := DBUser2PBUser(&dbUser)
		if err != nil {
			return &mgmtv1alpha.ListUserResponse{}, err
		}
		pbUsers = append(pbUsers, pbUser)
	}

	resp := mgmtv1alpha.ListUserResponse{
		Users:         pbUsers,
		NextPageToken: nextPageToken,
		TotalSize:     int64(totalSize),
	}
	return &resp, nil
}

// CreateUser creates a user. This endpoint is not supported.
func (h *handler) CreateUser(ctx context.Context, req *mgmtv1alpha.CreateUserRequest) (*mgmtv1alpha.CreateUserResponse, error) {
	// Validate the user id conforms to RFC-1034, which restricts to letters, numbers,
	// and hyphen, with the first character a letter, the last a letter or a
	// number, and a 63 character maximum.
	id := req.GetUser().GetId()
	err := checkfield.CheckResourceID(id)
	if err != nil {
		return &mgmtv1alpha.CreateUserResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}

	return &mgmtv1alpha.CreateUserResponse{}, status.Error(codes.Unimplemented, "this endpoint is not supported")
}

// GetUser gets a user
func (h *handler) GetUser(ctx context.Context, req *mgmtv1alpha.GetUserRequest) (*mgmtv1alpha.GetUserResponse, error) {
	id := strings.TrimPrefix(req.GetName(), "users/")

	dbUser, err := h.service.GetUserByID(id)
	if err != nil {
		return &mgmtv1alpha.GetUserResponse{}, err
	}

	pbUser, err := DBUser2PBUser(dbUser)
	if err != nil {
		return &mgmtv1alpha.GetUserResponse{}, err
	}

	resp := mgmtv1alpha.GetUserResponse{
		User: pbUser,
	}
	return &resp, nil
}

// LookUpUser gets a user by permalink
func (h *handler) LookUpUser(ctx context.Context, req *mgmtv1alpha.LookUpUserRequest) (*mgmtv1alpha.LookUpUserResponse, error) {
	uidStr := strings.TrimPrefix(req.GetPermalink(), "users/")
	// Validation: `uid` in request is valid
	uid, err := uuid.FromString(uidStr)
	if err != nil {
		return &mgmtv1alpha.LookUpUserResponse{}, status.Error(codes.InvalidArgument, "permalink is invalid")
	}

	dbUser, err := h.service.GetUser(uid)
	if err != nil {
		return &mgmtv1alpha.LookUpUserResponse{}, err
	}

	pbUser, err := DBUser2PBUser(dbUser)
	if err != nil {
		return &mgmtv1alpha.LookUpUserResponse{}, err
	}
	resp := mgmtv1alpha.LookUpUserResponse{
		User: pbUser,
	}
	return &resp, nil
}

// UpdateUser updates an existing user
func (h *handler) UpdateUser(ctx context.Context, req *mgmtv1alpha.UpdateUserRequest) (*mgmtv1alpha.UpdateUserResponse, error) {
	reqUser := req.GetUser()

	// Validate the field mask
	if !req.GetUpdateMask().IsValid(reqUser) {
		return &mgmtv1alpha.UpdateUserResponse{}, status.Error(codes.InvalidArgument, "`update_mask` is invalid")
	}

	reqFieldMask, err := checkfield.CheckUpdateOutputOnlyFields(req.GetUpdateMask(), outputOnlyFields)
	if err != nil {
		return &mgmtv1alpha.UpdateUserResponse{}, err
	}

	mask, err := fieldmask_utils.MaskFromProtoFieldMask(reqFieldMask, strcase.ToCamel)
	if err != nil {
		return &mgmtv1alpha.UpdateUserResponse{}, err
	}

	// Get current user
	GResp, err := h.GetUser(ctx, &mgmtv1alpha.GetUserRequest{Name: reqUser.GetName()})
	if err != nil {
		return &mgmtv1alpha.UpdateUserResponse{}, err
	}
	pbUserToUpdate := GResp.GetUser()

	if mask.IsEmpty() {
		// return the un-changed user `pbUserToUpdate`
		resp := mgmtv1alpha.UpdateUserResponse{
			User: pbUserToUpdate,
		}
		return &resp, nil
	}

	// the current user `pbUserToUpdate`: a struct to copy to
	uid, err := uuid.FromString(pbUserToUpdate.GetUid())
	if err != nil {
		return &mgmtv1alpha.UpdateUserResponse{}, err
	}

	// Handle immutable fields from the update mask
	err = checkfield.CheckUpdateImmutableFields(reqUser, pbUserToUpdate, immutableFields)
	if err != nil {
		return &mgmtv1alpha.UpdateUserResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}

	// // Handle IMMUTABLE fields from the update mask:
	// // TODO: we hard coded the IMMUTABLE field "Id" here
	// _, ok := mask.Filter("Id")
	// if ok {
	// 	if reqUser.GetId() != pbUserToUpdate.GetId() {
	// 		return &mgmtv1alpha.UpdateUserResponse{}, status.Error(codes.InvalidArgument, "`id` is not allowed to be updated")
	// 	}
	// }

	// Only the fields mentioned in the field mask will be copied to `pbUserToUpdate`, other fields are left intact
	err = fieldmask_utils.StructToStruct(mask, reqUser, pbUserToUpdate)
	if err != nil {
		return &mgmtv1alpha.UpdateUserResponse{}, err
	}

	dbUserToUpd, err := PBUser2DBUser(pbUserToUpdate)
	if err != nil {
		return &mgmtv1alpha.UpdateUserResponse{}, err
	}

	dbUserUpdated, err := h.service.UpdateUser(uid, dbUserToUpd)
	if err != nil {
		return &mgmtv1alpha.UpdateUserResponse{}, err
	}

	pbUserUpdated, err := DBUser2PBUser(dbUserUpdated)
	if err != nil {
		return &mgmtv1alpha.UpdateUserResponse{}, err
	}
	resp := mgmtv1alpha.UpdateUserResponse{
		User: pbUserUpdated,
	}
	return &resp, nil
}

// DeleteUser deletes a user. This endpoint is not supported.
func (h *handler) DeleteUser(ctx context.Context, req *mgmtv1alpha.DeleteUserRequest) (*mgmtv1alpha.DeleteUserResponse, error) {
	return &mgmtv1alpha.DeleteUserResponse{}, status.Error(codes.Unimplemented, "this endpoint is not supported")
}
