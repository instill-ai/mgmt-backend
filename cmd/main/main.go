package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
	"gorm.io/gorm"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"

	"github.com/instill-ai/mgmt-backend/config"
	"github.com/instill-ai/mgmt-backend/pkg/acl"
	"github.com/instill-ai/mgmt-backend/pkg/handler"
	"github.com/instill-ai/mgmt-backend/pkg/middleware"
	"github.com/instill-ai/mgmt-backend/pkg/repository"
	"github.com/instill-ai/mgmt-backend/pkg/service"
	"github.com/instill-ai/mgmt-backend/pkg/usage"

	database "github.com/instill-ai/mgmt-backend/pkg/db"
	mgmtpb "github.com/instill-ai/protogen-go/core/mgmt/v1beta"
	usagepb "github.com/instill-ai/protogen-go/core/usage/v1beta"
	pipelinepb "github.com/instill-ai/protogen-go/pipeline/pipeline/v1beta"
	clientx "github.com/instill-ai/x/client"
	clientgrpcx "github.com/instill-ai/x/client/grpc"
	logx "github.com/instill-ai/x/log"
	openfgax "github.com/instill-ai/x/openfga"
	otelx "github.com/instill-ai/x/otel"
	servergrpcx "github.com/instill-ai/x/server/grpc"
	gatewayx "github.com/instill-ai/x/server/grpc/gateway"
)

var (
	// These variables might be overridden at buildtime.
	serviceName    = "mgmt-backend"
	serviceVersion = "dev"
)

