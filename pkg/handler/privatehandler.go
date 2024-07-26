package handler

import (
	"context"

	"github.com/gofrs/uuid"
	"go.einride.tech/aip/filtering"
	"google.golang.org/genproto/googleapis/rpc/errdetails"

	"github.com/instill-ai/mgmt-backend/pkg/logger"
	"github.com/instill-ai/mgmt-backend/pkg/service"
	"github.com/instill-ai/x/sterr"

	mgmtPB "github.com/instill-ai/protogen-go/core/mgmt/v1beta"
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

	pageSize := req.GetPageSize()
	if pageSize == 0 {
		pageSize = defaultPageSize
	} else if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	pbUsers, totalSize, nextPageToken, err := h.Service.ListUsersAdmin(ctx, int(pageSize), req.GetPageToken(), filtering.Filter{})
	if err != nil {
		return nil, err
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

	pbUser, err := h.Service.GetUserAdmin(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	resp := mgmtPB.GetUserAdminResponse{
		User: pbUser,
	}
	return &resp, nil
}

// LookUpUserAdmin gets a user by permalink
func (h *PrivateHandler) LookUpUserAdmin(ctx context.Context, req *mgmtPB.LookUpUserAdminRequest) (*mgmtPB.LookUpUserAdminResponse, error) {
	logger, _ := logger.GetZapLogger(ctx)

	// Validation: `uid` in request is valid
	uid, err := uuid.FromString(req.UserUid)
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
		return nil, err
	}

	resp := mgmtPB.LookUpUserAdminResponse{
		User: pbUser,
	}
	return &resp, nil
}

// ListOrganizationsAdmin lists all organizations
func (h *PrivateHandler) ListOrganizationsAdmin(ctx context.Context, req *mgmtPB.ListOrganizationsAdminRequest) (*mgmtPB.ListOrganizationsAdminResponse, error) {

	pageSize := req.GetPageSize()
	if pageSize == 0 {
		pageSize = defaultPageSize
	} else if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	pbOrganizations, totalSize, nextPageToken, err := h.Service.ListOrganizationsAdmin(ctx, int(pageSize), req.GetPageToken(), filtering.Filter{})
	if err != nil {
		return nil, err
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

	pbOrganization, err := h.Service.GetOrganizationAdmin(ctx, req.OrganizationId)
	if err != nil {
		return nil, err
	}

	resp := mgmtPB.GetOrganizationAdminResponse{
		Organization: pbOrganization,
	}
	return &resp, nil
}

// LookUpOrganizationAdmin gets a organization by permalink
func (h *PrivateHandler) LookUpOrganizationAdmin(ctx context.Context, req *mgmtPB.LookUpOrganizationAdminRequest) (*mgmtPB.LookUpOrganizationAdminResponse, error) {

	// Validation: `uid` in request is valid
	uid, err := uuid.FromString(req.OrganizationUid)
	if err != nil {
		return nil, err
	}

	pbOrganization, err := h.Service.GetOrganizationByUIDAdmin(ctx, uid)
	if err != nil {
		return nil, err
	}

	resp := mgmtPB.LookUpOrganizationAdminResponse{
		Organization: pbOrganization,
	}
	return &resp, nil
}

func (h *PrivateHandler) CheckNamespaceAdmin(ctx context.Context, req *mgmtPB.CheckNamespaceAdminRequest) (*mgmtPB.CheckNamespaceAdminResponse, error) {

	err := checkfield.CheckResourceID(req.GetId())
	if err != nil {
		return nil, ErrResourceID
	}

	user, err := h.Service.GetUserAdmin(ctx, req.GetId())
	if err == nil {
		return &mgmtPB.CheckNamespaceAdminResponse{
			Type: mgmtPB.CheckNamespaceAdminResponse_NAMESPACE_USER,
			Uid:  *user.Uid,
			Owner: &mgmtPB.CheckNamespaceAdminResponse_User{
				User: user,
			},
		}, nil
	}
	org, err := h.Service.GetOrganizationAdmin(ctx, req.GetId())
	if err == nil {
		return &mgmtPB.CheckNamespaceAdminResponse{
			Type: mgmtPB.CheckNamespaceAdminResponse_NAMESPACE_ORGANIZATION,
			Uid:  org.Uid,
			Owner: &mgmtPB.CheckNamespaceAdminResponse_Organization{
				Organization: org,
			},
		}, nil
	}

	return &mgmtPB.CheckNamespaceAdminResponse{
		Type: mgmtPB.CheckNamespaceAdminResponse_NAMESPACE_AVAILABLE,
	}, nil
}

func (h *PrivateHandler) CheckNamespaceByUIDAdmin(ctx context.Context, req *mgmtPB.CheckNamespaceByUIDAdminRequest) (*mgmtPB.CheckNamespaceByUIDAdminResponse, error) {

	user, err := h.Service.GetUserByUIDAdmin(ctx, uuid.FromStringOrNil(req.GetUid()))
	if err == nil {
		return &mgmtPB.CheckNamespaceByUIDAdminResponse{
			Type: mgmtPB.CheckNamespaceByUIDAdminResponse_NAMESPACE_USER,
			Id:   user.Id,
			Owner: &mgmtPB.CheckNamespaceByUIDAdminResponse_User{
				User: user,
			},
		}, nil
	}
	org, err := h.Service.GetOrganizationByUIDAdmin(ctx, uuid.FromStringOrNil(req.GetUid()))
	if err == nil {
		return &mgmtPB.CheckNamespaceByUIDAdminResponse{
			Type: mgmtPB.CheckNamespaceByUIDAdminResponse_NAMESPACE_ORGANIZATION,
			Id:   org.Id,
			Owner: &mgmtPB.CheckNamespaceByUIDAdminResponse_Organization{
				Organization: org,
			},
		}, nil
	}

	return &mgmtPB.CheckNamespaceByUIDAdminResponse{
		Type: mgmtPB.CheckNamespaceByUIDAdminResponse_NAMESPACE_AVAILABLE,
	}, nil
}
