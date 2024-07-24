package service

import (
	"context"

	"google.golang.org/grpc/metadata"
)

func InjectOwnerToContext(ctx context.Context, uid string) context.Context {
	ctx = metadata.AppendToOutgoingContext(ctx, "instill-auth-type", "user")
	ctx = metadata.AppendToOutgoingContext(ctx, "instill-user-uid", uid)
	return ctx
}
