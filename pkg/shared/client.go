package shared

import (
	"context"
	"net/rpc"

	mgmtPB "github.com/instill-ai/protogen-go/vdp/mgmt/v1alpha"
)

// Admin endpoints

// HandlerAdminRPC is an implementation that talks over RPC
type HandlerAdminRPC struct {
	client *rpc.Client
}

// ListUser is method interface for plugin client
func (h *HandlerAdminRPC) ListUser(ctx context.Context, req *mgmtPB.ListUserRequest) (resp *mgmtPB.ListUserResponse, err error) {
	respW := &ResponseWrapper{}
	if err := h.client.Call("Plugin.ListUser", clientWrapRequest(req), respW); err != nil {
		return &mgmtPB.ListUserResponse{}, err
	}
	resp = &mgmtPB.ListUserResponse{}
	err = clientGlue(respW, resp)
	return resp, err
}

// CreateUser is method interface for plugin client
func (h *HandlerAdminRPC) CreateUser(ctx context.Context, req *mgmtPB.CreateUserRequest) (resp *mgmtPB.CreateUserResponse, err error) {
	respW := &ResponseWrapper{}
	if err := h.client.Call("Plugin.CreateUser", clientWrapRequest(req), respW); err != nil {
		return &mgmtPB.CreateUserResponse{}, err
	}
	resp = &mgmtPB.CreateUserResponse{}
	err = clientGlue(respW, resp)
	return resp, err
}

// GetUser is method interface for plugin client
func (h *HandlerAdminRPC) GetUser(ctx context.Context, req *mgmtPB.GetUserRequest) (resp *mgmtPB.GetUserResponse, err error) {
	respW := &ResponseWrapper{}
	if err := h.client.Call("Plugin.GetUser", clientWrapRequest(req), respW); err != nil {
		return &mgmtPB.GetUserResponse{}, err
	}
	resp = &mgmtPB.GetUserResponse{}
	err = clientGlue(respW, resp)
	return resp, err
}

// UpdateUser is method interface for plugin client
func (h *HandlerAdminRPC) UpdateUser(ctx context.Context, req *mgmtPB.UpdateUserRequest) (resp *mgmtPB.UpdateUserResponse, err error) {
	respW := &ResponseWrapper{}
	if err := h.client.Call("Plugin.UpdateUser", clientWrapRequest(req), respW); err != nil {
		return &mgmtPB.UpdateUserResponse{}, err
	}
	resp = &mgmtPB.UpdateUserResponse{}
	err = clientGlue(respW, resp)
	return resp, err
}

// DeleteUser is method interface for plugin client
func (h *HandlerAdminRPC) DeleteUser(ctx context.Context, req *mgmtPB.DeleteUserRequest) (resp *mgmtPB.DeleteUserResponse, err error) {
	respW := &ResponseWrapper{}
	if err := h.client.Call("Plugin.DeleteUser", clientWrapRequest(req), respW); err != nil {
		return &mgmtPB.DeleteUserResponse{}, err
	}
	resp = &mgmtPB.DeleteUserResponse{}
	err = clientGlue(respW, resp)
	return resp, err
}

// LookUpUser is method interface for plugin client
func (h *HandlerAdminRPC) LookUpUser(ctx context.Context, req *mgmtPB.LookUpUserRequest) (resp *mgmtPB.LookUpUserResponse, err error) {
	respW := &ResponseWrapper{}
	if err := h.client.Call("Plugin.LookUpUser", clientWrapRequest(req), respW); err != nil {
		return &mgmtPB.LookUpUserResponse{}, err
	}
	resp = &mgmtPB.LookUpUserResponse{}
	err = clientGlue(respW, resp)
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
	if err := h.client.Call("Plugin.Liveness", clientWrapRequest(req), respW); err != nil {
		return &mgmtPB.LivenessResponse{}, err
	}
	resp = &mgmtPB.LivenessResponse{}
	err = clientGlue(respW, resp)
	return resp, err
}

// Readiness is method interface for plugin client
func (h *HandlerPublicRPC) Readiness(ctx context.Context, req *mgmtPB.ReadinessRequest) (resp *mgmtPB.ReadinessResponse, err error) {
	respW := &ResponseWrapper{}
	if err := h.client.Call("Plugin.Readiness", clientWrapRequest(req), respW); err != nil {
		return &mgmtPB.ReadinessResponse{}, err
	}
	resp = &mgmtPB.ReadinessResponse{}
	err = clientGlue(respW, resp)
	return resp, err
}

// GetAuthenticatedUser is method interface for plugin client
func (h *HandlerPublicRPC) GetAuthenticatedUser(ctx context.Context, req *mgmtPB.GetAuthenticatedUserRequest) (resp *mgmtPB.GetAuthenticatedUserResponse, err error) {
	respW := &ResponseWrapper{}
	if err := h.client.Call("Plugin.GetAuthenticatedUser", clientWrapRequest(req), respW); err != nil {
		return &mgmtPB.GetAuthenticatedUserResponse{}, err
	}
	resp = &mgmtPB.GetAuthenticatedUserResponse{}
	err = clientGlue(respW, resp)
	return resp, err
}

// UpdateAuthenticatedUser is method interface for plugin client
func (h *HandlerPublicRPC) UpdateAuthenticatedUser(ctx context.Context, req *mgmtPB.UpdateAuthenticatedUserRequest) (resp *mgmtPB.UpdateAuthenticatedUserResponse, err error) {
	respW := &ResponseWrapper{}
	if err = h.client.Call("Plugin.UpdateAuthenticatedUser", clientWrapRequest(req), respW); err != nil {
		return &mgmtPB.UpdateAuthenticatedUserResponse{}, err
	}
	resp = &mgmtPB.UpdateAuthenticatedUserResponse{}
	err = clientGlue(respW, resp)
	return resp, err
}

// ExistUsername is method interface for plugin client
func (h *HandlerPublicRPC) ExistUsername(ctx context.Context, req *mgmtPB.ExistUsernameRequest) (resp *mgmtPB.ExistUsernameResponse, err error) {
	respW := &ResponseWrapper{}
	if err := h.client.Call("Plugin.ExistUsername", clientWrapRequest(req), respW); err != nil {
		return &mgmtPB.ExistUsernameResponse{}, err
	}
	resp = &mgmtPB.ExistUsernameResponse{}
	err = clientGlue(respW, resp)
	return resp, err
}
