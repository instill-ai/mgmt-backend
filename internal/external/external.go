package external

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/instill-ai/mgmt-backend/config"
	"github.com/instill-ai/mgmt-backend/internal/logger"

	usagePB "github.com/instill-ai/protogen-go/vdp/usage/v1alpha"
)

// InitUsageServiceClient initializes a UsageServiceClient instance
func InitUsageServiceClient() (usagePB.UsageServiceClient, *grpc.ClientConn) {
	logger, _ := logger.GetZapLogger()

	var clientDialOpts grpc.DialOption
	var usageCreds credentials.TransportCredentials
	var err error
	if config.Config.UsageBackend.HTTPS.Cert != "" && config.Config.UsageBackend.HTTPS.Key != "" {
		usageCreds, err = credentials.NewServerTLSFromFile(config.Config.UsageBackend.HTTPS.Cert, config.Config.UsageBackend.HTTPS.Key)
		if err != nil {
			logger.Fatal(err.Error())
		}
		clientDialOpts = grpc.WithTransportCredentials(usageCreds)
	} else {
		clientDialOpts = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	clientConn, err := grpc.Dial(fmt.Sprintf("%v:%v", config.Config.UsageBackend.Host, config.Config.UsageBackend.Port), clientDialOpts)
	if err != nil {
		logger.Fatal(err.Error())
	}

	return usagePB.NewUsageServiceClient(clientConn), clientConn

}
