package shared

import (
	"context"
	"net/rpc"

	mgmtPB "github.com/instill-ai/protogen-go/vdp/mgmt/v1alpha"
)

// Private endpoints

// HandlerPrivateRPC is an implementation that talks over RPC
type HandlerPrivateRPC struct {
	client *rpc.Client
}

// ListUsersAdmin is method interface for plugin client
func (h *HandlerPrivateRPC) ListUsersAdmin(ctx context.Context, req *mgmtPB.ListUsersAdminRequest) (resp *mgmtPB.ListUsersAdminResponse, err error) {
	respW := &ResponseWrapper{}
	if err := h.client.Call("Plugin.ListUsersAdmin", clientWrapRequest(ctx, req), respW); err != nil {
		return &mgmtPB.ListUsersAdminResponse{}, err
	}
	resp = &mgmtPB.ListUsersAdminResponse{}
	err = clientGlue(ctx, respW, resp)
	return resp, err
}

// CreateUserAdmin is method interface for plugin client
func (h *HandlerPrivateRPC) CreateUserAdmin(ctx context.Context, req *mgmtPB.CreateUserAdminRequest) (resp *mgmtPB.CreateUserAdminResponse, err error) {
	respW := &ResponseWrapper{}
	if err := h.client.Call("Plugin.CreateUserAdmin", clientWrapRequest(ctx, req), respW); err != nil {
		return &mgmtPB.CreateUserAdminResponse{}, err
	}
	resp = &mgmtPB.CreateUserAdminResponse{}
	err = clientGlue(ctx, respW, resp)
	return resp, err
}

// GetUserAdmin is method interface for plugin client
func (h *HandlerPrivateRPC) GetUserAdmin(ctx context.Context, req *mgmtPB.GetUserAdminRequest) (resp *mgmtPB.GetUserAdminResponse, err error) {
	respW := &ResponseWrapper{}
	if err := h.client.Call("Plugin.GetUserAdmin", clientWrapRequest(ctx, req), respW); err != nil {
		return &mgmtPB.GetUserAdminResponse{}, err
	}
	resp = &mgmtPB.GetUserAdminResponse{}
	err = clientGlue(ctx, respW, resp)
	return resp, err
}

// UpdateUserAdmin is method interface for plugin client
func (h *HandlerPrivateRPC) UpdateUserAdmin(ctx context.Context, req *mgmtPB.UpdateUserAdminRequest) (resp *mgmtPB.UpdateUserAdminResponse, err error) {
	respW := &ResponseWrapper{}
	if err := h.client.Call("Plugin.UpdateUserAdmin", clientWrapRequest(ctx, req), respW); err != nil {
		return &mgmtPB.UpdateUserAdminResponse{}, err
	}
	resp = &mgmtPB.UpdateUserAdminResponse{}
	err = clientGlue(ctx, respW, resp)
	return resp, err
}

// DeleteUserAdmin is method interface for plugin client
func (h *HandlerPrivateRPC) DeleteUserAdmin(ctx context.Context, req *mgmtPB.DeleteUserAdminRequest) (resp *mgmtPB.DeleteUserAdminResponse, err error) {
	respW := &ResponseWrapper{}
	if err := h.client.Call("Plugin.DeleteUserAdmin", clientWrapRequest(ctx, req), respW); err != nil {
		return &mgmtPB.DeleteUserAdminResponse{}, err
	}
	resp = &mgmtPB.DeleteUserAdminResponse{}
	err = clientGlue(ctx, respW, resp)
	return resp, err
}

// LookUpUserAdmin is method interface for plugin client
func (h *HandlerPrivateRPC) LookUpUserAdmin(ctx context.Context, req *mgmtPB.LookUpUserAdminRequest) (resp *mgmtPB.LookUpUserAdminResponse, err error) {
	respW := &ResponseWrapper{}
	if err := h.client.Call("Plugin.LookUpUserAdmin", clientWrapRequest(ctx, req), respW); err != nil {
		return &mgmtPB.LookUpUserAdminResponse{}, err
	}
	resp = &mgmtPB.LookUpUserAdminResponse{}
	err = clientGlue(ctx, respW, resp)
	return resp, err
}

// Public endpoints

// HandlerPublicRPC is an implementation that talks over RPC
type HandlerPublicRPC struct {
	client *rpc.Client
}

// Liveness is method interface for plugin client
func (h *HandlerPublicRPC) Liveness(ctx context.Context, req *mgmtPB.LivenessRequest) (resp *mgmtPB.LivenessResponse, err error) {
	respW := &ResponseWrapper{}
	if err := h.client.Call("Plugin.Liveness", clientWrapRequest(ctx, req), respW); err != nil {
		return &mgmtPB.LivenessResponse{}, err
	}
	resp = &mgmtPB.LivenessResponse{}
	err = clientGlue(ctx, respW, resp)
	return resp, err
}

// Readiness is method interface for plugin client
func (h *HandlerPublicRPC) Readiness(ctx context.Context, req *mgmtPB.ReadinessRequest) (resp *mgmtPB.ReadinessResponse, err error) {
	respW := &ResponseWrapper{}
	if err := h.client.Call("Plugin.Readiness", clientWrapRequest(ctx, req), respW); err != nil {
		return &mgmtPB.ReadinessResponse{}, err
	}
	resp = &mgmtPB.ReadinessResponse{}
	err = clientGlue(ctx, respW, resp)
	return resp, err
}

// QueryAuthenticatedUser is method interface for plugin client
func (h *HandlerPublicRPC) QueryAuthenticatedUser(ctx context.Context, req *mgmtPB.QueryAuthenticatedUserRequest) (resp *mgmtPB.QueryAuthenticatedUserResponse, err error) {
	respW := &ResponseWrapper{}
	if err := h.client.Call("Plugin.QueryAuthenticatedUser", clientWrapRequest(ctx, req), respW); err != nil {
		return &mgmtPB.QueryAuthenticatedUserResponse{}, err
	}
	resp = &mgmtPB.QueryAuthenticatedUserResponse{}
	err = clientGlue(ctx, respW, resp)
	return resp, err
}

// PatchAuthenticatedUser is method interface for plugin client
func (h *HandlerPublicRPC) PatchAuthenticatedUser(ctx context.Context, req *mgmtPB.PatchAuthenticatedUserRequest) (resp *mgmtPB.PatchAuthenticatedUserResponse, err error) {
	respW := &ResponseWrapper{}
	if err = h.client.Call("Plugin.PatchAuthenticatedUser", clientWrapRequest(ctx, req), respW); err != nil {
		return &mgmtPB.PatchAuthenticatedUserResponse{}, err
	}
	resp = &mgmtPB.PatchAuthenticatedUserResponse{}
	err = clientGlue(ctx, respW, resp)
	return resp, err
}

// ExistUsername is method interface for plugin client
func (h *HandlerPublicRPC) ExistUsername(ctx context.Context, req *mgmtPB.ExistUsernameRequest) (resp *mgmtPB.ExistUsernameResponse, err error) {
	respW := &ResponseWrapper{}
	if err := h.client.Call("Plugin.ExistUsername", clientWrapRequest(ctx, req), respW); err != nil {
		return &mgmtPB.ExistUsernameResponse{}, err
	}
	resp = &mgmtPB.ExistUsernameResponse{}
	err = clientGlue(ctx, respW, resp)
	return resp, err
}
