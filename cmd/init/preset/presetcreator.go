package preset

import (
	"context"
	"database/sql"
	"errors"

	"github.com/gofrs/uuid"
	"github.com/gogo/status"
	"google.golang.org/grpc/codes"

	"github.com/instill-ai/mgmt-backend/pkg/constant"
	"github.com/instill-ai/mgmt-backend/pkg/datamodel"
	"github.com/instill-ai/mgmt-backend/pkg/repository"
	"github.com/instill-ai/mgmt-backend/pkg/service"

	mgmtpb "github.com/instill-ai/protogen-go/mgmt/v1beta"
	errorsx "github.com/instill-ai/x/errors"
)

// CreatePresetUser creates a preset user namespace for storing preset
// resources, such as preset pipelines.
// If the preset user already exists with incorrect owner_type (e.g., from EE),
// it will be corrected to ensure it's a "user" type.
func CreatePresetUser(ctx context.Context, r repository.Repository) error {
	expectedOwnerType := service.PBUserType2DBUserType[mgmtpb.OwnerType_OWNER_TYPE_USER]

	presetUser := &datamodel.Owner{
		Base:        datamodel.Base{UID: uuid.FromStringOrNil(constant.PresetUserUID)},
		ID:          constant.PresetUserID,
		OwnerType:   sql.NullString{String: expectedOwnerType, Valid: true},
		DisplayName: sql.NullString{String: constant.PresetUserDisplayName, Valid: true},
	}

	existingUser, err := r.GetUser(ctx, constant.PresetUserID, false)
	if err == nil {
		// User exists - verify owner_type is correct
		// If preset was previously created as "organization" (e.g., from EE), fix it
		if existingUser.OwnerType.String != expectedOwnerType {
			// Update the owner_type to "user"
			existingUser.OwnerType = sql.NullString{String: expectedOwnerType, Valid: true}
			if updateErr := r.UpdateUser(ctx, constant.PresetUserID, existingUser); updateErr != nil {
				return status.Errorf(codes.Internal, "failed to update preset user owner_type: %v", updateErr)
			}
		}
		return nil
	}

	if !errors.Is(err, errorsx.ErrNotFound) {
		return status.Errorf(codes.Internal, "error %v", err)
	}

	// Create the preset user
	err = r.CreateUser(ctx, presetUser)
	if err != nil {
		// Handle race condition: another process may have created the user
		if errors.Is(err, errorsx.ErrAlreadyExists) {
			return nil
		}
		return err
	}
	return nil
}
