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
	"github.com/instill-ai/mgmt-backend/pkg/datamodel"
	"github.com/instill-ai/mgmt-backend/pkg/logger"
	"github.com/instill-ai/mgmt-backend/pkg/service"
	"github.com/instill-ai/x/sterr"

	mgmtPB "github.com/instill-ai/protogen-go/vdp/mgmt/v1alpha"
	checkfield "github.com/instill-ai/x/checkfield"
)

const defaultPageSize = int64(10)
const maxPageSize = int64(100)

type PrivateHandler struct {
	mgmtPB.UnimplementedMgmtPrivateServiceServer
	Service service.Service
}

// NewPrivateHandler initiates an private handler instance
func NewPrivateHandler(s service.Service) mgmtPB.MgmtPrivateServiceServer {
	return &PrivateHandler{
		Service: s,
	}
}

// ListUsersAdmin lists all users
func (h *PrivateHandler) ListUsersAdmin(ctx context.Context, req *mgmtPB.ListUsersAdminRequest) (*mgmtPB.ListUsersAdminResponse, error) {
	logger, _ := logger.GetZapLogger(config.Config.Server.Debug)

	pageSize := req.GetPageSize()
	if pageSize == 0 {
		pageSize = defaultPageSize
	} else if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	dbUsers, nextPageToken, totalSize, err := h.Service.ListUser(int(pageSize), req.GetPageToken())
	if err != nil {
		sta := status.Convert(err)
		switch sta.Code() {
		case codes.InvalidArgument:
			st, e := sterr.CreateErrorBadRequest(
				"list user error", []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "ListUsersAdminRequest.page_token",
						Description: sta.Message(),
					},
				})
			if e != nil {
				logger.Error(e.Error())
			}
			return &mgmtPB.ListUsersAdminResponse{}, st.Err()
		default:
			st, e := sterr.CreateErrorResourceInfo(
				sta.Code(),
				"list user error",
				"user",
				"",
				"",
				sta.Message(),
			)
			if e != nil {
				logger.Error(e.Error())
			}
			return &mgmtPB.ListUsersAdminResponse{}, st.Err()
		}
	}

	pbUsers := []*mgmtPB.User{}
	for _, dbUser := range dbUsers {
		pbUser, err := datamodel.DBUser2PBUser(&dbUser)
		if err != nil {
			logger.Error(err.Error())
			st, e := sterr.CreateErrorResourceInfo(
				codes.Internal,
				"list user error",
				"user",
				fmt.Sprintf("id %s", dbUser.ID),
				"",
				err.Error(),
			)
			if e != nil {
				logger.Error(e.Error())
			}
			return &mgmtPB.ListUsersAdminResponse{}, st.Err()
		}
		pbUsers = append(pbUsers, pbUser)
	}

	resp := mgmtPB.ListUsersAdminResponse{
		Users:         pbUsers,
		NextPageToken: nextPageToken,
		TotalSize:     totalSize,
	}
	return &resp, nil
}

