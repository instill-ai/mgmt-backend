package shared

import (
	mgmtPB "github.com/instill-ai/protogen-go/vdp/mgmt/v1alpha"
)

// Private API

// HandlerPrivateRPCServer is the RPC server that HandlerRPC talks to, conforming to
// the requirements of net/rpc
type HandlerPrivateRPCServer struct {
	// This is the real implementation
	Impl mgmtPB.MgmtPrivateServiceServer
}

// ListUsersAdmin is the implementation for plugin server
func (h *HandlerPrivateRPCServer) ListUsersAdmin(reqW *RequestWrapper, respW *ResponseWrapper) error {
	resp, err := h.Impl.ListUsersAdmin(reqW.MetadataContext, reqW.RequestMessage.(*mgmtPB.ListUsersAdminRequest))
	passMetadataContext(reqW, respW)
	return serverGlue(serverWrapResponse(respW, resp), resp, err)
}

// CreateUserAdmin is the implementation for plugin server
func (h *HandlerPrivateRPCServer) CreateUserAdmin(reqW *RequestWrapper, respW *ResponseWrapper) error {
	resp, err := h.Impl.CreateUserAdmin(reqW.MetadataContext, reqW.RequestMessage.(*mgmtPB.CreateUserAdminRequest))
	passMetadataContext(reqW, respW)
	return serverGlue(serverWrapResponse(respW, resp), resp, err)
}

// GetUserAdmin is the implementation for plugin server
func (h *HandlerPrivateRPCServer) GetUserAdmin(reqW *RequestWrapper, respW *ResponseWrapper) error {
	resp, err := h.Impl.GetUserAdmin(reqW.MetadataContext, reqW.RequestMessage.(*mgmtPB.GetUserAdminRequest))
	passMetadataContext(reqW, respW)
	return serverGlue(serverWrapResponse(respW, resp), resp, err)
}

// UpdateUserAdmin is the implementation for plugin server
func (h *HandlerPrivateRPCServer) UpdateUserAdmin(reqW *RequestWrapper, respW *ResponseWrapper) error {
	resp, err := h.Impl.UpdateUserAdmin(reqW.MetadataContext, reqW.RequestMessage.(*mgmtPB.UpdateUserAdminRequest))
	passMetadataContext(reqW, respW)
	return serverGlue(serverWrapResponse(respW, resp), resp, err)
}

// DeleteUserAdmin is the implementation for plugin server
func (h *HandlerPrivateRPCServer) DeleteUserAdmin(reqW *RequestWrapper, respW *ResponseWrapper) error {
	resp, err := h.Impl.DeleteUserAdmin(reqW.MetadataContext, reqW.RequestMessage.(*mgmtPB.DeleteUserAdminRequest))
	passMetadataContext(reqW, respW)
	err = serverGlue(serverWrapResponse(respW, resp), resp, err)
	return err
}

// LookUpUserAdmin is the implementation for plugin server
func (h *HandlerPrivateRPCServer) LookUpUserAdmin(reqW *RequestWrapper, respW *ResponseWrapper) error {
	resp, err := h.Impl.LookUpUserAdmin(reqW.MetadataContext, reqW.RequestMessage.(*mgmtPB.LookUpUserAdminRequest))
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

// QueryAuthenticatedUser is method interface for plugin server
func (h *HandlerPublicRPCServer) QueryAuthenticatedUser(reqW *RequestWrapper, respW *ResponseWrapper) error {
	resp, err := h.Impl.QueryAuthenticatedUser(reqW.MetadataContext, &mgmtPB.QueryAuthenticatedUserRequest{})
	passMetadataContext(reqW, respW)
	return serverGlue(serverWrapResponse(respW, resp), resp, err)
}

// PatchAuthenticatedUser is method interface for plugin server
func (h *HandlerPublicRPCServer) PatchAuthenticatedUser(reqW *RequestWrapper, respW *ResponseWrapper) error {
	resp, err := h.Impl.PatchAuthenticatedUser(reqW.MetadataContext, reqW.RequestMessage.(*mgmtPB.PatchAuthenticatedUserRequest))
	passMetadataContext(reqW, respW)
	return serverGlue(serverWrapResponse(respW, resp), resp, err)
}

// ExistUsername is method interface for plugin server
func (h *HandlerPublicRPCServer) ExistUsername(reqW *RequestWrapper, respW *ResponseWrapper) error {
	resp, err := h.Impl.ExistUsername(reqW.MetadataContext, reqW.RequestMessage.(*mgmtPB.ExistUsernameRequest))
	passMetadataContext(reqW, respW)
	return serverGlue(serverWrapResponse(respW, resp), resp, err)
}
