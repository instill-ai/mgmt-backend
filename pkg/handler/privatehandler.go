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

// ListOrganizationsAdmin lists all organizations
func (h *PrivateHandler) ListOrganizationsAdmin(ctx context.Context, req *mgmtPB.ListOrganizationsAdminRequest) (*mgmtPB.ListOrganizationsAdminResponse, error) {
	logger, _ := logger.GetZapLogger(ctx)

	pageSize := req.GetPageSize()
	if pageSize == 0 {
		pageSize = defaultPageSize
	} else if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	pbOrganizations, totalSize, nextPageToken, err := h.Service.ListOrganizationsAdmin(ctx, int(pageSize), req.GetPageToken(), filtering.Filter{})
	if err != nil {
		sta := status.Convert(err)
		switch sta.Code() {
		case codes.InvalidArgument:
			st, e := sterr.CreateErrorBadRequest(
				"list organization error", []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "ListOrganizationsAdminRequest.page_token",
						Description: sta.Message(),
					},
				})
			if e != nil {
				logger.Error(e.Error())
			}
			return &mgmtPB.ListOrganizationsAdminResponse{}, st.Err()
		default:
			st, e := sterr.CreateErrorResourceInfo(
				sta.Code(),
				"list organization error",
				"organization",
				"",
				"",
				sta.Message(),
			)
			if e != nil {
				logger.Error(e.Error())
			}
			return &mgmtPB.ListOrganizationsAdminResponse{}, st.Err()
		}
	}

	resp := mgmtPB.ListOrganizationsAdminResponse{
		Organizations: pbOrganizations,
		NextPageToken: nextPageToken,
		TotalSize:     int32(totalSize),
	}
	return &resp, nil
}

// GetOrganizationAdmin gets a organization
func (h *PrivateHandler) GetOrganizationAdmin(ctx context.Context, req *mgmtPB.GetOrganizationAdminRequest) (*mgmtPB.GetOrganizationAdminResponse, error) {
	logger, _ := logger.GetZapLogger(ctx)

	id := strings.TrimPrefix(req.GetName(), "organizations/")

	pbOrganization, err := h.Service.GetOrganizationAdmin(ctx, id)
	if err != nil {
		sta := status.Convert(err)
		switch sta.Code() {
		case codes.InvalidArgument:
			st, e := sterr.CreateErrorBadRequest(
				"get organization error", []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "GetOrganizationAdminRequest.name",
						Description: sta.Message(),
					},
				})
			if e != nil {
				logger.Error(e.Error())
			}
			return &mgmtPB.GetOrganizationAdminResponse{}, st.Err()
		default:
			st, e := sterr.CreateErrorResourceInfo(
				sta.Code(),
				"get organization error",
				"organization",
				fmt.Sprintf("id %s", id),
				"",
				sta.Message(),
			)
			if e != nil {
				logger.Error(e.Error())
			}
			return &mgmtPB.GetOrganizationAdminResponse{}, st.Err()
		}
	}

	resp := mgmtPB.GetOrganizationAdminResponse{
		Organization: pbOrganization,
	}
	return &resp, nil
}

// LookUpOrganizationAdmin gets a organization by permalink
func (h *PrivateHandler) LookUpOrganizationAdmin(ctx context.Context, req *mgmtPB.LookUpOrganizationAdminRequest) (*mgmtPB.LookUpOrganizationAdminResponse, error) {
	logger, _ := logger.GetZapLogger(ctx)

	uidStr := strings.TrimPrefix(req.GetPermalink(), "organizations/")
	// Validation: `uid` in request is valid
	uid, err := uuid.FromString(uidStr)
	if err != nil {
		st, e := sterr.CreateErrorBadRequest(
			"look up organization invalid uuid error", []*errdetails.BadRequest_FieldViolation{
				{
					Field:       "LookUpOrganizationAdminRequest.permalink",
					Description: err.Error(),
				},
			},
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return &mgmtPB.LookUpOrganizationAdminResponse{}, st.Err()
	}

	pbOrganization, err := h.Service.GetOrganizationByUIDAdmin(ctx, uid)
	if err != nil {
		sta := status.Convert(err)
		switch sta.Code() {
		case codes.InvalidArgument:
			st, e := sterr.CreateErrorBadRequest(
				"look up organization error", []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "LookUpOrganizationAdminRequest.permalink",
						Description: sta.Message(),
					},
				})
			if e != nil {
				logger.Error(e.Error())
			}
			return &mgmtPB.LookUpOrganizationAdminResponse{}, st.Err()
		default:
			st, e := sterr.CreateErrorResourceInfo(
				sta.Code(),
				"look up organization error",
				"organization",
				fmt.Sprintf("uid %s", uid),
				"",
				sta.Message(),
			)
			if e != nil {
				logger.Error(e.Error())
			}
			return &mgmtPB.LookUpOrganizationAdminResponse{}, st.Err()
		}
	}

	resp := mgmtPB.LookUpOrganizationAdminResponse{
		Organization: pbOrganization,
	}
	return &resp, nil
}
