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

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/instill-ai/mgmt-backend/configs"
	"github.com/instill-ai/mgmt-backend/internal/logger"
	"github.com/instill-ai/mgmt-backend/pkg/handler"
	"github.com/instill-ai/mgmt-backend/pkg/repository"
	"github.com/instill-ai/mgmt-backend/pkg/service"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"

	database "github.com/instill-ai/mgmt-backend/internal/db"
	mgmtPB "github.com/instill-ai/protogen-go/mgmt/v1alpha"
)

func grpcHandlerFunc(grpcServer *grpc.Server, gwHandler http.Handler) http.Handler {
	return h2c.NewHandler(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
				grpcServer.ServeHTTP(w, r)
			} else {
				gwHandler.ServeHTTP(w, r)
			}
		}),
		&http2.Server{})
}

func main() {

	logger, _ := logger.GetZapLogger()
	defer func() {
		// can't handle the error due to https://github.com/uber-go/zap/issues/880
		_ = logger.Sync()
	}()
	grpc_zap.ReplaceGrpcLoggerV2(logger)

	if err := configs.Init(); err != nil {
		logger.Fatal(err.Error())
	}

	db := database.GetConnection()
	defer database.Close(db)

	// Create tls based credential
	var creds credentials.TransportCredentials
	var err error
	if configs.Config.Server.HTTPS.Cert != "" && configs.Config.Server.HTTPS.Key != "" {
		creds, err = credentials.NewServerTLSFromFile(configs.Config.Server.HTTPS.Cert, configs.Config.Server.HTTPS.Key)
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
	if configs.Config.Server.HTTPS.Cert != "" && configs.Config.Server.HTTPS.Key != "" {
		grpcServerOpts = append(grpcServerOpts, grpc.Creds(creds))
	}

	grpcS := grpc.NewServer(grpcServerOpts...)
	mgmtPB.RegisterUserServiceServer(
		grpcS, handler.NewHandler(
			service.NewService(
				repository.NewRepository(db))))

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

	var dialOpts []grpc.DialOption
	if configs.Config.Server.HTTPS.Cert != "" && configs.Config.Server.HTTPS.Key != "" {
		dialOpts = []grpc.DialOption{grpc.WithTransportCredentials(creds)}
	} else {
		dialOpts = []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := mgmtPB.RegisterUserServiceHandlerFromEndpoint(ctx, gwS, fmt.Sprintf(":%v", configs.Config.Server.Port), dialOpts); err != nil {
		logger.Fatal(err.Error())
	}

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%v", configs.Config.Server.Port),
		Handler: grpcHandlerFunc(grpcS, gwS),
	}

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 5 seconds.
	quitSig := make(chan os.Signal, 1)
	errSig := make(chan error)
	if configs.Config.Server.HTTPS.Cert != "" && configs.Config.Server.HTTPS.Key != "" {
		go func() {
			if err := httpServer.ListenAndServeTLS(configs.Config.Server.HTTPS.Cert, configs.Config.Server.HTTPS.Key); err != nil {
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
