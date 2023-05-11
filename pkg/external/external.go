package external

import (
	"crypto/tls"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/instill-ai/mgmt-backend/config"
	"github.com/instill-ai/mgmt-backend/pkg/logger"

	usagePB "github.com/instill-ai/protogen-go/vdp/usage/v1alpha"
)

// InitUsageServiceClient initializes a UsageServiceClient instance
func InitUsageServiceClient(usageServerConfig *config.UsageServerConfig, debug bool) (usagePB.UsageServiceClient, *grpc.ClientConn) {
	logger, _ := logger.GetZapLogger(debug)

	var clientDialOpts grpc.DialOption
	if usageServerConfig.TLSEnabled {
		tlsConfig := &tls.Config{}
		clientDialOpts = grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))
	} else {
		clientDialOpts = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	clientConn, err := grpc.Dial(fmt.Sprintf("%v:%v", usageServerConfig.Host, usageServerConfig.Port), clientDialOpts)
	if err != nil {
		logger.Error(err.Error())
		return nil, nil
	}

	return usagePB.NewUsageServiceClient(clientConn), clientConn
}
