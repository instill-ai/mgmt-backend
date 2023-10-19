package external

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	influxdb3 "github.com/InfluxCommunity/influxdb3-go/influx"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"

	"github.com/instill-ai/mgmt-backend/config"
	"github.com/instill-ai/mgmt-backend/pkg/logger"

	usagePB "github.com/instill-ai/protogen-go/core/usage/v1alpha"
	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1alpha"
)

// InitConnectorPublicServiceClient initialises a ConnectorPublicServiceClient instance
func InitConnectorPublicServiceClient(ctx context.Context, appConfig *config.AppConfig) (connectorPB.ConnectorPublicServiceClient, *grpc.ClientConn) {
	logger, _ := logger.GetZapLogger(ctx)

	var clientDialOpts grpc.DialOption
	if appConfig.ConnectorBackend.HTTPS.Cert != "" && appConfig.ConnectorBackend.HTTPS.Key != "" {
		creds, err := credentials.NewServerTLSFromFile(appConfig.ConnectorBackend.HTTPS.Cert, appConfig.ConnectorBackend.HTTPS.Key)
		if err != nil {
			logger.Fatal(err.Error())
		}
		clientDialOpts = grpc.WithTransportCredentials(creds)
	} else {
		clientDialOpts = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	clientConn, err := grpc.Dial(fmt.Sprintf("%v:%v", appConfig.ConnectorBackend.Host, appConfig.ConnectorBackend.PublicPort), clientDialOpts)
	if err != nil {
		logger.Error(err.Error())
		return nil, nil
	}

	return connectorPB.NewConnectorPublicServiceClient(clientConn), clientConn
}

// InitPipelinePublicServiceClient initialises a PipelinePublicServiceClient instance
func InitPipelinePublicServiceClient(ctx context.Context, appConfig *config.AppConfig) (pipelinePB.PipelinePublicServiceClient, *grpc.ClientConn) {
	logger, _ := logger.GetZapLogger(ctx)

	var clientDialOpts grpc.DialOption
	if appConfig.PipelineBackend.HTTPS.Cert != "" && appConfig.PipelineBackend.HTTPS.Key != "" {
		creds, err := credentials.NewServerTLSFromFile(appConfig.PipelineBackend.HTTPS.Cert, appConfig.PipelineBackend.HTTPS.Key)
		if err != nil {
			logger.Fatal(err.Error())
		}
		clientDialOpts = grpc.WithTransportCredentials(creds)
	} else {
		clientDialOpts = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	clientConn, err := grpc.Dial(fmt.Sprintf("%v:%v", appConfig.PipelineBackend.Host, appConfig.PipelineBackend.PublicPort), clientDialOpts)
	if err != nil {
		logger.Error(err.Error())
		return nil, nil
	}

	return pipelinePB.NewPipelinePublicServiceClient(clientConn), clientConn
}

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

// InitInfluxDBServiceClientV2 initialises a InfluxDBServiceClientV2 instance
func InitInfluxDBServiceClientV2(ctx context.Context, appConfig *config.AppConfig) (influxdb2.Client, api.QueryAPI) {

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

// InitInfluxDBServiceClientV3 initialises a InfluxDBServiceClientV3 instance
func InitInfluxDBServiceClientV3(ctx context.Context, appConfig *config.AppConfig) *influxdb3.Client {

	logger, _ := logger.GetZapLogger(ctx)

	client, err := influxdb3.New(influxdb3.Configs{
		HostURL:   appConfig.InfluxDB.URL,
		AuthToken: appConfig.InfluxDB.Token,
	})

	if err != nil {
		logger.Error(err.Error())
	}

	return client
}