func main() {
	if err := config.Init(config.ParseConfigFlag()); err != nil {
		log.Fatal(err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup all OpenTelemetry components
	cleanup := otelx.SetupWithCleanup(ctx,
		otelx.WithServiceName(serviceName),
		otelx.WithServiceVersion(serviceVersion),
		otelx.WithHost(config.Config.OTELCollector.Host),
		otelx.WithPort(config.Config.OTELCollector.Port),
		otelx.WithCollectorEnable(config.Config.OTELCollector.Enable),
	)
	defer cleanup()

	logx.Debug = config.Config.Server.Debug
	logger, _ := logx.GetZapLogger(ctx)
	defer func() {
		// can't handle the error due to https://github.com/uber-go/zap/issues/880
		_ = logger.Sync()
	}()

	// Set gRPC logging based on debug mode
	if config.Config.Server.Debug {
		grpczap.ReplaceGrpcLoggerV2WithVerbosity(logger, 0) // All logs
	} else {
		grpczap.ReplaceGrpcLoggerV2WithVerbosity(logger, 3) // verbosity 3 will avoid [transport] from emitting
	}

	// Get gRPC server options and credentials
	grpcServerOpts, err := servergrpcx.NewServerOptionsAndCreds(
		servergrpcx.WithServiceName(serviceName),
		servergrpcx.WithServiceVersion(serviceVersion),
		servergrpcx.WithServiceConfig(clientx.HTTPSConfig{
			Cert: config.Config.Server.HTTPS.Cert,
			Key:  config.Config.Server.HTTPS.Key,
		}),
		servergrpcx.WithSetOTELServerHandler(config.Config.OTELCollector.Enable),
	)
	if err != nil {
		logger.Fatal("failed to create gRPC server options and credentials", zap.Error(err))
	}

	privateGrpcS := grpc.NewServer(grpcServerOpts...)
	reflection.Register(privateGrpcS)

	publicGrpcS := grpc.NewServer(grpcServerOpts...)
	reflection.Register(publicGrpcS)

	pipelinePublicServiceClient, redisClient, db, influxDB, closeClients := newClients(ctx, logger)
	defer closeClients()

	// Initialize OpenFGA client using x/openfga package
	fgaClient, err := openfgax.NewClient(openfgax.ClientParams{
		Config: config.Config.OpenFGA,
		Logger: logger,
	})
	if err != nil {
		logger.Fatal("Failed to create OpenFGA client", zap.Error(err))
	}

	fgaData, err := database.GetFGAMigrationData(db)
	if err != nil {
		logger.Fatal("Failed to get FGA migration data", zap.Error(err))
	}

	logger.Info("Using stored FGA data",
		zap.String("store_id", fgaData.StoreID),
		zap.String("authorization_model_id", fgaData.AuthorizationModelID))

	err = fgaClient.SetStoreID(fgaData.StoreID)
	if err != nil {
		logger.Fatal("Failed to set FGA store ID", zap.Error(err))
	}
	err = fgaClient.SetAuthorizationModelID(fgaData.AuthorizationModelID)
	if err != nil {
		logger.Fatal("Failed to set FGA authorization model ID", zap.Error(err))
	}

	aclClient := acl.NewFGAClient(fgaClient)

	repository := repository.NewRepository(db, redisClient)
	service := service.NewService(
		pipelinePublicServiceClient,
		repository,
		redisClient,
		influxDB,
		aclClient,
		config.Config.Server.InstillCoreHost,
	)

	// Start usage reporter
	var usg usage.Usage
	if config.Config.Server.Usage.Enabled {
		usageServiceClient, usageServiceClientClose, err := clientgrpcx.NewClient[usagepb.UsageServiceClient](
			clientgrpcx.WithServiceConfig(clientx.ServiceConfig{
				Host:       config.Config.Server.Usage.Host,
				PublicPort: config.Config.Server.Usage.Port,
			}),
			clientgrpcx.WithSetOTELClientHandler(config.Config.OTELCollector.Enable),
		)
		if err != nil {
			logger.Error("failed to create usage service client", zap.Error(err))
		}
		defer func() {
			if err := usageServiceClientClose(); err != nil {
				logger.Error("failed to close usage service client", zap.Error(err))
			}
		}()
		logger.Info("try to start usage reporter")
		go func() {
			for {
				usg = usage.NewUsage(ctx, service, usageServiceClient, serviceVersion)
				if usg != nil {
					usg.StartReporter(ctx)
					logger.Info("usage reporter started")
					break
				}
				logger.Warn("retry to start usage reporter after 5 minutes")
				time.Sleep(5 * time.Minute)
			}
		}()
	}

	mgmtpb.RegisterMgmtPrivateServiceServer(
		privateGrpcS,
		handler.NewPrivateHandler(service),
	)
	mgmtpb.RegisterMgmtPublicServiceServer(
		publicGrpcS,
		handler.NewPublicHandler(service, usg, config.Config.Server.Usage.Enabled),
	)

	publicServeMux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(gatewayx.CustomHeaderMatcher),
		runtime.WithForwardResponseOption(gatewayx.HTTPResponseModifier),
		runtime.WithErrorHandler(gatewayx.ErrorHandler),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				EmitUnpopulated: true,
				UseEnumNumbers:  false,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
	)
	if err := publicServeMux.HandlePath("GET", "/v1beta/{name=users/*}/avatar", middleware.AppendCustomHeaderMiddleware(publicServeMux, repository, middleware.HandleAvatar)); err != nil {
		logger.Fatal(err.Error())
	}
	if err := publicServeMux.HandlePath("GET", "/v1beta/{name=organizations/*}/avatar", middleware.AppendCustomHeaderMiddleware(publicServeMux, repository, middleware.HandleAvatar)); err != nil {
		logger.Fatal(err.Error())
	}

	dialOpts, err := clientgrpcx.NewClientOptionsAndCreds(
		clientgrpcx.WithServiceConfig(clientx.ServiceConfig{
			HTTPS: clientx.HTTPSConfig{
				Cert: config.Config.Server.HTTPS.Cert,
				Key:  config.Config.Server.HTTPS.Key,
			},
		}),
		clientgrpcx.WithSetOTELClientHandler(false),
	)
	if err != nil {
		logger.Fatal("failed to create client options and credentials", zap.Error(err))
	}

	if err := mgmtpb.RegisterMgmtPublicServiceHandlerFromEndpoint(ctx, publicServeMux, fmt.Sprintf(":%v", config.Config.Server.PublicPort), dialOpts); err != nil {
		logger.Fatal(err.Error())
	}

	publicHTTPServer := &http.Server{
		Addr:    fmt.Sprintf(":%v", config.Config.Server.PublicPort),
		Handler: grpcHandlerFunc(publicGrpcS, publicServeMux),
	}

	quitSig := make(chan os.Signal, 1)
	errSig := make(chan error)

	go func() {
		privatePort := fmt.Sprintf(":%d", config.Config.Server.PrivatePort)
		privateListener, err := net.Listen("tcp", privatePort)
		if err != nil {
			errSig <- fmt.Errorf("failed to listen: %w", err)
		}
		if err := privateGrpcS.Serve(privateListener); err != nil {
			errSig <- fmt.Errorf("failed to serve: %w", err)
		}
	}()

	go func() {
		var err error
		switch {
		case config.Config.Server.HTTPS.Cert != "" && config.Config.Server.HTTPS.Key != "":
			err = publicHTTPServer.ListenAndServeTLS(config.Config.Server.HTTPS.Cert, config.Config.Server.HTTPS.Key)
		default:
			err = publicHTTPServer.ListenAndServe()
		}
		if err != nil {
			errSig <- err
		}
	}()

	logger.Info("gRPC servers are running.")

	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quitSig, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errSig:
		logger.Error(fmt.Sprintf("Fatal error: %v\n", err))
	case <-quitSig:
		// send out the usage report at exit
		if config.Config.Server.Usage.Enabled && usg != nil {
			usg.TriggerSingleReporter(ctx)
		}
		logger.Info("Shutting down server...")
		privateGrpcS.GracefulStop()
		publicGrpcS.GracefulStop()
	}
}

