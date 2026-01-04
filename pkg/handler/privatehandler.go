package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gofrs/uuid"
	"go.einride.tech/aip/filtering"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/instill-ai/mgmt-backend/pkg/service"

	mgmtpb "github.com/instill-ai/protogen-go/core/mgmt/v1beta"
	checkfield "github.com/instill-ai/x/checkfield"
	errorsx "github.com/instill-ai/x/errors"
)

const defaultPageSize = int32(10)
const maxPageSize = int32(100)

// PrivateHandler is the handler for private endpoints.
type PrivateHandler struct {
	mgmtpb.UnimplementedMgmtPrivateServiceServer
	Service service.Service
}

// NewPrivateHandler initiates an private handler instance
func NewPrivateHandler(s service.Service) mgmtpb.MgmtPrivateServiceServer {
	return &PrivateHandler{
		Service: s,
	}
}

// ListUsersAdmin lists all users
func (h *PrivateHandler) ListUsersAdmin(ctx context.Context, req *mgmtpb.ListUsersAdminRequest) (*mgmtpb.ListUsersAdminResponse, error) {

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

	resp := mgmtpb.ListUsersAdminResponse{
		Users:         pbUsers,
		NextPageToken: nextPageToken,
		TotalSize:     int32(totalSize),
	}
	return &resp, nil
}

