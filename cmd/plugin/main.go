// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"log"

	"github.com/hashicorp/go-plugin"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"

	"github.com/instill-ai/mgmt-backend/config"
	"github.com/instill-ai/mgmt-backend/internal/external"
	"github.com/instill-ai/mgmt-backend/internal/shared"
	"github.com/instill-ai/mgmt-backend/pkg/handler"
	"github.com/instill-ai/mgmt-backend/pkg/logger"
	"github.com/instill-ai/mgmt-backend/pkg/repository"
	"github.com/instill-ai/mgmt-backend/pkg/service"
	"github.com/instill-ai/mgmt-backend/pkg/usage"

	database "github.com/instill-ai/mgmt-backend/internal/db"
)

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

	repository := repository.NewRepository(db)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Usage collection
	var usg usage.Usage
	if !config.Config.Server.DisableUsage {
		usageServiceClient, usageServiceClientConn := external.InitUsageServiceClient()
		if usageServiceClientConn != nil {
			defer usageServiceClientConn.Close()
			usg = usage.NewUsage(ctx, repository, usageServiceClient)
			if usg != nil {
				usg.StartReporter(ctx)
			}
		}
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: shared.Handshake,
		Plugins: map[string]plugin.Plugin{
			"admin handler": &shared.HandlerAdminPlugin{
				Impl: handler.NewAdminHandler(service.NewService(repository)),
			},
			"public handler": &shared.HandlerPublicPlugin{
				Impl: handler.NewPublicHandler(service.NewService(repository), usg),
			},
		},
	})
}
