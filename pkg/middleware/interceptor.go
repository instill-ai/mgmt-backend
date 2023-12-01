package middleware

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/instill-ai/mgmt-backend/pkg/acl"
	"github.com/instill-ai/mgmt-backend/pkg/handler"
	"github.com/instill-ai/mgmt-backend/pkg/repository"
	"github.com/instill-ai/mgmt-backend/pkg/service"
)

// RecoveryInterceptor - panic handler
func RecoveryInterceptorOpt() grpc_recovery.Option {
	return grpc_recovery.WithRecoveryHandler(func(p interface{}) (err error) {
		return status.Errorf(codes.Unknown, "panic triggered: %v", p)
	})
}

// CustomInterceptor - append metadatas for unary
func UnaryAppendMetadataInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "can not extract metadata")
	}

	newCtx := metadata.NewIncomingContext(ctx, md)
	h, err := handler(newCtx, req)

	return h, InjectErrCode(err)
}

// CustomInterceptor - append metadatas for stream
func StreamAppendMetadataInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return status.Error(codes.Internal, "can not extract metadata")
	}

	newCtx := metadata.NewIncomingContext(stream.Context(), md)
	wrapped := grpc_middleware.WrapServerStream(stream)
	wrapped.WrappedContext = newCtx

	err := handler(srv, wrapped)

	return err
}

func InjectErrCode(err error) error {
	if err == nil {
		return nil
	}

	switch {

	case
		errors.Is(err, gorm.ErrDuplicatedKey):
		return status.Error(codes.AlreadyExists, err.Error())
	case
		errors.Is(err, gorm.ErrRecordNotFound):
		return status.Error(codes.NotFound, err.Error())

	case
		errors.Is(err, repository.ErrNoDataDeleted):
		return status.Error(codes.NotFound, err.Error())

	case
		errors.Is(err, repository.ErrOwnerTypeNotMatch),
		errors.Is(err, repository.ErrPageTokenDecode):
		return status.Error(codes.InvalidArgument, err.Error())

	case
		errors.Is(err, service.ErrCanNotRemoveOwnerFromOrganization),
		errors.Is(err, service.ErrCanNotSetAnotherOwner),
		errors.Is(err, service.ErrInvalidRole),
		errors.Is(err, service.ErrInvalidTokenTTL),
		errors.Is(err, service.ErrStateCanOnlyBeActive),
		errors.Is(err, service.ErrPasswordNotMatch):
		return status.Error(codes.InvalidArgument, err.Error())

	case
		errors.Is(err, service.ErrNoPermission):
		return status.Error(codes.PermissionDenied, err.Error())

	case
		errors.Is(err, service.ErrUnauthenticated):
		return status.Error(codes.Unauthenticated, err.Error())

	case
		errors.Is(err, acl.ErrMembershipNotFound):
		return status.Error(codes.NotFound, err.Error())

	case
		errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
		return status.Error(codes.InvalidArgument, err.Error())

	case
		errors.Is(err, handler.ErrCheckUpdateImmutableFields),
		errors.Is(err, handler.ErrCheckOutputOnlyFields),
		errors.Is(err, handler.ErrCheckRequiredFields),
		errors.Is(err, handler.ErrFieldMask),
		errors.Is(err, handler.ErrResourceID),
		errors.Is(err, handler.ErrUpdateMask):
		return status.Error(codes.InvalidArgument, err.Error())

	default:
		return err
	}
}
