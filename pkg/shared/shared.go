package shared

import (
	"encoding/gob"
	"fmt"
	"net/rpc"
	"reflect"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/runtime/protoiface"

	"github.com/instill-ai/mgmt-backend/pkg/datamodel"
	"github.com/instill-ai/x/sterr"

	mgmtPB "github.com/instill-ai/protogen-go/vdp/mgmt/v1alpha"
)

// HandlerAdminPlugin is the implementation of plugin.Plugin so we can serve/consume this
type HandlerAdminPlugin struct {
	// Impl Injection
	Impl mgmtPB.MgmtAdminServiceServer
}

// Server is to implement the plugin.Plugin interface method
func (h *HandlerAdminPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	sterr.RegisterMessageTypes()
	registerMessageTypes()
	return &HandlerAdminRPCServer{Impl: h.Impl}, nil
}

// Client is to implement the plugin.Plugin interface method
func (HandlerAdminPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	sterr.RegisterMessageTypes()
	registerMessageTypes()
	return &HandlerAdminRPC{
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
	sterr.RegisterMessageTypes()
	registerMessageTypes()
	return &HandlerPublicRPCServer{Impl: h.Impl}, nil
}

// Client is to implement the plugin.Plugin interface method
func (HandlerPublicPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	sterr.RegisterMessageTypes()
	registerMessageTypes()
	return &HandlerPublicRPC{
		client: c,
	}, nil
}

// Handshake is a common handshake that is shared by plugin and host.
var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "PLUGIN",
	MagicCookieValue: "handler",
}

// RequestWrapper wrap a request protobuf message
type RequestWrapper struct {
	RequestMessage interface{}
}

// ResponseWrapper wrap a response protobuf message with its status returned by gRPC
type ResponseWrapper struct {
	ResponseMessage interface{}
	StatusCode      codes.Code
	StatusMessage   string
	StatusDetails   []interface{}
}

// registerMessageTypes registers common and custom types in gob
func registerMessageTypes() {

	gob.Register(new(interface{}))

	// admin api
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

// this is used in only plugin client
func clientWrapRequest(req interface{}) *RequestWrapper {
	hasExportedFields := false
	for _, f := range reflect.VisibleFields(reflect.TypeOf(req).Elem()) {
		if f.IsExported() {
			hasExportedFields = true
			break
		}
	}
	if hasExportedFields {
		return &RequestWrapper{
			RequestMessage: req,
		}
	}
	return &RequestWrapper{
		RequestMessage: new(interface{}),
	}
}

func clientGlue(respW *ResponseWrapper, resp interface{}) error {
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
	return nil
}

// To bring zero-value back after rpc data copy
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

// this is used in only plugin server
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
	if !ok {
		return fmt.Errorf("unable to convert error to status")
	}
	respW.StatusCode = s.Code()
	respW.StatusMessage = s.Message()
	respW.StatusDetails = append(respW.StatusDetails, s.Details()...)
	return nil
}
