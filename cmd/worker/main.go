package main

import (
	"context"
	"fmt"
	"log"

	"go.opentelemetry.io/otel"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/opentelemetry"
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/worker"
	"go.uber.org/zap"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"

	"github.com/instill-ai/mgmt-backend/config"
	"github.com/instill-ai/x/temporal"

	mgmtworker "github.com/instill-ai/mgmt-backend/pkg/worker"
	logx "github.com/instill-ai/x/log"
	otelx "github.com/instill-ai/x/otel"
)

var (
	// These variables might be overridden at buildtime.
	serviceName    = "mgmt-backend-worker"
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

	var err error

	temporalClientOptions, err := temporal.ClientOptions(config.Config.Temporal, logger)
	if err != nil {
		logger.Fatal("Unable to build Temporal client options", zap.Error(err))
	}

	// Only add interceptor if tracing is enabled
	if config.Config.OTELCollector.Enable {
		temporalTracingInterceptor, err := opentelemetry.NewTracingInterceptor(opentelemetry.TracerOptions{
			Tracer:            otel.Tracer(serviceName),
			TextMapPropagator: otel.GetTextMapPropagator(),
		})
		if err != nil {
			logger.Fatal("Unable to create temporal tracing interceptor", zap.Error(err))
		}
		temporalClientOptions.Interceptors = []interceptor.ClientInterceptor{temporalTracingInterceptor}
	}

	temporalClient, err := client.Dial(temporalClientOptions)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Unable to create client: %s", err))
	}
	defer temporalClient.Close()

	w := worker.New(temporalClient, mgmtworker.TaskQueue, worker.Options{
		MaxConcurrentActivityExecutionSize: 2,
		Interceptors: func() []interceptor.WorkerInterceptor {
			if !config.Config.OTELCollector.Enable {
				return nil
			}
			workerInterceptor, err := opentelemetry.NewTracingInterceptor(opentelemetry.TracerOptions{
				Tracer:            otel.Tracer(serviceName),
				TextMapPropagator: otel.GetTextMapPropagator(),
			})
			if err != nil {
				logger.Fatal("Unable to create worker tracing interceptor", zap.Error(err))
			}
			return []interceptor.WorkerInterceptor{workerInterceptor}
		}(),
	})

	err = w.Run(worker.InterruptCh())
	if err != nil {
		logger.Fatal(fmt.Sprintf("Unable to start worker: %s", err))
	}
}
