package shared

import (
	mgmtPB "github.com/instill-ai/protogen-go/vdp/mgmt/v1alpha"
)

// Public API

// HandlerPrivateRPCServer is the RPC server that HandlerRPC talks to, conforming to
// the requirements of net/rpc
type HandlerPrivateRPCServer struct {
	// This is the real implementation
	Impl mgmtPB.MgmtAdminServiceServer
}

// ListUser is the implementation for plugin server
func (h *HandlerPrivateRPCServer) ListUser(reqW *RequestWrapper, respW *ResponseWrapper) error {
	resp, err := h.Impl.ListUser(reqW.MetadataContext, reqW.RequestMessage.(*mgmtPB.ListUserRequest))
	passMetadataContext(reqW, respW)
	return serverGlue(serverWrapResponse(respW, resp), resp, err)
}

// CreateUser is the implementation for plugin server
func (h *HandlerPrivateRPCServer) CreateUser(reqW *RequestWrapper, respW *ResponseWrapper) error {
	resp, err := h.Impl.CreateUser(reqW.MetadataContext, reqW.RequestMessage.(*mgmtPB.CreateUserRequest))
	passMetadataContext(reqW, respW)
	return serverGlue(serverWrapResponse(respW, resp), resp, err)
}

// GetUser is the implementation for plugin server
func (h *HandlerPrivateRPCServer) GetUser(reqW *RequestWrapper, respW *ResponseWrapper) error {
	resp, err := h.Impl.GetUser(reqW.MetadataContext, reqW.RequestMessage.(*mgmtPB.GetUserRequest))
	passMetadataContext(reqW, respW)
	return serverGlue(serverWrapResponse(respW, resp), resp, err)
}

// UpdateUser is the implementation for plugin server
func (h *HandlerPrivateRPCServer) UpdateUser(reqW *RequestWrapper, respW *ResponseWrapper) error {
	resp, err := h.Impl.UpdateUser(reqW.MetadataContext, reqW.RequestMessage.(*mgmtPB.UpdateUserRequest))
	passMetadataContext(reqW, respW)
	return serverGlue(serverWrapResponse(respW, resp), resp, err)
}

// DeleteUser is the implementation for plugin server
func (h *HandlerPrivateRPCServer) DeleteUser(reqW *RequestWrapper, respW *ResponseWrapper) error {
	resp, err := h.Impl.DeleteUser(reqW.MetadataContext, reqW.RequestMessage.(*mgmtPB.DeleteUserRequest))
	passMetadataContext(reqW, respW)
	err = serverGlue(serverWrapResponse(respW, resp), resp, err)
	return err
}

// LookUpUser is the implementation for plugin server
func (h *HandlerPrivateRPCServer) LookUpUser(reqW *RequestWrapper, respW *ResponseWrapper) error {
	resp, err := h.Impl.LookUpUser(reqW.MetadataContext, reqW.RequestMessage.(*mgmtPB.LookUpUserRequest))
	passMetadataContext(reqW, respW)
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
	resp, err := h.Impl.Liveness(reqW.MetadataContext, reqW.RequestMessage.(*mgmtPB.LivenessRequest))
	passMetadataContext(reqW, respW)
	return serverGlue(serverWrapResponse(respW, resp), resp, err)
}

// Readiness is method interface for plugin server
func (h *HandlerPublicRPCServer) Readiness(reqW *RequestWrapper, respW *ResponseWrapper) error {
	resp, err := h.Impl.Readiness(reqW.MetadataContext, reqW.RequestMessage.(*mgmtPB.ReadinessRequest))
	passMetadataContext(reqW, respW)
	return serverGlue(serverWrapResponse(respW, resp), resp, err)
}

// GetAuthenticatedUser is method interface for plugin server
func (h *HandlerPublicRPCServer) GetAuthenticatedUser(reqW *RequestWrapper, respW *ResponseWrapper) error {
	resp, err := h.Impl.GetAuthenticatedUser(reqW.MetadataContext, &mgmtPB.GetAuthenticatedUserRequest{})
	passMetadataContext(reqW, respW)
	return serverGlue(serverWrapResponse(respW, resp), resp, err)
}

// UpdateAuthenticatedUser is method interface for plugin server
func (h *HandlerPublicRPCServer) UpdateAuthenticatedUser(reqW *RequestWrapper, respW *ResponseWrapper) error {
	resp, err := h.Impl.UpdateAuthenticatedUser(reqW.MetadataContext, reqW.RequestMessage.(*mgmtPB.UpdateAuthenticatedUserRequest))
	passMetadataContext(reqW, respW)
	return serverGlue(serverWrapResponse(respW, resp), resp, err)
}

// ExistUsername is method interface for plugin server
func (h *HandlerPublicRPCServer) ExistUsername(reqW *RequestWrapper, respW *ResponseWrapper) error {
	resp, err := h.Impl.ExistUsername(reqW.MetadataContext, reqW.RequestMessage.(*mgmtPB.ExistUsernameRequest))
	passMetadataContext(reqW, respW)
	return serverGlue(serverWrapResponse(respW, resp), resp, err)
}
