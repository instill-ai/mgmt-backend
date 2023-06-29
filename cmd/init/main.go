package main

import (
	"context"
	"database/sql"
	"log"

	"github.com/gofrs/uuid"
	"github.com/instill-ai/mgmt-backend/config"
	"github.com/instill-ai/mgmt-backend/pkg/constant"
	"github.com/instill-ai/mgmt-backend/pkg/datamodel"
	"github.com/instill-ai/mgmt-backend/pkg/logger"
	"github.com/instill-ai/mgmt-backend/pkg/repository"
	"go.opentelemetry.io/otel"
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
	defaultUserUID, err := uuid.NewV4()
	if err != nil {
		return status.Errorf(codes.Internal, "error %v", err)
	}

	r := repository.NewRepository(db)

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

	_, err = r.GetUserByID(defaultUser.ID)
	// Default user already exists
	if err == nil {
		return nil
	}

	if s, ok := status.FromError(err); !ok || s.Code() != codes.NotFound {
		return status.Errorf(codes.Internal, "error %v", err)
	}

	// Create the default user
	return r.CreateUser(ctx, &defaultUser)
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
