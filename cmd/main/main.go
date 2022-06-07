package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"

	"github.com/instill-ai/mgmt-backend/config"
	"github.com/instill-ai/mgmt-backend/internal/external"
	"github.com/instill-ai/mgmt-backend/internal/logger"
	"github.com/instill-ai/mgmt-backend/pkg/handler"
	"github.com/instill-ai/mgmt-backend/pkg/repository"
	"github.com/instill-ai/mgmt-backend/pkg/service"
	"github.com/instill-ai/mgmt-backend/pkg/usage"
	"github.com/instill-ai/x/repo"

	database "github.com/instill-ai/mgmt-backend/internal/db"
	mgmtPB "github.com/instill-ai/protogen-go/vdp/mgmt/v1alpha"
	usagePB "github.com/instill-ai/protogen-go/vdp/usage/v1alpha"
	usageclient "github.com/instill-ai/usage-client/client"
)

func grpcHandlerFunc(grpcServer *grpc.Server, gwHandler http.Handler, CORSOrigins []string) http.Handler {
	return h2c.NewHandler(
		cors.New(cors.Options{
			AllowedOrigins:   CORSOrigins,
			AllowCredentials: true,
			AllowedMethods:   []string{http.MethodHead, http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete},
			Debug:            false,
		}).Handler(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
					grpcServer.ServeHTTP(w, r)
				} else {
					gwHandler.ServeHTTP(w, r)
				}
			})),
		&http2.Server{})
}

func startReporter(ctx context.Context, usageServiceClient usagePB.UsageServiceClient, repository repository.Repository) {
	if config.Config.Server.DisableUsage {
		return
	}

	logger, _ := logger.GetZapLogger()

	version, err := repo.ReadReleaseManifest("release-please/manifest.json")
	if err != nil {
		logger.Fatal(err.Error())
	}

	go func() {
		time.Sleep(5 * time.Second)

		usg := usage.NewUsage(repository)
		err = usageclient.StartReporter(ctx, usageServiceClient, usagePB.Session_SERVICE_MGMT, config.Config.Server.Edition, version, usg.RetrieveUsageData)
		if err != nil {
			logger.Error(fmt.Sprintf("unable to start reporter: %v\n", err))
		}
	}()
}

func main() {

	logger, _ := logger.GetZapLogger()
	defer func() {
		// can't handle the error due to https://github.com/uber-go/zap/issues/880
		_ = logger.Sync()
	}()
	grpc_zap.ReplaceGrpcLoggerV2(logger)

	if err := config.Init(); err != nil {
		logger.Fatal(err.Error())
	}

	db := database.GetConnection()
	defer database.Close(db)

	// Create tls based credential
	var creds credentials.TransportCredentials
	var err error
	if config.Config.Server.HTTPS.Cert != "" && config.Config.Server.HTTPS.Key != "" {
		creds, err = credentials.NewServerTLSFromFile(config.Config.Server.HTTPS.Cert, config.Config.Server.HTTPS.Key)
		if err != nil {
			logger.Fatal(fmt.Sprintf("failed to create credentials: %v", err))
		}
	}

	// Shared options for the logger, with a custom gRPC code to log level functions.
	opts := []grpc_zap.Option{
		grpc_zap.WithDecider(func(fullMethodName string, err error) bool {
			// will not log gRPC calls if it was a call to liveness or readiness and no error was raised
			if err == nil {
				if match, _ := regexp.MatchString("instill.connector.v1alpha.MgmtService/.*ness$", fullMethodName); match {
					return false
				}
			}
			// by default everything will be logged
			return true
		}),
	}

	grpcServerOpts := []grpc.ServerOption{
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_zap.StreamServerInterceptor(logger, opts...),
			grpc_recovery.StreamServerInterceptor(recoveryInterceptorOpt()),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_zap.UnaryServerInterceptor(logger, opts...),
			grpc_recovery.UnaryServerInterceptor(recoveryInterceptorOpt()),
		)),
	}
	if config.Config.Server.HTTPS.Cert != "" && config.Config.Server.HTTPS.Key != "" {
		grpcServerOpts = append(grpcServerOpts, grpc.Creds(creds))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	usageServiceClient, usageServiceClientConn := external.InitUsageServiceClient()
	defer usageServiceClientConn.Close()

	repository := repository.NewRepository(db)

	grpcS := grpc.NewServer(grpcServerOpts...)
	mgmtPB.RegisterUserServiceServer(grpcS, handler.NewHandler(service.NewService(repository)))

	gwS := runtime.NewServeMux(
		runtime.WithForwardResponseOption(httpResponseModifier),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames:   true,
				EmitUnpopulated: true,
				UseEnumNumbers:  false,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
	)

	// Usage collection
	startReporter(ctx, usageServiceClient, repository)

	// Start gRPC server
	var dialOpts []grpc.DialOption
	if config.Config.Server.HTTPS.Cert != "" && config.Config.Server.HTTPS.Key != "" {
		dialOpts = []grpc.DialOption{grpc.WithTransportCredentials(creds)}
	} else {
		dialOpts = []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	}

	if err := mgmtPB.RegisterUserServiceHandlerFromEndpoint(ctx, gwS, fmt.Sprintf(":%v", config.Config.Server.Port), dialOpts); err != nil {
		logger.Fatal(err.Error())
	}

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%v", config.Config.Server.Port),
		Handler: grpcHandlerFunc(grpcS, gwS, config.Config.Server.CORSOrigins),
	}

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 5 seconds.
	quitSig := make(chan os.Signal, 1)
	errSig := make(chan error)
	if config.Config.Server.HTTPS.Cert != "" && config.Config.Server.HTTPS.Key != "" {
		go func() {
			if err := httpServer.ListenAndServeTLS(config.Config.Server.HTTPS.Cert, config.Config.Server.HTTPS.Key); err != nil {
				errSig <- err
			}
		}()
	} else {
		go func() {
			if err := httpServer.ListenAndServe(); err != nil {
				errSig <- err
			}
		}()
	}
	logger.Info("gRPC server is running.")

	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quitSig, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errSig:
		logger.Error(fmt.Sprintf("Fatal error: %v\n", err))
	case <-quitSig:
		logger.Info("Shutting down server...")
		grpcS.GracefulStop()
	}
}
