package main

import (
	"context"
	"fmt"
	"log"

	"go.opentelemetry.io/otel"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	custom_otel "github.com/instill-ai/mgmt-backend/pkg/logger/otel"
	"github.com/instill-ai/x/temporal"
	"github.com/instill-ai/x/zapadapter"

	"github.com/instill-ai/mgmt-backend/config"
	"github.com/instill-ai/mgmt-backend/pkg/logger"

	mgmtWorker "github.com/instill-ai/mgmt-backend/pkg/worker"
)

func main() {

	if err := config.Init(); err != nil {
		log.Fatal(err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())

	// setup tracing and metrics
	if tp, err := custom_otel.SetupTracing(ctx, "mgmt-backend-worker"); err != nil {
		panic(err)
	} else {
		defer func() {
			err = tp.Shutdown(ctx)
		}()
	}

	ctx, span := otel.Tracer("worker-tracer").Start(ctx,
		"main",
	)
	defer cancel()

	logger, _ := logger.GetZapLogger(ctx)
	defer func() {
		// can't handle the error due to https://github.com/uber-go/zap/issues/880
		_ = logger.Sync()
	}()

	var err error

	var temporalClientOptions client.Options
	if config.Config.Temporal.Ca != "" && config.Config.Temporal.Cert != "" && config.Config.Temporal.Key != "" {
		if temporalClientOptions, err = temporal.GetTLSClientOption(
			config.Config.Temporal.HostPort,
			config.Config.Temporal.Namespace,
			zapadapter.NewZapAdapter(logger),
			config.Config.Temporal.Ca,
			config.Config.Temporal.Cert,
			config.Config.Temporal.Key,
			config.Config.Temporal.ServerName,
			true,
		); err != nil {
			logger.Fatal(fmt.Sprintf("Unable to get Temporal client options: %s", err))
		}
	} else {
		if temporalClientOptions, err = temporal.GetClientOption(
			config.Config.Temporal.HostPort,
			config.Config.Temporal.Namespace,
			zapadapter.NewZapAdapter(logger)); err != nil {
			logger.Fatal(fmt.Sprintf("Unable to get Temporal client options: %s", err))
		}
	}

	temporalClient, err := client.Dial(temporalClientOptions)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Unable to create client: %s", err))
	}
	defer temporalClient.Close()

	w := worker.New(temporalClient, mgmtWorker.TaskQueue, worker.Options{
		MaxConcurrentActivityExecutionSize: 2,
	})

	span.End()
	err = w.Run(worker.InterruptCh())
	if err != nil {
		logger.Fatal(fmt.Sprintf("Unable to start worker: %s", err))
	}
}
