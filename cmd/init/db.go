package main

import (
	"database/sql"

	"github.com/gofrs/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"github.com/instill-ai/mgmt-backend/config"
	"github.com/instill-ai/mgmt-backend/pkg/datamodel"
	"github.com/instill-ai/mgmt-backend/pkg/repository"

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
		Plan:                   sql.NullString{String: "plans/open-source", Valid: true},
		BillingId:              sql.NullString{String: "", Valid: false},
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
