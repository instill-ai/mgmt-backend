package external

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
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
	if config.Config.UsageBackend.TLSEnabled {
		roots, err := x509.SystemCertPool()
		if err != nil {
			logger.Fatal(err.Error())
		}

		tlsConfig := tls.Config{
			RootCAs:            roots,
			InsecureSkipVerify: true,
			NextProtos:         []string{"h2"},
		}
		clientDialOpts = grpc.WithTransportCredentials(credentials.NewTLS(&tlsConfig))
	} else {
		clientDialOpts = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	clientConn, err := grpc.Dial(
		fmt.Sprintf("%v:%v", config.Config.UsageBackend.Host, config.Config.UsageBackend.Port),
		clientDialOpts,
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.Config{
				BaseDelay:  500 * time.Millisecond,
				Multiplier: 1.5,
				Jitter:     0.2,
				MaxDelay:   19 * time.Second,
			},
			MinConnectTimeout: 5 * time.Second,
		}),
	)

	if err != nil {
		logger.Fatal(err.Error())
	}

	return usagePB.NewUsageServiceClient(clientConn), clientConn

}
