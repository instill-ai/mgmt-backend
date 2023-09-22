package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/gofrs/uuid"
	"github.com/instill-ai/mgmt-backend/config"
	"github.com/instill-ai/mgmt-backend/pkg/constant"
	"github.com/instill-ai/mgmt-backend/pkg/datamodel"
	"github.com/instill-ai/mgmt-backend/pkg/logger"
	"github.com/instill-ai/mgmt-backend/pkg/repository"
	"go.opentelemetry.io/otel"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"

	database "github.com/instill-ai/mgmt-backend/pkg/db"
	custom_otel "github.com/instill-ai/mgmt-backend/pkg/logger/otel"
	mgmtPB "github.com/instill-ai/protogen-go/base/mgmt/v1alpha"
)

// CreateDefaultUser creates a default user in the database
// Return error types
//   - codes.Internal
func createDefaultUser(ctx context.Context, db *gorm.DB) error {

	// Generate a random uid to the user
	var defaultUserUID uuid.UUID
	var err error
	if config.Config.Server.DefaultUserUid != "" {
		defaultUserUID, err = uuid.FromString(config.Config.Server.DefaultUserUid)
	} else {
		defaultUserUID, err = uuid.NewV4()
	}
	if err != nil {
		return status.Errorf(codes.Internal, "uuid generation error %v", err)
	}

	r := repository.NewRepository(db)

	passwordBytes, err := bcrypt.GenerateFromPassword([]byte(constant.DefaultUserPassword), 10)
	if err != nil {
		return err
	}

	defaultUser := datamodel.User{
		Base:                   datamodel.Base{UID: defaultUserUID},
		ID:                     constant.DefaultUserID,
		OwnerType:              sql.NullString{String: datamodel.PBUserType2DBUserType[mgmtPB.OwnerType_OWNER_TYPE_USER], Valid: true},
		Email:                  constant.DefaultUserEmail,
		CustomerId:             constant.DefaultUserCustomerId,
		FirstName:              sql.NullString{String: constant.DefaultUserFirstName, Valid: true},
		LastName:               sql.NullString{String: constant.DefaultUserLastName, Valid: true},
		OrgName:                sql.NullString{String: constant.DefaultUserOrgName, Valid: true},
		Role:                   sql.NullString{String: constant.DefaultUserRole, Valid: true},
		NewsletterSubscription: constant.DefaultUserNewsletterSubscription,
		CookieToken:            sql.NullString{String: "", Valid: false},
	}

	user, err := r.GetUserByID(constant.DefaultUserID)
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

	if s, ok := status.FromError(err); !ok || s.Code() != codes.NotFound {
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

}
