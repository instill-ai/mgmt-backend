package main

import (
	"database/sql"
	"log"

	"github.com/gofrs/uuid"
	"github.com/instill-ai/mgmt-backend/config"
	"github.com/instill-ai/mgmt-backend/pkg/datamodel"
	"github.com/instill-ai/mgmt-backend/pkg/logger"
	"github.com/instill-ai/mgmt-backend/pkg/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"

	database "github.com/instill-ai/mgmt-backend/pkg/db"
	mgmtPB "github.com/instill-ai/protogen-go/vdp/mgmt/v1alpha"
)

// CreateDefaultUser creates a default user in the database
// Return error types
//   - codes.Internal
func createDefaultUser(db *gorm.DB) error {

	// Generate a random uid to the user
	defaultUserUID, err := uuid.NewV4()
	if err != nil {
		return status.Errorf(codes.Internal, "error %v", err)
	}

	r := repository.NewRepository(db)

	defaultUser := datamodel.User{
		Base:                   datamodel.Base{UID: defaultUserUID},
		ID:                     config.DefaultUserID,
		OwnerType:              sql.NullString{String: datamodel.PBUserType2DBUserType[mgmtPB.OwnerType_OWNER_TYPE_USER], Valid: true},
		Email:                  config.DefaultUserEmail,
		CustomerId:             "",
		FirstName:              sql.NullString{String: "", Valid: false},
		LastName:               sql.NullString{String: "", Valid: false},
		OrgName:                sql.NullString{String: "", Valid: false},
		Role:                   sql.NullString{String: "", Valid: false},
		NewsletterSubscription: false,
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
	return r.CreateUser(&defaultUser)
}

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

	db := database.GetConnection(&config.Config.Database)
	defer database.Close(db)

	// Create a default user
	if err := createDefaultUser(db); err != nil {
		logger.Fatal(err.Error())
	}

}
