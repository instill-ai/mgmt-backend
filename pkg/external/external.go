package external

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/instill-ai/mgmt-backend/config"
	"github.com/instill-ai/mgmt-backend/pkg/logger"

	usagePB "github.com/instill-ai/protogen-go/base/usage/v1alpha"
)

// InitUsageServiceClient initializes a UsageServiceClient instance
func InitUsageServiceClient(ctx context.Context, serverConfig *config.ServerConfig) (usagePB.UsageServiceClient, *grpc.ClientConn) {
	logger, _ := logger.GetZapLogger(ctx)

	var clientDialOpts grpc.DialOption
	if serverConfig.Usage.TLSEnabled {
		tlsConfig := &tls.Config{}
		clientDialOpts = grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))
	} else {
		clientDialOpts = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	clientConn, err := grpc.Dial(fmt.Sprintf("%v:%v", serverConfig.Usage.Host, serverConfig.Usage.Port), clientDialOpts)
	if err != nil {
		logger.Error(err.Error())
		return nil, nil
	}

	return usagePB.NewUsageServiceClient(clientConn), clientConn
}

// InitInfluxDBServiceClient initialises a InfluxDBServiceClient instance
func InitInfluxDBServiceClient(ctx context.Context, appConfig *config.AppConfig) (influxdb2.Client, api.QueryAPI) {

	logger, _ := logger.GetZapLogger(ctx)

	var creds credentials.TransportCredentials
	var err error

	influxOptions := influxdb2.DefaultOptions()
	if appConfig.Server.Debug {
		influxOptions = influxOptions.SetLogLevel(log.DebugLevel)
	}
	influxOptions = influxOptions.SetFlushInterval(uint(time.Duration(appConfig.InfluxDB.FlushInterval * int(time.Second)).Milliseconds()))

	if appConfig.InfluxDB.HTTPS.Cert != "" && appConfig.InfluxDB.HTTPS.Key != "" {
		// TODO: support TLS
		creds, err = credentials.NewServerTLSFromFile(appConfig.InfluxDB.HTTPS.Cert, appConfig.InfluxDB.HTTPS.Key)
		if err != nil {
			logger.Fatal(err.Error())
		}
		logger.Info(creds.Info().ServerName)
	}

	client := influxdb2.NewClientWithOptions(
		appConfig.InfluxDB.URL,
		appConfig.InfluxDB.Token,
		influxOptions,
	)

	if _, err := client.Ping(ctx); err != nil {
		logger.Warn(err.Error())
	}

	queryAPI := client.QueryAPI(appConfig.InfluxDB.Org)

	return client, queryAPI
}
