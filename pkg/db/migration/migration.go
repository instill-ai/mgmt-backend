package migration

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/instill-ai/mgmt-backend/config"
	openfgaClient "github.com/openfga/go-sdk/client"
)

type migration interface {
	Migrate() error
}

// CodeMigrator orchestrates the execution of the code associated with the
// different database migrations and holds their dependencies.
type CodeMigrator struct {
	Logger *zap.Logger
	DB     *gorm.DB
	Config *config.AppConfig
}

// Migrate executes custom code as part of a database migration. This code is
// intended to be run only once and typically goes along a change
// in the database schemas. Some use cases might be backfilling a table or
// updating some existing records according to the schema changes.
//
// Note that the changes in the database schemas shouldn't be run here, only
// code accompanying them.
func (cm *CodeMigrator) Migrate(version uint) error {
	var m migration

	switch version {
	case 6:
		m = &FGAMigration{
			DB:     cm.DB,
			Logger: cm.Logger,
			Config: cm.Config,
		}
	default:
		return nil
	}

	return m.Migrate()
}

// FGAMigration handles the migration of OpenFGA store_id and authorization_model_id
type FGAMigration struct {
	DB     *gorm.DB
	Logger *zap.Logger
	Config *config.AppConfig
}

// Migrate reads store_id and authorization_model_id from OpenFGA and stores them in fga_migrations table
func (fm *FGAMigration) Migrate() error {
	fm.Logger.Info("Starting FGA migration to read store_id and authorization_model_id")

	// Create OpenFGA client
	fgaClient, err := openfgaClient.NewSdkClient(&openfgaClient.ClientConfiguration{
		ApiScheme: "http",
		ApiHost:   fmt.Sprintf("%s:%d", fm.Config.OpenFGA.Host, fm.Config.OpenFGA.Port),
	})
	if err != nil {
		return fmt.Errorf("failed to create OpenFGA client: %w", err)
	}

	// Get store_id
	stores, err := fgaClient.ListStores(context.Background()).Execute()
	if err != nil {
		return fmt.Errorf("failed to list OpenFGA stores: %w", err)
	}

	if len(stores.Stores) == 0 {
		return fmt.Errorf("no OpenFGA stores found")
	}

	storeID := stores.Stores[0].Id
	fm.Logger.Info("Found OpenFGA store", zap.String("store_id", storeID))

	// Insert or update the fga_migrations table
	result := fm.DB.Exec(`
		INSERT INTO fga_migrations (store_id, authorization_model_id, md5_hash)
		VALUES (?, ?, ?)
	`, storeID, "", "")

	if result.Error != nil {
		return fmt.Errorf("failed to insert/update fga_migrations table: %w", result.Error)
	}

	fm.Logger.Info("Successfully migrated FGA data",
		zap.String("store_id", storeID))

	return nil
}