// CreateUserAdmin creates a user. This endpoint is not supported yet.
func (h *PrivateHandler) CreateUserAdmin(ctx context.Context, req *mgmtPB.CreateUserAdminRequest) (*mgmtPB.CreateUserAdminResponse, error) {
	logger, _ := logger.GetZapLogger(config.Config.Server.Debug)
	resp := &mgmtPB.CreateUserAdminResponse{}

	// Return error if REQUIRED fields are not provided in the requested payload resource
	if err := checkfield.CheckRequiredFields(req.GetUser(), createRequiredFields); err != nil {
		st, e := sterr.CreateErrorBadRequest(
			"create user bad request error", []*errdetails.BadRequest_FieldViolation{
				{
					Field:       fmt.Sprintf("%v", createRequiredFields),
					Description: err.Error(),
				},
			},
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return resp, st.Err()
	}

	// Validate the user id conforms to RFC-1034, which restricts to letters, numbers,
	// and hyphen, with the first character a letter, the last a letter or a
	// number, and a 63 character maximum.
	id := req.GetUser().GetId()
	err := checkfield.CheckResourceID(id)
	if err != nil {
		st, e := sterr.CreateErrorBadRequest(
			"create user bad request error", []*errdetails.BadRequest_FieldViolation{
				{
					Field:       "id",
					Description: err.Error(),
				},
			},
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return resp, st.Err()
	}

	st, err := sterr.CreateErrorResourceInfo(
		codes.Unimplemented,
		"create user not implemented error",
		"endpoint",
		"/users",
		"",
		"not implemented",
	)
	if err != nil {
		logger.Error(err.Error())
	}
	return resp, st.Err()
}

// GetUserAdmin gets a user
func (h *PrivateHandler) GetUserAdmin(ctx context.Context, req *mgmtPB.GetUserAdminRequest) (*mgmtPB.GetUserAdminResponse, error) {
	logger, _ := logger.GetZapLogger(config.Config.Server.Debug)

	id := strings.TrimPrefix(req.GetName(), "users/")

	dbUser, err := h.Service.GetUserByID(id)
	if err != nil {
		sta := status.Convert(err)
		switch sta.Code() {
		case codes.InvalidArgument:
			st, e := sterr.CreateErrorBadRequest(
				"get user error", []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "GetUserAdminRequest.name",
						Description: sta.Message(),
					},
				})
			if e != nil {
				logger.Error(e.Error())
			}
			return &mgmtPB.GetUserAdminResponse{}, st.Err()
		default:
			st, e := sterr.CreateErrorResourceInfo(
				sta.Code(),
				"get user error",
				"user",
				fmt.Sprintf("id %s", id),
				"",
				sta.Message(),
			)
			if e != nil {
				logger.Error(e.Error())
			}
			return &mgmtPB.GetUserAdminResponse{}, st.Err()
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
		return &mgmtPB.GetUserAdminResponse{}, st.Err()
	}

	resp := mgmtPB.GetUserAdminResponse{
		User: pbUser,
	}
	return &resp, nil
}

// LookUpUserAdmin gets a user by permalink
func (h *PrivateHandler) LookUpUserAdmin(ctx context.Context, req *mgmtPB.LookUpUserAdminRequest) (*mgmtPB.LookUpUserAdminResponse, error) {
	logger, _ := logger.GetZapLogger(config.Config.Server.Debug)

	uidStr := strings.TrimPrefix(req.GetPermalink(), "users/")
	// Validation: `uid` in request is valid
	uid, err := uuid.FromString(uidStr)
	if err != nil {
		st, e := sterr.CreateErrorBadRequest(
			"look up user invalid uuid error", []*errdetails.BadRequest_FieldViolation{
				{
					Field:       "LookUpUserAdminRequest.permalink",
					Description: err.Error(),
				},
			},
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return &mgmtPB.LookUpUserAdminResponse{}, st.Err()
	}

	dbUser, err := h.Service.GetUser(uid)
	if err != nil {
		sta := status.Convert(err)
		switch sta.Code() {
		case codes.InvalidArgument:
			st, e := sterr.CreateErrorBadRequest(
				"look up user error", []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "LookUpUserAdminRequest.permalink",
						Description: sta.Message(),
					},
				})
			if e != nil {
				logger.Error(e.Error())
			}
			return &mgmtPB.LookUpUserAdminResponse{}, st.Err()
		default:
			st, e := sterr.CreateErrorResourceInfo(
				sta.Code(),
				"look up user error",
				"user",
				fmt.Sprintf("uid %s", uid),
				"",
				sta.Message(),
			)
			if e != nil {
				logger.Error(e.Error())
			}
			return &mgmtPB.LookUpUserAdminResponse{}, st.Err()
		}
	}

	pbUser, err := datamodel.DBUser2PBUser(dbUser)
	if err != nil {
		logger.Error(err.Error())
		st, e := sterr.CreateErrorResourceInfo(
			codes.Internal,
			"look up user error",
			"user",
			fmt.Sprintf("uid %s", dbUser.UID),
			"",
			err.Error(),
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return &mgmtPB.LookUpUserAdminResponse{}, st.Err()
	}
	resp := mgmtPB.LookUpUserAdminResponse{
		User: pbUser,
	}
	return &resp, nil
}

// UpdateUserAdmin updates an existing user
func (h *PrivateHandler) UpdateUserAdmin(ctx context.Context, req *mgmtPB.UpdateUserAdminRequest) (*mgmtPB.UpdateUserAdminResponse, error) {
	logger, _ := logger.GetZapLogger(config.Config.Server.Debug)

	reqUser := req.GetUser()

	// Validate the field mask
	if !req.GetUpdateMask().IsValid(reqUser) {
		st, e := sterr.CreateErrorBadRequest(
			"update user invalid fieldmask error", []*errdetails.BadRequest_FieldViolation{
				{
					Field:       "UpdateUserAdminRequest.update_mask",
					Description: "invalid",
				},
			},
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return &mgmtPB.UpdateUserAdminResponse{}, st.Err()
	}

	reqFieldMask, err := checkfield.CheckUpdateOutputOnlyFields(req.GetUpdateMask(), outputOnlyFields)
	if err != nil {
		st, e := sterr.CreateErrorBadRequest(
			"update user update OUTPUT_ONLY fields error", []*errdetails.BadRequest_FieldViolation{
				{
					Field:       "UpdateUserAdminRequest OUTPUT_ONLY fields",
					Description: err.Error(),
				},
			},
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return &mgmtPB.UpdateUserAdminResponse{}, st.Err()
	}

	mask, err := fieldmask_utils.MaskFromProtoFieldMask(reqFieldMask, strcase.ToCamel)
	if err != nil {
		logger.Error(err.Error())
		st, e := sterr.CreateErrorBadRequest(
			"update user update mask error", []*errdetails.BadRequest_FieldViolation{
				{
					Field:       "UpdateUserAdminRequest.update_mask",
					Description: err.Error(),
				},
			},
		)
		if e != nil {
			logger.Error(e.Error())
		}

		return &mgmtPB.UpdateUserAdminResponse{}, st.Err()
	}

	// Get current user
	GResp, err := h.GetUserAdmin(ctx, &mgmtPB.GetUserAdminRequest{Name: reqUser.GetName()})
	if err != nil {
		return &mgmtPB.UpdateUserAdminResponse{}, err
	}
	pbUserToUpdate := GResp.GetUser()

	if mask.IsEmpty() {
		// return the un-changed user `pbUserToUpdate`
		resp := mgmtPB.UpdateUserAdminResponse{
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
			"update user error",
			"user",
			fmt.Sprintf("user %v", pbUserToUpdate),
			"",
			err.Error(),
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return &mgmtPB.UpdateUserAdminResponse{}, st.Err()
	}

	// Handle immutable fields from the update mask
	err = checkfield.CheckUpdateImmutableFields(reqUser, pbUserToUpdate, immutableFields)
	if err != nil {
		st, e := sterr.CreateErrorBadRequest(
			"update user update IMMUTABLE fields error", []*errdetails.BadRequest_FieldViolation{
				{
					Field:       "UpdateUserAdminRequest IMMUTABLE fields",
					Description: err.Error(),
				},
			},
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return &mgmtPB.UpdateUserAdminResponse{}, st.Err()
	}

	// Only the fields mentioned in the field mask will be copied to `pbUserToUpdate`, other fields are left intact
	err = fieldmask_utils.StructToStruct(mask, reqUser, pbUserToUpdate)
	if err != nil {
		logger.Error(err.Error())
		st, e := sterr.CreateErrorResourceInfo(
			codes.Internal,
			"update user error", "user", fmt.Sprintf("uid %s", *reqUser.Uid),
			"",
			err.Error(),
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return &mgmtPB.UpdateUserAdminResponse{}, st.Err()
	}

	dbUserToUpd, err := datamodel.PBUser2DBUser(pbUserToUpdate)
	if err != nil {
		logger.Error(err.Error())
		st, e := sterr.CreateErrorResourceInfo(
			codes.Internal,
			"update user error",
			"user",
			fmt.Sprintf("id %s", pbUserToUpdate.GetId()),
			"",
			err.Error(),
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return &mgmtPB.UpdateUserAdminResponse{}, st.Err()
	}

	dbUserUpdated, err := h.Service.UpdateUser(uid, dbUserToUpd)
	if err != nil {
		sta := status.Convert(err)
		switch sta.Code() {
		case codes.InvalidArgument:
			st, e := sterr.CreateErrorBadRequest(
				"update user error", []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "UpdateUserAdminRequest",
						Description: sta.Message(),
					},
				})
			if e != nil {
				logger.Error(e.Error())
			}
			return &mgmtPB.UpdateUserAdminResponse{}, st.Err()
		default:
			st, e := sterr.CreateErrorResourceInfo(
				sta.Code(),
				"update user error",
				"user",
				fmt.Sprintf("uid %s", uid.String()),
				"",
				sta.Message(),
			)
			if e != nil {
				logger.Error(e.Error())
			}
			return &mgmtPB.UpdateUserAdminResponse{}, st.Err()
		}
	}

	pbUserUpdated, err := datamodel.DBUser2PBUser(dbUserUpdated)
	if err != nil {
		logger.Error(err.Error())
		st, e := sterr.CreateErrorResourceInfo(
			codes.Internal,
			"get user error",
			"user",
			fmt.Sprintf("uid %s", dbUserUpdated.UID),
			"",
			err.Error(),
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return &mgmtPB.UpdateUserAdminResponse{}, st.Err()
	}
	resp := mgmtPB.UpdateUserAdminResponse{
		User: pbUserUpdated,
	}

	return &resp, nil
}

// DeleteUserAdmin deletes a user. This endpoint is not supported yet.
func (h *PrivateHandler) DeleteUserAdmin(ctx context.Context, req *mgmtPB.DeleteUserAdminRequest) (*mgmtPB.DeleteUserAdminResponse, error) {
	logger, _ := logger.GetZapLogger(config.Config.Server.Debug)

	st, err := sterr.CreateErrorResourceInfo(
		codes.Unimplemented,
		"delete user not implemented error",
		"endpoint",
		"/users/{user}",
		"",
		"not implemented",
	)
	if err != nil {
		logger.Error(err.Error())
	}
	return &mgmtPB.DeleteUserAdminResponse{}, st.Err()
}
