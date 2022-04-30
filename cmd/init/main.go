package main

import (
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/instill-ai/mgmt-backend/configs"
	"github.com/instill-ai/mgmt-backend/internal/logger"

	database "github.com/instill-ai/mgmt-backend/internal/db"
)

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

	// Create a default user
	if err := createDefaultUser(db); err != nil {
		logger.Fatal(err.Error())
	}

}
