package usage

import (
	"context"
	"crypto/tls"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/instill-ai/mgmt-backend/config"
	"github.com/instill-ai/mgmt-backend/pkg/constant"

	usagepb "github.com/instill-ai/protogen-go/core/usage/v1beta"
	logx "github.com/instill-ai/x/log"
)

// InitUsageServiceClient initializes a UsageServiceClient instance.
func InitUsageServiceClient(ctx context.Context, serverConfig *config.ServerConfig) (usagepb.UsageServiceClient, *grpc.ClientConn) {
	logger, _ := logx.GetZapLogger(ctx)

	var clientDialOpts grpc.DialOption
	if serverConfig.Usage.TLSEnabled {
		tlsConfig := &tls.Config{}
		clientDialOpts = grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))
	} else {
		clientDialOpts = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	clientConn, err := grpc.NewClient(fmt.Sprintf("%v:%v", serverConfig.Usage.Host, serverConfig.Usage.Port), clientDialOpts, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(constant.MaxPayloadSize), grpc.MaxCallSendMsgSize(constant.MaxPayloadSize)))
	if err != nil {
		logger.Error(err.Error())
		return nil, nil
	}

	return usagepb.NewUsageServiceClient(clientConn), clientConn
}
