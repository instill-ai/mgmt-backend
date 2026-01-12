package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/gofrs/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"

	"github.com/instill-ai/mgmt-backend/cmd/init/preset"
	"github.com/instill-ai/mgmt-backend/config"
	"github.com/instill-ai/mgmt-backend/pkg/constant"
	"github.com/instill-ai/mgmt-backend/pkg/datamodel"
	"github.com/instill-ai/mgmt-backend/pkg/repository"
	"github.com/instill-ai/mgmt-backend/pkg/service"

	database "github.com/instill-ai/mgmt-backend/pkg/db"
	mgmtpb "github.com/instill-ai/protogen-go/core/mgmt/v1beta"
	errorsx "github.com/instill-ai/x/errors"
	logx "github.com/instill-ai/x/log"
)

func main() {
	if err := config.Init(config.ParseConfigFlag()); err != nil {
		log.Fatal(err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	db := database.GetConnection(&config.Config.Database)
	defer database.Close(db)

	redisClient := redis.NewClient(&config.Config.Cache.Redis.RedisOptions)
	defer redisClient.Close()
	r := repository.NewRepository(db, redisClient)

	// Create a default user
	if err := createDefaultUser(ctx, r); err != nil {
		logger.Fatal(err.Error())
	}

	// Create preset user namespace for storing preset resources
	if err := preset.CreatePresetUser(ctx, r); err != nil {
		logger.Fatal(err.Error())
	}

}

// CreateDefaultUser creates a default user in the database
// Return error types
//   - codes.Internal
func createDefaultUser(ctx context.Context, r repository.Repository) error {

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

	passwordBytes, err := bcrypt.GenerateFromPassword([]byte(constant.DefaultUserPassword), 10)
	if err != nil {
		return err
	}

	defaultUser := datamodel.Owner{
		Base:                   datamodel.Base{UID: defaultUserUID},
		ID:                     constant.DefaultUserID,
		OwnerType:              sql.NullString{String: service.PBUserType2DBUserType[mgmtpb.OwnerType_OWNER_TYPE_USER], Valid: true},
		Email:                  constant.DefaultUserEmail,
		CustomerID:             constant.DefaultUserCustomerID,
		DisplayName:            sql.NullString{String: constant.DefaultUserDisplayName, Valid: true},
		CompanyName:            sql.NullString{String: constant.DefaultUserCompanyName, Valid: true},
		Role:                   sql.NullString{String: constant.DefaultUserRole, Valid: true},
		NewsletterSubscription: constant.DefaultUserNewsletterSubscription,
		CookieToken:            sql.NullString{String: "", Valid: false},
		OnboardingStatus:       datamodel.OnboardingStatusInProgress,
	}

	user, err := r.GetUser(ctx, constant.DefaultUserID, false)
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

	if !errors.Is(err, errorsx.ErrNotFound) {
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
