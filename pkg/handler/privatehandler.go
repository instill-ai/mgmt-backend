package handler

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofrs/uuid"
	"go.einride.tech/aip/filtering"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/instill-ai/mgmt-backend/pkg/logger"
	"github.com/instill-ai/mgmt-backend/pkg/service"
	"github.com/instill-ai/x/sterr"

	mgmtPB "github.com/instill-ai/protogen-go/core/mgmt/v1alpha"
	checkfield "github.com/instill-ai/x/checkfield"
)

const defaultPageSize = int32(10)
const maxPageSize = int32(100)

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
	logger, _ := logger.GetZapLogger(ctx)

	pageSize := req.GetPageSize()
	if pageSize == 0 {
		pageSize = defaultPageSize
	} else if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	pbUsers, totalSize, nextPageToken, err := h.Service.ListUsersAdmin(ctx, int(pageSize), req.GetPageToken(), filtering.Filter{})
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

	resp := mgmtPB.ListUsersAdminResponse{
		Users:         pbUsers,
		NextPageToken: nextPageToken,
		TotalSize:     int32(totalSize),
	}
	return &resp, nil
}

// CreateUserAdmin creates a user. This endpoint is not supported yet.
func (h *PrivateHandler) CreateUserAdmin(ctx context.Context, req *mgmtPB.CreateUserAdminRequest) (*mgmtPB.CreateUserAdminResponse, error) {
	logger, _ := logger.GetZapLogger(ctx)
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
	logger, _ := logger.GetZapLogger(ctx)

	id := strings.TrimPrefix(req.GetName(), "users/")

	pbUser, err := h.Service.GetUserAdmin(ctx, id)
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

	resp := mgmtPB.GetUserAdminResponse{
		User: pbUser,
	}
	return &resp, nil
}

// LookUpUserAdmin gets a user by permalink
func (h *PrivateHandler) LookUpUserAdmin(ctx context.Context, req *mgmtPB.LookUpUserAdminRequest) (*mgmtPB.LookUpUserAdminResponse, error) {
	logger, _ := logger.GetZapLogger(ctx)

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

	pbUser, err := h.Service.GetUserByUIDAdmin(ctx, uid)
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

	resp := mgmtPB.LookUpUserAdminResponse{
		User: pbUser,
	}
	return &resp, nil
}
