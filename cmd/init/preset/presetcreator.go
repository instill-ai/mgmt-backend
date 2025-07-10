package preset

import (
	"context"
	"database/sql"
	"errors"

	"github.com/gofrs/uuid"
	"github.com/gogo/status"
	"google.golang.org/grpc/codes"
	"gorm.io/gorm"

	"github.com/instill-ai/mgmt-backend/pkg/constant"
	"github.com/instill-ai/mgmt-backend/pkg/datamodel"
	"github.com/instill-ai/mgmt-backend/pkg/repository"
	"github.com/instill-ai/mgmt-backend/pkg/service"

	mgmtpb "github.com/instill-ai/protogen-go/core/mgmt/v1beta"
)

func CreatePresetOrg(ctx context.Context, r repository.Repository) error {

	// In Instill Core, we provide a "preset" namespace for storing preset
	// resources, such as preset pipelines.
	presetOrg := &datamodel.Owner{
		Base:        datamodel.Base{UID: uuid.FromStringOrNil(constant.PresetOrgUID)},
		ID:          constant.PresetOrgID,
		OwnerType:   sql.NullString{String: service.PBUserType2DBUserType[mgmtpb.OwnerType_OWNER_TYPE_ORGANIZATION], Valid: true},
		DisplayName: sql.NullString{String: constant.PresetOrgDisplayName, Valid: true},
	}

	_, err := r.GetOrganization(context.Background(), constant.PresetOrgID, false)
	if err == nil {
		return nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return status.Errorf(codes.Internal, "error %v", err)
	}

	// Create the default preset organization
	err = r.CreateOrganization(ctx, presetOrg)
	if err != nil {
		return err
	}
	return nil
}
