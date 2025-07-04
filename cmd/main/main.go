package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	openfgaClient "github.com/openfga/go-sdk/client"

	"github.com/instill-ai/mgmt-backend/config"
	"github.com/instill-ai/mgmt-backend/pkg/acl"
	"github.com/instill-ai/mgmt-backend/pkg/constant"
	"github.com/instill-ai/mgmt-backend/pkg/external"
	"github.com/instill-ai/mgmt-backend/pkg/handler"
	"github.com/instill-ai/mgmt-backend/pkg/logger"
	"github.com/instill-ai/mgmt-backend/pkg/middleware"
	"github.com/instill-ai/mgmt-backend/pkg/repository"
	"github.com/instill-ai/mgmt-backend/pkg/service"
	"github.com/instill-ai/mgmt-backend/pkg/usage"

	database "github.com/instill-ai/mgmt-backend/pkg/db"
	custom_otel "github.com/instill-ai/mgmt-backend/pkg/logger/otel"
	mgmtPB "github.com/instill-ai/protogen-go/core/mgmt/v1beta"
)

var (
	// These variables might be overridden at buildtime.
	serviceVersion = "dev"
	// serviceName = "mgmt-backend"

	propagator propagation.TextMapPropagator
)

func grpcHandlerFunc(grpcServer *grpc.Server, gwHandler http.Handler) http.Handler {
	return h2c.NewHandler(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			propagator = b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader))
			ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
			r = r.WithContext(ctx)

			if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
				grpcServer.ServeHTTP(w, r)
			} else {
				gwHandler.ServeHTTP(w, r)
			}
		}),
		&http2.Server{})
}

