package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"regexp"
	"strings"
	"syscall"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hashicorp/go-plugin"
	"github.com/rs/cors"
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
	hclog "github.com/hashicorp/go-hclog"

	"github.com/instill-ai/mgmt-backend/config"
	"github.com/instill-ai/mgmt-backend/internal/shared"
	"github.com/instill-ai/mgmt-backend/pkg/logger"

	database "github.com/instill-ai/mgmt-backend/internal/db"
	mgmtPB "github.com/instill-ai/protogen-go/vdp/mgmt/v1alpha"
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

func main() {
	if err := config.Init(); err != nil {
		log.Fatal(err.Error())
	}

	logger, _ := logger.GetZapLogger()
	defer func() {
		// can't handle the error due to https://github.com/uber-go/zap/issues/880
		_ = logger.Sync()
	}()
	grpc_zap.ReplaceGrpcLoggerV2(logger)

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
				if match, _ := regexp.MatchString("vdp.mgmt.v1alpha.MgmtAdminService/.*ness$", fullMethodName); match {
					return false
				}
				if match, _ := regexp.MatchString("vdp.mgmt.v1alpha.MgmtPublicService/.*ness$", fullMethodName); match {
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
			grpc_recovery.StreamServerInterceptor(RecoveryInterceptorOpt()),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_zap.UnaryServerInterceptor(logger, opts...),
			grpc_recovery.UnaryServerInterceptor(RecoveryInterceptorOpt()),
		)),
	}
	if config.Config.Server.HTTPS.Cert != "" && config.Config.Server.HTTPS.Key != "" {
		grpcServerOpts = append(grpcServerOpts, grpc.Creds(creds))
	}

	ex, err := os.Executable()
	if err != nil {
		logger.Fatal(err.Error())
	}

	// host and launch the plugin process.
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: shared.Handshake,
		Plugins: map[string]plugin.Plugin{
			"admin handler":  &shared.HandlerAdminPlugin{},
			"public handler": &shared.HandlerPublicPlugin{},
		},
		Cmd: exec.Command(fmt.Sprintf("%s/%s", path.Dir(ex), config.Config.Server.Plugin)),
		Logger: hclog.New(&hclog.LoggerOptions{
			Name:   "plugin",
			Output: os.Stdout,
			Level: func() hclog.Level {
				if config.Config.Server.Debug {
					return hclog.Debug
				}
				return hclog.Info
			}(),
		}),
	})
	defer client.Kill()

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		log.Fatal(err)
	}

	// Request the plugin
	rawAdminHandler, err := rpcClient.Dispense("admin handler")
	if err != nil {
		logger.Fatal(err.Error())
	}

	rawPublicHandler, err := rpcClient.Dispense("public handler")
	if err != nil {
		logger.Fatal(err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	adminGrpcS := grpc.NewServer(grpcServerOpts...)
	reflection.Register(adminGrpcS)

	publicGrpcS := grpc.NewServer(grpcServerOpts...)
	reflection.Register(publicGrpcS)

	adminHandler := rawAdminHandler.(mgmtPB.MgmtAdminServiceServer)
	mgmtPB.RegisterMgmtAdminServiceServer(adminGrpcS, adminHandler)

	publicHandler := rawPublicHandler.(mgmtPB.MgmtPublicServiceServer)
	mgmtPB.RegisterMgmtPublicServiceServer(publicGrpcS, publicHandler)

	adminServeMux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(customMatcher),
		runtime.WithForwardResponseOption(httpResponseModifier),
		runtime.WithErrorHandler(errorHandler),
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

	publicServeMux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(customMatcher),
		runtime.WithForwardResponseOption(httpResponseModifier),
		runtime.WithErrorHandler(errorHandler),
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

	// Start gRPC server
	var dialOpts []grpc.DialOption
	if config.Config.Server.HTTPS.Cert != "" && config.Config.Server.HTTPS.Key != "" {
		dialOpts = []grpc.DialOption{grpc.WithTransportCredentials(creds)}
	} else {
		dialOpts = []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	}

	if err := mgmtPB.RegisterMgmtAdminServiceHandlerFromEndpoint(ctx, adminServeMux, fmt.Sprintf(":%v", config.Config.Server.AdminPort), dialOpts); err != nil {
		logger.Fatal(err.Error())
	}

	if err := mgmtPB.RegisterMgmtPublicServiceHandlerFromEndpoint(ctx, publicServeMux, fmt.Sprintf(":%v", config.Config.Server.PublicPort), dialOpts); err != nil {
		logger.Fatal(err.Error())
	}

	adminHTTPServer := &http.Server{
		Addr:    fmt.Sprintf(":%v", config.Config.Server.AdminPort),
		Handler: grpcHandlerFunc(adminGrpcS, adminServeMux, config.Config.Server.CORSOrigins),
	}

	publicHTTPServer := &http.Server{
		Addr:    fmt.Sprintf(":%v", config.Config.Server.PublicPort),
		Handler: grpcHandlerFunc(publicGrpcS, publicServeMux, config.Config.Server.CORSOrigins),
	}

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 5 seconds.
	quitSig := make(chan os.Signal, 1)
	errSig := make(chan error)
	if config.Config.Server.HTTPS.Cert != "" && config.Config.Server.HTTPS.Key != "" {
		go func() {
			if err := adminHTTPServer.ListenAndServeTLS(config.Config.Server.HTTPS.Cert, config.Config.Server.HTTPS.Key); err != nil {
				errSig <- err
			}
		}()
		go func() {
			if err := publicHTTPServer.ListenAndServeTLS(config.Config.Server.HTTPS.Cert, config.Config.Server.HTTPS.Key); err != nil {
				errSig <- err
			}
		}()
	} else {
		go func() {
			if err := adminHTTPServer.ListenAndServe(); err != nil {
				errSig <- err
			}
		}()
		go func() {
			if err := publicHTTPServer.ListenAndServe(); err != nil {
				errSig <- err
			}
		}()
	}
	logger.Info("gRPC servers are running.")

	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quitSig, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errSig:
		logger.Error(fmt.Sprintf("Fatal error: %v\n", err))
	case <-quitSig:
		logger.Info("Shutting down server...")
		adminGrpcS.GracefulStop()
		publicGrpcS.GracefulStop()
	}
}
