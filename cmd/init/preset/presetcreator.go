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

	mgmtpb "github.com/instill-ai/protogen-go/core/mgmt/v1beta"
	errorsx "github.com/instill-ai/x/errors"
)

// CreatePresetUser creates a preset user namespace for storing preset
// resources, such as preset pipelines.
func CreatePresetUser(ctx context.Context, r repository.Repository) error {

	presetUser := &datamodel.Owner{
		Base:        datamodel.Base{UID: uuid.FromStringOrNil(constant.PresetUserUID)},
		ID:          constant.PresetUserID,
		OwnerType:   sql.NullString{String: service.PBUserType2DBUserType[mgmtpb.OwnerType_OWNER_TYPE_USER], Valid: true},
		DisplayName: sql.NullString{String: constant.PresetUserDisplayName, Valid: true},
	}

	_, err := r.GetUser(ctx, constant.PresetUserID, false)
	if err == nil {
		return nil
	}

	if !errors.Is(err, errorsx.ErrNotFound) {
		return status.Errorf(codes.Internal, "error %v", err)
	}

	// Create the preset user
	err = r.CreateUser(ctx, presetUser)
	if err != nil {
		return err
	}
	return nil
}
