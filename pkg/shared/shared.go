package shared

import (
	"context"
	"encoding/gob"
	"fmt"
	"net/rpc"
	"reflect"
	"strings"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/runtime/protoiface"

	"github.com/instill-ai/mgmt-backend/pkg/datamodel"
	mgmtPB "github.com/instill-ai/protogen-go/vdp/mgmt/v1alpha"
)

func init() {

	gob.Register(new(interface{}))

	// Google error details
	gob.Register(&errdetails.ErrorInfo{})
	gob.Register(&errdetails.RetryInfo{})
	gob.Register(&errdetails.DebugInfo{})
	gob.Register(&errdetails.QuotaFailure{})
	gob.Register(&errdetails.PreconditionFailure{})
	gob.Register(&errdetails.BadRequest{})
	gob.Register(&errdetails.RequestInfo{})
	gob.Register(&errdetails.ResourceInfo{})
	gob.Register(&errdetails.Help{})
	gob.Register(&errdetails.LocalizedMessage{})
	gob.Register(&errdetails.QuotaFailure_Violation{})
	gob.Register(&errdetails.PreconditionFailure_Violation{})
	gob.Register(&errdetails.BadRequest_FieldViolation{})
	gob.Register(&errdetails.Help_Link{})

	// private api
	gob.Register(&mgmtPB.ListUserRequest{})
	gob.Register(&mgmtPB.ListUserResponse{})

	gob.Register(&mgmtPB.CreateUserRequest{})
	gob.Register(&mgmtPB.CreateUserResponse{})

	gob.Register(&mgmtPB.GetUserRequest{})
	gob.Register(&mgmtPB.GetUserResponse{})

	gob.Register(&mgmtPB.UpdateUserRequest{})
	gob.Register(&mgmtPB.UpdateUserResponse{})

	gob.Register(&mgmtPB.DeleteUserRequest{})
	gob.Register(&mgmtPB.DeleteUserResponse{})

	gob.Register(&mgmtPB.LookUpUserRequest{})
	gob.Register(&mgmtPB.LookUpUserResponse{})

	// public api
	gob.Register(&mgmtPB.LivenessResponse{})
	gob.Register(&mgmtPB.LivenessRequest{})

	gob.Register(&mgmtPB.ReadinessResponse{})
	gob.Register(&mgmtPB.ReadinessRequest{})

	gob.Register(&mgmtPB.GetAuthenticatedUserResponse{})
	gob.Register(&mgmtPB.GetAuthenticatedUserRequest{})

	gob.Register(&mgmtPB.UpdateAuthenticatedUserResponse{})
	gob.Register(&mgmtPB.UpdateAuthenticatedUserRequest{})

	gob.Register(&mgmtPB.ExistUsernameResponse{})
	gob.Register(&mgmtPB.ExistUsernameRequest{})

}

// HandlerPrivatePlugin is the implementation of plugin.Plugin so we can serve/consume this
type HandlerPrivatePlugin struct {
	// Impl Injection
	Impl mgmtPB.MgmtAdminServiceServer
}

// Server is to implement the plugin.Plugin interface method
func (h *HandlerPrivatePlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &HandlerPrivateRPCServer{Impl: h.Impl}, nil
}

// Client is to implement the plugin.Plugin interface method
func (HandlerPrivatePlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &HandlerPrivateRPC{
		client: c,
	}, nil
}

// HandlerPublicPlugin is the implementation of plugin.Plugin so we can serve/consume this
type HandlerPublicPlugin struct {
	// Impl Injection
	Impl mgmtPB.MgmtPublicServiceServer
}

// Server is to implement the plugin.Plugin interface method
func (h *HandlerPublicPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &HandlerPublicRPCServer{Impl: h.Impl}, nil
}

// Client is to implement the plugin.Plugin interface method
func (HandlerPublicPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &HandlerPublicRPC{
		client: c,
	}, nil
}

// MetadataContext is to be shared with plugin handler to pass custom HTTP code using grpc/metadata
// in the server side. This is a workaround for that we can't directly pass context through gob.
type MetadataContext struct {
	context.Context
	MD *metadata.MD
}

// Handshake is a common handshake that is shared by plugin and host.
var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "PLUGIN",
	MagicCookieValue: "handler",
}

// RequestWrapper wrap a request protobuf message
type RequestWrapper struct {
	MetadataContext *MetadataContext
	RequestMessage  interface{}
}

// ResponseWrapper wrap a response protobuf message with its status returned by gRPC
type ResponseWrapper struct {
	MetadataContext *MetadataContext
	ResponseMessage interface{}
	StatusCode      codes.Code
	StatusMessage   string
	StatusDetails   []interface{}
}

