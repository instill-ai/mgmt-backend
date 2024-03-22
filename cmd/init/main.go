package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/gofrs/uuid"
	"github.com/instill-ai/mgmt-backend/config"
	"github.com/instill-ai/mgmt-backend/pkg/acl"
	"github.com/instill-ai/mgmt-backend/pkg/constant"
	"github.com/instill-ai/mgmt-backend/pkg/datamodel"
	"github.com/instill-ai/mgmt-backend/pkg/logger"
	"github.com/instill-ai/mgmt-backend/pkg/repository"
	"github.com/instill-ai/mgmt-backend/pkg/service"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	openfga "github.com/openfga/go-sdk/client"

	database "github.com/instill-ai/mgmt-backend/pkg/db"
	custom_otel "github.com/instill-ai/mgmt-backend/pkg/logger/otel"
	mgmtPB "github.com/instill-ai/protogen-go/core/mgmt/v1beta"
)

// CreateDefaultUser creates a default user in the database
// Return error types
//   - codes.Internal
func createDefaultUser(ctx context.Context, db *gorm.DB) error {

	// Generate a random uid to the user
	var defaultUserUID uuid.UUID
	var err error
	if config.Config.Server.DefaultUserUID != "" {
		defaultUserUID, err = uuid.FromString(config.Config.Server.DefaultUserUID)
	} else {
		defaultUserUID, err = uuid.NewV4()
	}
	if err != nil {
		return status.Errorf(codes.Internal, "uuid generation error %v", err)
	}

	redisClient := redis.NewClient(&config.Config.Cache.Redis.RedisOptions)
	defer redisClient.Close()
	r := repository.NewRepository(db, redisClient)

	passwordBytes, err := bcrypt.GenerateFromPassword([]byte(constant.DefaultUserPassword), 10)
	if err != nil {
		return err
	}

	defaultUser := datamodel.Owner{
		Base:                   datamodel.Base{UID: defaultUserUID},
		ID:                     constant.DefaultUserID,
		OwnerType:              sql.NullString{String: service.PBUserType2DBUserType[mgmtPB.OwnerType_OWNER_TYPE_USER], Valid: true},
		Email:                  constant.DefaultUserEmail,
		CustomerID:             constant.DefaultUserCustomerID,
		DisplayName:            sql.NullString{String: constant.DefaultUserDisplayName, Valid: true},
		CompanyName:            sql.NullString{String: constant.DefaultUserCompanyName, Valid: true},
		Role:                   sql.NullString{String: constant.DefaultUserRole, Valid: true},
		NewsletterSubscription: constant.DefaultUserNewsletterSubscription,
		CookieToken:            sql.NullString{String: "", Valid: false},
		OnboardingStatus:       datamodel.OnboardingStatusInProgress,
	}

	user, err := r.GetUser(context.Background(), constant.DefaultUserID)
	// Default user already exists
	if err == nil {
		passwordHash, _, err := r.GetUserPasswordHash(ctx, user.UID)
		if err != nil {
			return err
		}

		if passwordHash == "" {
			err = r.UpdateUserPasswordHash(ctx, user.UID, string(passwordBytes), time.Now())
			if err != nil {
				return err
			}
		}
		return nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return status.Errorf(codes.Internal, "error %v", err)
	}

	// Create the default user
	err = r.CreateUser(ctx, &defaultUser)
	if err != nil {
		return err
	}
	err = r.UpdateUserPasswordHash(ctx, defaultUser.UID, string(passwordBytes), time.Now())
	if err != nil {
		return err
	}
	return nil
}

func main() {
	if err := config.Init(); err != nil {
		log.Fatal(err.Error())
	}

	// setup tracing and metrics
	ctx, cancel := context.WithCancel(context.Background())

	if tp, err := custom_otel.SetupTracing(ctx, "mgmt-backend-init"); err != nil {
		panic(err)
	} else {
		defer func() {
			err = tp.Shutdown(ctx)
		}()
	}

	ctx, span := otel.Tracer("init-tracer").Start(ctx,
		"main",
	)
	defer span.End()
	defer cancel()

	logger, _ := logger.GetZapLogger(ctx)
	defer func() {
		// can't handle the error due to https://github.com/uber-go/zap/issues/880
		_ = logger.Sync()
	}()
	grpc_zap.ReplaceGrpcLoggerV2(logger)

	db := database.GetConnection(&config.Config.Database)
	defer database.Close(db)

	// Create a default user
	if err := createDefaultUser(ctx, db); err != nil {
		logger.Fatal(err.Error())
	}

	fgaClient, err := openfga.NewSdkClient(&openfga.ClientConfiguration{
		ApiScheme: "http",
		ApiHost:   fmt.Sprintf("%s:%d", config.Config.OpenFGA.Host, config.Config.OpenFGA.Port),
	})

	if err != nil {
		panic(err)
		// .. Handle error
	}

	stores, err := fgaClient.ListStores(context.Background()).Execute()
	if err != nil {
		panic(err)
	}
	storeID := ""
	if len(*stores.Stores) == 0 {
		data, err := fgaClient.CreateStore(context.Background()).Body(openfga.ClientCreateStoreRequest{Name: "instill"}).Execute()
		if err != nil {
			panic(err)
		}
		storeID = *data.Id
	} else {
		storeID = *(*stores.Stores)[0].Id
	}

	fgaClient.SetStoreId(storeID)

	models, err := fgaClient.ReadAuthorizationModels(context.Background()).Execute()
	if err != nil {
		panic(err)
	}
	if len(*models.AuthorizationModels) == 0 {
		var body openfga.ClientWriteAuthorizationModelRequest
		if err := json.Unmarshal([]byte(acl.ACLModel), &body); err != nil {
			panic(err)
		}

		_, err = fgaClient.WriteAuthorizationModel(context.Background()).Body(body).Execute()
		if err != nil {
			panic(err)
		}
	}

}
