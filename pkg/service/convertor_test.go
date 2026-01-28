package service

import (
	"context"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/instill-ai/mgmt-backend/pkg/datamodel"

	mgmtpb "github.com/instill-ai/protogen-go/mgmt/v1beta"
)

func TestPBAuthenticatedUser2DBUser_CreateCase(t *testing.T) {
	// Create a minimal service instance for testing
	s := &service{}

	pbUser := &mgmtpb.AuthenticatedUser{
		Email: "test@example.com",
		Profile: &mgmtpb.UserProfile{
			DisplayName: "Test User",
			CompanyName: proto.String("Test Company"),
			Bio:         proto.String("Test bio"),
		},
		Role:                   proto.String("engineer"),
		NewsletterSubscription: true,
	}

	// Create case: existingUser is nil
	dbUser, err := s.PBAuthenticatedUser2DBUser(context.Background(), pbUser, nil)

	require.NoError(t, err)
	assert.NotNil(t, dbUser)

	// UID should be generated (non-zero)
	assert.NotEqual(t, uuid.UUID{}, dbUser.UID, "UID should be generated for new user")

	// ID should be generated with "usr-" prefix
	assert.NotEmpty(t, dbUser.ID, "ID should be generated for new user")
	assert.Contains(t, dbUser.ID, "usr-", "ID should have 'usr-' prefix")

	// Other fields should be set correctly
	assert.Equal(t, "test@example.com", dbUser.Email)
	assert.Equal(t, "Test User", dbUser.DisplayName.String)
	assert.Equal(t, "Test Company", dbUser.CompanyName.String)
	assert.Equal(t, "Test bio", dbUser.Bio.String)
	assert.Equal(t, "engineer", dbUser.Role.String)
	assert.True(t, dbUser.NewsletterSubscription)
}

func TestPBAuthenticatedUser2DBUser_UpdateCase(t *testing.T) {
	// Create a minimal service instance for testing
	s := &service{}

	// Existing user with known UID and ID
	existingUID := uuid.Must(uuid.NewV4())
	existingID := "usr-existingID123"

	existingUser := &datamodel.Owner{
		Base: datamodel.Base{
			UID: existingUID,
		},
		ID: existingID,
	}

	pbUser := &mgmtpb.AuthenticatedUser{
		Email: "updated@example.com",
		Profile: &mgmtpb.UserProfile{
			DisplayName: "Updated User",
			CompanyName: proto.String("Updated Company"),
		},
		Role: proto.String("manager"),
	}

	// Update case: existingUser is provided
	dbUser, err := s.PBAuthenticatedUser2DBUser(context.Background(), pbUser, existingUser)

	require.NoError(t, err)
	assert.NotNil(t, dbUser)

	// UID and ID should be preserved from existing user (immutable)
	assert.Equal(t, existingUID, dbUser.UID, "UID should be preserved from existing user")
	assert.Equal(t, existingID, dbUser.ID, "ID should be preserved from existing user")

	// Other fields should be updated
	assert.Equal(t, "updated@example.com", dbUser.Email)
	assert.Equal(t, "Updated User", dbUser.DisplayName.String)
	assert.Equal(t, "Updated Company", dbUser.CompanyName.String)
	assert.Equal(t, "manager", dbUser.Role.String)
}

func TestPBAuthenticatedUser2DBUser_NilUser(t *testing.T) {
	s := &service{}

	// Nil pbUser should return error
	dbUser, err := s.PBAuthenticatedUser2DBUser(context.Background(), nil, nil)

	assert.Error(t, err)
	assert.Nil(t, dbUser)
	assert.Contains(t, err.Error(), "can't convert a nil user")
}

func TestPBAuthenticatedUser2DBUser_CreateGeneratesUniqueIDs(t *testing.T) {
	s := &service{}

	pbUser := &mgmtpb.AuthenticatedUser{
		Email: "test@example.com",
		Profile: &mgmtpb.UserProfile{
			DisplayName: "Test User",
		},
	}

	// Create two users and verify they get different UIDs and IDs
	dbUser1, err1 := s.PBAuthenticatedUser2DBUser(context.Background(), pbUser, nil)
	dbUser2, err2 := s.PBAuthenticatedUser2DBUser(context.Background(), pbUser, nil)

	require.NoError(t, err1)
	require.NoError(t, err2)

	assert.NotEqual(t, dbUser1.UID, dbUser2.UID, "Each new user should get a unique UID")
	assert.NotEqual(t, dbUser1.ID, dbUser2.ID, "Each new user should get a unique ID")
}
