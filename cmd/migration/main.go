package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"gorm.io/gorm"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	openfgaclient "github.com/openfga/go-sdk/client"

	"github.com/instill-ai/mgmt-backend/config"
	"github.com/instill-ai/mgmt-backend/pkg/acl/fga"
	"github.com/instill-ai/mgmt-backend/pkg/datamodel"
	"github.com/instill-ai/mgmt-backend/pkg/db"
	"github.com/instill-ai/mgmt-backend/pkg/db/migration"
	"github.com/instill-ai/mgmt-backend/pkg/logger"
)

func checkExist(databaseConfig config.DatabaseConfig) error {
	db, err := sql.Open(
		"postgres",
		fmt.Sprintf("host=%s user=%s password=%s dbname=postgres port=%d sslmode=disable TimeZone=%s",
			databaseConfig.Host,
			databaseConfig.Username,
			databaseConfig.Password,
			databaseConfig.Port,
			databaseConfig.TimeZone,
		),
	)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	// Open() may just validate its arguments without creating a connection to the database.
	// To verify that the data source name is valid, call Ping().
	if err = db.Ping(); err != nil {
		panic(err)
	}

	var rows *sql.Rows
	rows, err = db.Query(fmt.Sprintf("SELECT datname FROM pg_catalog.pg_database WHERE lower(datname) = lower('%s');", databaseConfig.Name))

	if err != nil {
		panic(err)
	}

	dbExist := false
	defer rows.Close()
	for rows.Next() {
		var databaseName string
		if err := rows.Scan(&databaseName); err != nil {
			panic(err)
		}

		if databaseConfig.Name == databaseName {
			dbExist = true
			fmt.Printf("Database %s exist\n", databaseName)
		}
	}

	if err := rows.Err(); err != nil {
		panic(err)
	}

	if !dbExist {
		fmt.Printf("Create database %s\n", databaseConfig.Name)
		if _, err := db.Exec(fmt.Sprintf("CREATE DATABASE %s;", databaseConfig.Name)); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	ctx := context.Background()

	if err := config.Init(config.ParseConfigFlag()); err != nil {
		panic(err)
	}

	databaseConfig := config.Config.Database

	if err := checkExist(databaseConfig); err != nil {
		panic(err)
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?%s",
		databaseConfig.Username,
		databaseConfig.Password,
		databaseConfig.Host,
		databaseConfig.Port,
		databaseConfig.Name,
		"sslmode=disable",
	)

	codeMigrator, cleanup := initCodeMigrator(ctx)
	defer cleanup()

	runMigration(dsn, uint(db.TargetSchemaVersion), codeMigrator.Migrate)

}

func runMigration(
	dsn string,
	expectedVersion uint,
	execCode func(version uint) error,
) {
	migrateFolder, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	m, err := migrate.New(fmt.Sprintf("file:///%s/pkg/db/migration", migrateFolder), dsn)

	if err != nil {
		panic(err)
	}

	curVersion, dirty, err := m.Version()

	if err != nil && curVersion != 0 {
		panic(err)
	}

	fmt.Printf("Expected migration version is %d\n", expectedVersion)
	fmt.Printf("The current schema version is %d, and dirty flag is %t\n", curVersion, dirty)

	if dirty {
		panic("The database has dirty flag, please fix it")
	}

	step := curVersion
	for {
		if expectedVersion <= step {
			fmt.Printf("Migration to version %d complete\n", expectedVersion)
			break
		}

		fmt.Printf("Step up to version %d\n", step+1)
		if err := m.Steps(1); err != nil {
			panic(err)
		}

		if step, _, err = m.Version(); err != nil {
			panic(err)
		}

		if err := execCode(step); err != nil {
			panic(err)
		}
	}

	ctx := context.Background()
	gormDB := db.GetConnection(&config.Config.Database)
	err = runFGAMigration(ctx, gormDB)
	if err != nil {
		panic(err)
	}

}

func initCodeMigrator(ctx context.Context) (cm *migration.CodeMigrator, cleanup func()) {
	l, _ := logger.GetZapLogger(ctx)
	cleanups := make([]func(), 0)

	gormDB := db.GetConnection(&config.Config.Database)
	cleanups = append(cleanups, func() { db.Close(gormDB) })

	codeMigrator := &migration.CodeMigrator{
		Logger: l,
		DB:     gormDB,
		Config: &config.Config,
	}

	return codeMigrator, func() {
		for _, cleanup := range cleanups {
			cleanup()
		}
	}
}

func runFGAMigration(ctx context.Context, db *gorm.DB) error {

	log, _ := logger.GetZapLogger(ctx)
	var fgaClient *openfgaclient.OpenFgaClient
	var err error

	fgaClient, err = openfgaclient.NewSdkClient(&openfgaclient.ClientConfiguration{
		ApiScheme: "http",
		ApiHost:   fmt.Sprintf("%s:%d", config.Config.OpenFGA.Host, config.Config.OpenFGA.Port),
	})
	if err != nil {
		return fmt.Errorf("creating FGA client: %w", err)
	}

	var existingFgaMigration datamodel.FGAMigration
	err = db.Raw("SELECT store_id, authorization_model_id, md5_hash FROM fga_migrations LIMIT 1").Scan(&existingFgaMigration).Error
	// If no record found or existing record has empty store ID, create a new store
	if err != nil || existingFgaMigration.StoreID == "" {
		log.Info("Creating new store")
		store, err := fgaClient.CreateStore(context.Background()).Body(openfgaclient.ClientCreateStoreRequest{Name: "instill"}).Execute()
		if err != nil {
			return fmt.Errorf("creating store: %w", err)
		}

		err = db.Model(&datamodel.FGAMigration{}).Create(&datamodel.FGAMigration{
			StoreID: store.Id,
		}).Error
		if err != nil {
			return fmt.Errorf("creating store: %w", err)
		}
		existingFgaMigration.StoreID = store.Id
	}

	err = fgaClient.SetStoreId(existingFgaMigration.StoreID)
	if err != nil {
		return fmt.Errorf("setting store ID: %w", err)
	}

	if existingFgaMigration.AuthorizationModelID == "" || existingFgaMigration.MD5Hash != fga.ACLModelMD5 {
		var body openfgaclient.ClientWriteAuthorizationModelRequest
		if err := json.Unmarshal([]byte(fga.ACLModelBytes), &body); err != nil {
			return fmt.Errorf("unmarshalling authorization model: %w", err)
		}

		am, err := fgaClient.WriteAuthorizationModel(context.Background()).Body(body).Execute()
		if err != nil {
			return fmt.Errorf("writing authorization model: %w", err)
		}

		existingFgaMigration.AuthorizationModelID = am.AuthorizationModelId
		existingFgaMigration.MD5Hash = fga.ACLModelMD5
		err = db.Model(&existingFgaMigration).Where("store_id = ?", existingFgaMigration.StoreID).Updates(existingFgaMigration).Error
		if err != nil {
			return fmt.Errorf("updating authorization model: %w", err)
		}
	}

	return nil
}
