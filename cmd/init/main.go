package main

import (
	"log"

	"github.com/instill-ai/mgmt-backend/config"
	"github.com/instill-ai/mgmt-backend/pkg/logger"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"

	database "github.com/instill-ai/mgmt-backend/pkg/db"
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

	// Create a default user
	if err := createDefaultUser(db); err != nil {
		logger.Fatal(err.Error())
	}

}