func grpcHandlerFunc(grpcServer *grpc.Server, gwHandler http.Handler) http.Handler {
	return h2c.NewHandler(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
				grpcServer.ServeHTTP(w, r)
			} else {
				gwHandler.ServeHTTP(w, r)
			}
		}),
		&http2.Server{},
	)
}

func newClients(ctx context.Context, logger *zap.Logger) (pipelinepb.PipelinePublicServiceClient, *redis.Client, *gorm.DB, repository.InfluxDB, func()) {
	closeFuncs := map[string]func() error{}

	pipelinePublicServiceClient, pipelinePublicClose, err := clientgrpcx.NewClient[pipelinepb.PipelinePublicServiceClient](
		clientgrpcx.WithServiceConfig(clientx.ServiceConfig{
			Host:       config.Config.PipelineBackend.Host,
			PublicPort: config.Config.PipelineBackend.PublicPort,
		}),
		clientgrpcx.WithSetOTELClientHandler(config.Config.OTELCollector.Enable),
	)
	if err != nil {
		logger.Fatal("failed to create pipeline public service client", zap.Error(err))
	}
	closeFuncs["pipelinePublic"] = pipelinePublicClose

	db := database.GetConnection(&config.Config.Database)
	closeFuncs["database"] = func() error {
		database.Close(db)
		return nil
	}

	redisClient := redis.NewClient(&config.Config.Cache.Redis.RedisOptions)
	closeFuncs["redis"] = redisClient.Close

	influxDB := repository.MustNewInfluxDB(ctx, config.Config)
	closeFuncs["influxDB"] = func() error {
		influxDB.Close()
		return nil
	}

	closer := func() {
		for conn, fn := range closeFuncs {
			if err := fn(); err != nil {
				logger.Error("Failed to close conn", zap.Error(err), zap.String("conn", conn))
			}
		}
	}

	return pipelinePublicServiceClient, redisClient, db, influxDB, closer
}
