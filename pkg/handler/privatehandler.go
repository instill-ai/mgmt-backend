package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofrs/uuid"
	"go.einride.tech/aip/filtering"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/instill-ai/mgmt-backend/pkg/service"

	mgmtpb "github.com/instill-ai/protogen-go/mgmt/v1beta"
	checkfield "github.com/instill-ai/x/checkfield"
	errorsx "github.com/instill-ai/x/errors"
)

const defaultPageSize = int32(10)
const maxPageSize = int32(100)

// parseUserIDFromName parses a user resource name of format "users/{user_id}"
// and returns the user_id
func parseUserIDFromName(name string) (string, error) {
	parts := strings.Split(name, "/")
	if len(parts) != 2 || parts[0] != "users" {
		return "", fmt.Errorf("invalid user name format, expected users/{user_id}")
	}
	return parts[1], nil
}

// parseUIDFromPermalink parses a permalink of format "users/{uid}" and returns the UUID
func parseUIDFromPermalink(permalink string) (uuid.UUID, error) {
	parts := strings.Split(permalink, "/")
	if len(parts) != 2 || parts[0] != "users" {
		return uuid.Nil, fmt.Errorf("invalid permalink format, expected users/{uid}")
	}
	return uuid.FromString(parts[1])
}

// parseTokenIDFromName parses a token resource name of format "users/{user_id}/tokens/{token_id}"
// and returns the token_id
func parseTokenIDFromName(name string) (string, error) {
	parts := strings.Split(name, "/")
	// Support both formats:
	// - "tokens/{token_id}" (from proto route /v1beta/{name=tokens/*})
	// - "users/{user_id}/tokens/{token_id}" (legacy full resource name)
	if len(parts) == 2 && parts[0] == "tokens" {
		return parts[1], nil
	}
	if len(parts) == 4 && parts[0] == "users" && parts[2] == "tokens" {
		return parts[3], nil
	}
	return "", fmt.Errorf("invalid token name format, expected tokens/{token_id} or users/{user_id}/tokens/{token_id}")
}

// PrivateHandler is the handler for private endpoints.
// NOTE: Organization admin endpoints are EE-only and implemented in mgmt-backend-ee.
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
	// Parse user ID from name (format: users/{user_id})
	userID, err := parseUserIDFromName(req.GetName())
	if err != nil {
		return nil, err
	}

	pbUser, err := h.Service.GetUserAdmin(ctx, userID)
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
	// Parse user UID from permalink (format: users/{user_uid})
	permalink := req.GetPermalink()
	userUID, err := parseUIDFromPermalink(permalink)
	if err != nil {
		// Manually set the custom header to have a StatusBadRequest http response for REST endpoint
		if err := grpc.SetHeader(ctx, metadata.Pairs("x-http-code", strconv.Itoa(http.StatusBadRequest))); err != nil {
			return nil, err
		}
		return &mgmtpb.LookUpUserAdminResponse{}, err
	}

	pbUser, err := h.Service.GetUserByUIDAdmin(ctx, userUID)
	if err != nil {
		return nil, err
	}

	resp := mgmtpb.LookUpUserAdminResponse{
		User: pbUser,
	}
	return &resp, nil
}

// CheckNamespaceByUIDAdmin checks if the namespace is available by UID.
// NOTE: Organization lookup is EE-only. This CE version only checks user namespaces.
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

	// NOTE: Organization lookup is EE-only.
	// In CE, we only check for user namespaces.

	return &mgmtpb.CheckNamespaceByUIDAdminResponse{
		Type: mgmtpb.CheckNamespaceByUIDAdminResponse_NAMESPACE_AVAILABLE,
	}, nil
}

// CheckNamespaceAdmin checks if the namespace is available.
// NOTE: Organization lookup is EE-only. This CE version only checks user namespaces.
func (h *PrivateHandler) CheckNamespaceAdmin(ctx context.Context, req *mgmtpb.CheckNamespaceAdminRequest) (*mgmtpb.CheckNamespaceAdminResponse, error) {

	err := checkfield.CheckResourceID(req.GetId())
	if err != nil {
		return nil, errorsx.ErrResourceID
	}

	user, err := h.Service.GetUserAdmin(ctx, req.GetId())
	if err == nil {
		// Look up user UID separately
		userUID, _ := h.Service.GetUserUIDByID(ctx, req.GetId())
		return &mgmtpb.CheckNamespaceAdminResponse{
			Type: mgmtpb.CheckNamespaceAdminResponse_NAMESPACE_USER,
			Uid:  userUID.String(),
			Owner: &mgmtpb.CheckNamespaceAdminResponse_User{
				User: user,
			},
		}, nil
	}

	// NOTE: Organization lookup is EE-only.
	// In CE, we only check for user namespaces.

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
	}

	return &mgmtpb.CheckNamespaceAdminResponse{
		Type: mgmtpb.CheckNamespaceAdminResponse_NAMESPACE_AVAILABLE,
	}, nil
}
