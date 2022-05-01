package main

import (
	"database/sql"

	"github.com/gofrs/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"github.com/instill-ai/mgmt-backend/pkg/datamodel"
	"github.com/instill-ai/mgmt-backend/pkg/repository"
)

// DefaultUserId is the UUID id of the default user
const DefaultUserID string = "2a06c2f7-8da9-4046-91ea-240f88a5d729"

// CreateDefaultUser creates a default user in the database
func createDefaultUser(db *gorm.DB) error {
	defaultID, err := uuid.FromString(DefaultUserID)
	if err != nil {
		return err
	}

	r := repository.NewRepository(db)

	defaultUser := datamodel.User{
		Base:                   datamodel.Base{ID: defaultID},
		Email:                  sql.NullString{String: "hello@instill.tech", Valid: true},
		Login:                  "instill",
		CompanyName:            sql.NullString{String: "Instill AI", Valid: true},
		Role:                   sql.NullString{String: "", Valid: false},
		UsageDataCollection:    false,
		NewsletterSubscription: false,
	}

	_, err = r.GetUser(defaultUser.Base.ID)
	// Already exist
	if err == nil {
		return nil
	}

	if s, ok := status.FromError(err); !ok || s.Code() != codes.NotFound {
		return status.Errorf(codes.Internal, "Error %v", err)
	}

	// Create the default user
	return r.CreateUser(&defaultUser)
}