// GetMetadataFromContext gets metadata from the asserted MetadataContext type from the generic context.Context interface
func GetMetadataFromContext(ctx context.Context, key string) []string {
	c, ok := ctx.(*MetadataContext)
	if !ok {
		return nil
	}

	if v, ok := (*c.MD)[key]; ok {
		vals := make([]string, len(v))
		copy(vals, v)
		return vals
	}
	for k, v := range *c.MD {
		// We need to manually convert all keys to lower case, because MD is a
		// map, and there's no guarantee that the MD attached to the context is
		// created using our helper functions.
		if strings.ToLower(k) == key {
			vals := make([]string, len(v))
			copy(vals, v)
			return vals
		}
	}
	return nil
}

// SetMetadataFromContext sets metadata into the asserted MetadataContext type from the generic context.Context interface
func SetMetadataFromContext(ctx context.Context, kv ...string) error {
	c, ok := ctx.(*MetadataContext)
	if !ok {
		return fmt.Errorf("cannot perform type assertion for ctx")
	}
	(*c.MD) = metadata.Pairs(kv...)
	return nil
}

func passMetadataContext(reqW *RequestWrapper, respW *ResponseWrapper) {
	respW.MetadataContext = &MetadataContext{
		MD: &metadata.MD{},
	}
	for k, v := range *reqW.MetadataContext.MD {
		s := make([]string, len(v))
		copy(s, v)
		(*respW.MetadataContext.MD)[k] = s
	}
}

// this is used in only plugin client to return corresponding request message with or without exported fields
func clientWrapRequest(ctx context.Context, req interface{}) *RequestWrapper {
	hasExportedFields := false
	for _, f := range reflect.VisibleFields(reflect.TypeOf(req).Elem()) {
		if f.IsExported() {
			hasExportedFields = true
			break
		}
	}
	c := &MetadataContext{
		MD: &metadata.MD{},
	}
	md, _ := metadata.FromIncomingContext(ctx)
	for k, v := range md {
		s := make([]string, len(v))
		copy(s, v)
		(*c.MD)[k] = s
	}
	if hasExportedFields {
		return &RequestWrapper{
			MetadataContext: c,
			RequestMessage:  req,
		}
	}
	return &RequestWrapper{
		MetadataContext: c,
		RequestMessage:  new(interface{}),
	}
}

func clientGlue(ctx context.Context, respW *ResponseWrapper, resp interface{}) error {
	// for unexported-field struct response
	if reflect.TypeOf(respW.ResponseMessage) != reflect.TypeOf(new(interface{})) {
		proto.Merge(resp.(proto.Message), respW.ResponseMessage.(proto.Message))
		if err := setUserStructZeroValue(resp); err != nil {
			return err
		}
	}
	if respW.StatusCode != codes.OK {
		s := status.New(respW.StatusCode, respW.StatusMessage)
		details := []protoiface.MessageV1{}
		for _, d := range respW.StatusDetails {
			details = append(details, d.(protoiface.MessageV1))
		}
		s, err := s.WithDetails(details...)
		if err != nil {
			return err
		}
		return s.Err()
	}
	if err := grpc.SetHeader(ctx, *respW.MetadataContext.MD); err != nil {
		return err
	}
	return nil
}

// this is used in only plugin server to return corresponding response message with or without exported fields
func serverWrapResponse(respW *ResponseWrapper, resp interface{}) *ResponseWrapper {
	hasExportedFields := false
	for _, f := range reflect.VisibleFields(reflect.TypeOf(resp).Elem()) {
		if f.IsExported() {
			hasExportedFields = true
			break
		}
	}
	if hasExportedFields {
		respW.ResponseMessage = reflect.New(reflect.ValueOf(resp).Elem().Type()).Interface()
	} else {
		respW.ResponseMessage = new(interface{})
	}

	return respW
}

func serverGlue(respW *ResponseWrapper, resp interface{}, err error) error {
	// for unexported-field struct response
	if reflect.TypeOf(respW.ResponseMessage) != reflect.TypeOf(new(interface{})) {
		proto.Merge(respW.ResponseMessage.(proto.Message), resp.(proto.Message))
	}
	s, ok := status.FromError(err)
	if ok {
		respW.StatusCode = s.Code()
		respW.StatusMessage = s.Message()
		respW.StatusDetails = append(respW.StatusDetails, s.Details()...)
	} else {
		respW.StatusCode = codes.Unknown
		respW.StatusMessage = err.Error()
	}
	return nil
}

// To bring zero-value back after rpc data copy (mgmt-backend-specific)
func setUserStructZeroValue(resp interface{}) error {
	userFound := false
	r := reflect.ValueOf(resp).Elem()
	for i := 0; i < r.NumField(); i++ {
		if r.Type().Field(i).Name == "User" {
			userFound = true
			break
		}
	}
	if userFound && !reflect.ValueOf(resp).Elem().FieldByName("User").IsNil() {
		dbUser, err := datamodel.PBUser2DBUser(reflect.ValueOf(resp).Elem().FieldByName("User").Interface().(*mgmtPB.User))
		if err != nil {
			return err
		}
		pbUser, err := datamodel.DBUser2PBUser(dbUser)
		if err != nil {
			return err
		}
		proto.Merge(reflect.ValueOf(resp).Elem().FieldByName("User").Interface().(*mgmtPB.User), pbUser)
	}
	return nil
}