// GetUserAdmin gets a user
func (h *PrivateHandler) GetUserAdmin(ctx context.Context, req *mgmtpb.GetUserAdminRequest) (*mgmtpb.GetUserAdminResponse, error) {

	pbUser, err := h.Service.GetUserAdmin(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	resp := mgmtpb.GetUserAdminResponse{
		User: pbUser,
	}
	return &resp, nil
}

// LookUpUserAdmin gets a user by permalink
func (h *PrivateHandler) LookUpUserAdmin(ctx context.Context, req *mgmtpb.LookUpUserAdminRequest) (*mgmtpb.LookUpUserAdminResponse, error) {
	// Validation: `uid` in request is valid
	uid, err := uuid.FromString(req.UserUid)
	if err != nil {
		// Manually set the custom header to have a StatusBadRequest http response for REST endpoint
		if err := grpc.SetHeader(ctx, metadata.Pairs("x-http-code", strconv.Itoa(http.StatusBadRequest))); err != nil {
			return nil, err
		}
		return &mgmtpb.LookUpUserAdminResponse{}, err
	}

	pbUser, err := h.Service.GetUserByUIDAdmin(ctx, uid)
	if err != nil {
		return nil, err
	}

	resp := mgmtpb.LookUpUserAdminResponse{
		User: pbUser,
	}
	return &resp, nil
}

// ListOrganizationsAdmin lists all organizations
func (h *PrivateHandler) ListOrganizationsAdmin(ctx context.Context, req *mgmtpb.ListOrganizationsAdminRequest) (*mgmtpb.ListOrganizationsAdminResponse, error) {

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

	resp := mgmtpb.ListOrganizationsAdminResponse{
		Organizations: pbOrganizations,
		NextPageToken: nextPageToken,
		TotalSize:     int32(totalSize),
	}
	return &resp, nil
}

// GetOrganizationAdmin gets a organization
func (h *PrivateHandler) GetOrganizationAdmin(ctx context.Context, req *mgmtpb.GetOrganizationAdminRequest) (*mgmtpb.GetOrganizationAdminResponse, error) {

	pbOrganization, err := h.Service.GetOrganizationAdmin(ctx, req.OrganizationId)
	if err != nil {
		return nil, err
	}

	resp := mgmtpb.GetOrganizationAdminResponse{
		Organization: pbOrganization,
	}
	return &resp, nil
}

// LookUpOrganizationAdmin gets a organization by permalink
func (h *PrivateHandler) LookUpOrganizationAdmin(ctx context.Context, req *mgmtpb.LookUpOrganizationAdminRequest) (*mgmtpb.LookUpOrganizationAdminResponse, error) {

	// Validation: `uid` in request is valid
	uid, err := uuid.FromString(req.OrganizationUid)
	if err != nil {
		return nil, err
	}

	pbOrganization, err := h.Service.GetOrganizationByUIDAdmin(ctx, uid)
	if err != nil {
		return nil, err
	}

	resp := mgmtpb.LookUpOrganizationAdminResponse{
		Organization: pbOrganization,
	}
	return &resp, nil
}

// CheckNamespaceAdmin checks if the namespace is available.
func (h *PrivateHandler) CheckNamespaceAdmin(ctx context.Context, req *mgmtpb.CheckNamespaceAdminRequest) (*mgmtpb.CheckNamespaceAdminResponse, error) {

	err := checkfield.CheckResourceID(req.GetId())
	if err != nil {
		return nil, errorsx.ErrResourceID
	}

	user, err := h.Service.GetUserAdmin(ctx, req.GetId())
	if err == nil {
		return &mgmtpb.CheckNamespaceAdminResponse{
			Type: mgmtpb.CheckNamespaceAdminResponse_NAMESPACE_USER,
			Uid:  *user.Uid,
			Owner: &mgmtpb.CheckNamespaceAdminResponse_User{
				User: user,
			},
		}, nil
	}
	org, err := h.Service.GetOrganizationAdmin(ctx, req.GetId())
	if err == nil {
		return &mgmtpb.CheckNamespaceAdminResponse{
			Type: mgmtpb.CheckNamespaceAdminResponse_NAMESPACE_ORGANIZATION,
			Uid:  org.Uid,
			Owner: &mgmtpb.CheckNamespaceAdminResponse_Organization{
				Organization: org,
			},
		}, nil
	}

	// Check for sanitized collision: internally, namespace IDs are normalized
	// by converting "-" to "_", so "foo-bar" and "foo_bar" would collide.
	// If a variant exists (e.g., "foo_bar" for "foo-bar"), check if it's taken.
	variant := getSanitizedNamespaceVariant(req.GetId())
	if variant != "" {
		_, err = h.Service.GetUserAdmin(ctx, variant)
		if err == nil {
			// Variant exists as user - this would cause a collision
			return &mgmtpb.CheckNamespaceAdminResponse{
				Type: mgmtpb.CheckNamespaceAdminResponse_NAMESPACE_RESERVED,
			}, nil
		}
		_, err = h.Service.GetOrganizationAdmin(ctx, variant)
		if err == nil {
			// Variant exists as organization - this would cause a collision
			return &mgmtpb.CheckNamespaceAdminResponse{
				Type: mgmtpb.CheckNamespaceAdminResponse_NAMESPACE_RESERVED,
			}, nil
		}
	}

	return &mgmtpb.CheckNamespaceAdminResponse{
		Type: mgmtpb.CheckNamespaceAdminResponse_NAMESPACE_AVAILABLE,
	}, nil
}

// CheckNamespaceByUIDAdmin checks if the namespace is available by UID.
func (h *PrivateHandler) CheckNamespaceByUIDAdmin(ctx context.Context, req *mgmtpb.CheckNamespaceByUIDAdminRequest) (*mgmtpb.CheckNamespaceByUIDAdminResponse, error) {

	user, err := h.Service.GetUserByUIDAdmin(ctx, uuid.FromStringOrNil(req.GetUid()))
	if err == nil {
		return &mgmtpb.CheckNamespaceByUIDAdminResponse{
			Type: mgmtpb.CheckNamespaceByUIDAdminResponse_NAMESPACE_USER,
			Id:   user.Id,
			Owner: &mgmtpb.CheckNamespaceByUIDAdminResponse_User{
				User: user,
			},
		}, nil
	}
	org, err := h.Service.GetOrganizationByUIDAdmin(ctx, uuid.FromStringOrNil(req.GetUid()))
	if err == nil {
		return &mgmtpb.CheckNamespaceByUIDAdminResponse{
			Type: mgmtpb.CheckNamespaceByUIDAdminResponse_NAMESPACE_ORGANIZATION,
			Id:   org.Id,
			Owner: &mgmtpb.CheckNamespaceByUIDAdminResponse_Organization{
				Organization: org,
			},
		}, nil
	}

	return &mgmtpb.CheckNamespaceByUIDAdminResponse{
		Type: mgmtpb.CheckNamespaceByUIDAdminResponse_NAMESPACE_AVAILABLE,
	}, nil
}