func main() {
	if err := config.Init(config.ParseConfigFlag()); err != nil {
		log.Fatal(err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())

	if tp, err := custom_otel.SetupTracing(ctx, "mgmt-backend"); err != nil {
		panic(err)
	} else {
		defer func() {
			err = tp.Shutdown(ctx)
		}()
	}

	ctx, span := otel.Tracer("main-tracer").Start(ctx,
		"main",
	)
	defer cancel()

	logger, _ := logger.GetZapLogger(ctx)
	defer func() {
		// can't handle the error due to https://github.com/uber-go/zap/issues/880
		_ = logger.Sync()
	}()

	// verbosity 3 will avoid [transport] from emitting
	grpc_zap.ReplaceGrpcLoggerV2WithVerbosity(logger, 3)

	db := database.GetConnection(&config.Config.Database)
	defer database.Close(db)

	// Shared options for the logger, with a custom gRPC code to log level functions.
	opts := []grpc_zap.Option{
		grpc_zap.WithDecider(func(fullMethodName string, err error) bool {
			// will not log gRPC calls if it was a call to liveness or readiness and no error was raised
			if err == nil {
				// stop logging successful private function calls
				if match, _ := regexp.MatchString("core.mgmt.v1beta.MgmtPrivateService/.*(?:Admin|ness)$", fullMethodName); match {
					return false
				}
				if match, _ := regexp.MatchString("core.mgmt.v1beta.MgmtPublicService/.*ness$", fullMethodName); match {
					return false
				}
			}
			// by default everything will be logged
			return true
		}),
	}

	grpcServerOpts := []grpc.ServerOption{
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			middleware.StreamAppendMetadataInterceptor,
			grpc_zap.StreamServerInterceptor(logger, opts...),
			grpc_recovery.StreamServerInterceptor(middleware.RecoveryInterceptorOpt()),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			middleware.UnaryAppendMetadataInterceptor,
			grpc_zap.UnaryServerInterceptor(logger, opts...),
			grpc_recovery.UnaryServerInterceptor(middleware.RecoveryInterceptorOpt()),
		)),
	}

	// Create tls based credential
	var creds credentials.TransportCredentials
	var tlsConfig *tls.Config
	var err error

	redisClient := redis.NewClient(&config.Config.Cache.Redis.RedisOptions)
	defer redisClient.Close()

	fgaClient, err := openfgaClient.NewSdkClient(&openfgaClient.ClientConfiguration{
		ApiScheme: "http",
		ApiHost:   fmt.Sprintf("%s:%d", config.Config.OpenFGA.Host, config.Config.OpenFGA.Port),
	})

	if err != nil {
		panic(err)
	}

	var fgaReplicaClient *openfgaClient.OpenFgaClient
	if config.Config.OpenFGA.Replica.Host != "" {

		fgaReplicaClient, err = openfgaClient.NewSdkClient(&openfgaClient.ClientConfiguration{
			ApiScheme: "http",
			ApiHost:   fmt.Sprintf("%s:%d", config.Config.OpenFGA.Replica.Host, config.Config.OpenFGA.Replica.Port),
		})
		if err != nil {
			panic(err)
		}
	}

	var aclClient acl.ACLClient
	if stores, err := fgaClient.ListStores(context.Background()).Execute(); err == nil {
		fgaClient.SetStoreId(*(*stores.Stores)[0].Id)
		if fgaReplicaClient != nil {
			fgaReplicaClient.SetStoreId(*(*stores.Stores)[0].Id)
		}
		if models, err := fgaClient.ReadAuthorizationModels(context.Background()).Execute(); err == nil {
			aclClient = acl.NewACLClient(fgaClient, fgaReplicaClient, redisClient, (*models.AuthorizationModels)[0].Id)
		}
		if err != nil {
			panic(err)
		}

	} else {
		panic(err)
	}

	if config.Config.Server.HTTPS.Cert != "" && config.Config.Server.HTTPS.Key != "" {
		tlsConfig = &tls.Config{
			ClientAuth: tls.RequireAndVerifyClientCert,
		}
		creds, err = credentials.NewServerTLSFromFile(config.Config.Server.HTTPS.Cert, config.Config.Server.HTTPS.Key)
		if err != nil {
			logger.Fatal(fmt.Sprintf("failed to create credentials: %v", err))
		}
		grpcServerOpts = append(grpcServerOpts, grpc.Creds(creds))
	}
	grpcServerOpts = append(grpcServerOpts, grpc.MaxRecvMsgSize(constant.MaxPayloadSize))
	grpcServerOpts = append(grpcServerOpts, grpc.MaxSendMsgSize(constant.MaxPayloadSize))

	pipelinePublicServiceClient, pipelinePublicServiceClientConn := external.InitPipelinePublicServiceClient(ctx, &config.Config)
	if pipelinePublicServiceClientConn != nil {
		defer pipelinePublicServiceClientConn.Close()
	}

	influxDB := repository.MustNewInfluxDB(ctx, config.Config)
	defer influxDB.Close()

	repository := repository.NewRepository(db, redisClient)
	service := service.NewService(repository, redisClient, influxDB, pipelinePublicServiceClient, &aclClient, config.Config.Server.InstillCoreHost)

	// Start usage reporter
	var usg usage.Usage
	if config.Config.Server.Usage.Enabled {
		usageServiceClient, usageServiceClientConn := external.InitUsageServiceClient(ctx, &config.Config.Server)
		if usageServiceClientConn != nil {
			defer usageServiceClientConn.Close()
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
	}

	privateGrpcS := grpc.NewServer(grpcServerOpts...)
	reflection.Register(privateGrpcS)

	publicGrpcS := grpc.NewServer(grpcServerOpts...)
	reflection.Register(publicGrpcS)

	mgmtPB.RegisterMgmtPrivateServiceServer(
		privateGrpcS,
		handler.NewPrivateHandler(service),
	)
	mgmtPB.RegisterMgmtPublicServiceServer(
		publicGrpcS,
		handler.NewPublicHandler(service, usg, config.Config.Server.Usage.Enabled),
	)

	publicServeMux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(middleware.CustomMatcher),
		runtime.WithForwardResponseOption(middleware.HTTPResponseModifier),
		runtime.WithErrorHandler(middleware.ErrorHandler),
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

	// Start gRPC server
	var dialOpts []grpc.DialOption
	if config.Config.Server.HTTPS.Cert != "" && config.Config.Server.HTTPS.Key != "" {
		dialOpts = []grpc.DialOption{grpc.WithTransportCredentials(creds), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(constant.MaxPayloadSize), grpc.MaxCallSendMsgSize(constant.MaxPayloadSize))}
	} else {
		dialOpts = []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(constant.MaxPayloadSize), grpc.MaxCallSendMsgSize(constant.MaxPayloadSize))}
	}

	if err := mgmtPB.RegisterMgmtPublicServiceHandlerFromEndpoint(ctx, publicServeMux, fmt.Sprintf(":%v", config.Config.Server.PublicPort), dialOpts); err != nil {
		logger.Fatal(err.Error())
	}

	publicHTTPServer := &http.Server{
		Addr:      fmt.Sprintf(":%v", config.Config.Server.PublicPort),
		Handler:   grpcHandlerFunc(publicGrpcS, publicServeMux),
		TLSConfig: tlsConfig,
	}

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 5 seconds.
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

	span.End()
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
