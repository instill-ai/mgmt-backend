package shared

import (
	"context"

	mgmtPB "github.com/instill-ai/protogen-go/vdp/mgmt/v1alpha"
)

// Admin API

// HandlerAdminRPCServer is the RPC server that HandlerRPC talks to, conforming to
// the requirements of net/rpc
type HandlerAdminRPCServer struct {
	// This is the real implementation
	Impl mgmtPB.MgmtAdminServiceServer
}

// ListUser is the implementation for plugin server
func (h *HandlerAdminRPCServer) ListUser(reqW *RequestWrapper, respW *ResponseWrapper) error {
	resp, err := h.Impl.ListUser(context.Background(), reqW.RequestMessage.(*mgmtPB.ListUserRequest))
	return serverGlue(serverWrapResponse(respW, resp), resp, err)
}

// CreateUser is the implementation for plugin server
func (h *HandlerAdminRPCServer) CreateUser(reqW *RequestWrapper, respW *ResponseWrapper) error {
	resp, err := h.Impl.CreateUser(context.Background(), reqW.RequestMessage.(*mgmtPB.CreateUserRequest))
	return serverGlue(serverWrapResponse(respW, resp), resp, err)
}

// GetUser is the implementation for plugin server
func (h *HandlerAdminRPCServer) GetUser(reqW *RequestWrapper, respW *ResponseWrapper) error {
	resp, err := h.Impl.GetUser(context.Background(), reqW.RequestMessage.(*mgmtPB.GetUserRequest))
	return serverGlue(serverWrapResponse(respW, resp), resp, err)
}

// UpdateUser is the implementation for plugin server
func (h *HandlerAdminRPCServer) UpdateUser(reqW *RequestWrapper, respW *ResponseWrapper) error {
	resp, err := h.Impl.UpdateUser(context.Background(), reqW.RequestMessage.(*mgmtPB.UpdateUserRequest))
	return serverGlue(serverWrapResponse(respW, resp), resp, err)
}

// DeleteUser is the implementation for plugin server
func (h *HandlerAdminRPCServer) DeleteUser(reqW *RequestWrapper, respW *ResponseWrapper) error {
	resp, err := h.Impl.DeleteUser(context.Background(), reqW.RequestMessage.(*mgmtPB.DeleteUserRequest))

	err = serverGlue(serverWrapResponse(respW, resp), resp, err)
	return err
}

// LookUpUser is the implementation for plugin server
func (h *HandlerAdminRPCServer) LookUpUser(reqW *RequestWrapper, respW *ResponseWrapper) error {
	resp, err := h.Impl.LookUpUser(context.Background(), reqW.RequestMessage.(*mgmtPB.LookUpUserRequest))
	return serverGlue(serverWrapResponse(respW, resp), resp, err)
}

// Public API

// HandlerPublicRPCServer is the RPC server that HandlerRPC talks to, conforming to
// the requirements of net/rpc
type HandlerPublicRPCServer struct {
	// This is the real implementation
	Impl mgmtPB.MgmtPublicServiceServer
}

// Liveness is method interface for plugin server
func (h *HandlerPublicRPCServer) Liveness(reqW *RequestWrapper, respW *ResponseWrapper) error {
	resp, err := h.Impl.Liveness(context.Background(), reqW.RequestMessage.(*mgmtPB.LivenessRequest))
	return serverGlue(serverWrapResponse(respW, resp), resp, err)
}

// Readiness is method interface for plugin server
func (h *HandlerPublicRPCServer) Readiness(reqW *RequestWrapper, respW *ResponseWrapper) error {
	resp, err := h.Impl.Readiness(context.Background(), reqW.RequestMessage.(*mgmtPB.ReadinessRequest))
	return serverGlue(serverWrapResponse(respW, resp), resp, err)
}

// GetAuthenticatedUser is method interface for plugin server
func (h *HandlerPublicRPCServer) GetAuthenticatedUser(reqW *RequestWrapper, respW *ResponseWrapper) error {
	// resp, err := h.Impl.GetAuthenticatedUser(context.Background(), reqW.RequestMessage.(*mgmtPB.GetAuthenticatedUserRequest))
	resp, err := h.Impl.GetAuthenticatedUser(context.Background(), &mgmtPB.GetAuthenticatedUserRequest{})
	return serverGlue(serverWrapResponse(respW, resp), resp, err)
}

// UpdateAuthenticatedUser is method interface for plugin server
func (h *HandlerPublicRPCServer) UpdateAuthenticatedUser(reqW *RequestWrapper, respW *ResponseWrapper) error {
	resp, err := h.Impl.UpdateAuthenticatedUser(context.Background(), reqW.RequestMessage.(*mgmtPB.UpdateAuthenticatedUserRequest))
	return serverGlue(serverWrapResponse(respW, resp), resp, err)
}

// ExistUsername is method interface for plugin server
func (h *HandlerPublicRPCServer) ExistUsername(reqW *RequestWrapper, respW *ResponseWrapper) error {
	resp, err := h.Impl.ExistUsername(context.Background(), reqW.RequestMessage.(*mgmtPB.ExistUsernameRequest))
	return serverGlue(serverWrapResponse(respW, resp), resp, err)
}
