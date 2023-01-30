package handler

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/iancoleman/strcase"
	"github.com/instill-ai/mgmt-backend/config"
	"github.com/instill-ai/mgmt-backend/pkg/datamodel"
	"github.com/instill-ai/mgmt-backend/pkg/logger"
	"github.com/instill-ai/mgmt-backend/pkg/service"
	"github.com/instill-ai/mgmt-backend/pkg/usage"
	"github.com/instill-ai/x/sterr"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	healthcheckPB "github.com/instill-ai/protogen-go/vdp/healthcheck/v1alpha"
	mgmtPB "github.com/instill-ai/protogen-go/vdp/mgmt/v1alpha"
	checkfield "github.com/instill-ai/x/checkfield"
	fieldmask_utils "github.com/mennanov/fieldmask-utils"
)

// TODO: Validate mask based on the field behavior. Currently, the fields are hard-coded.
// We stipulate that the ID of the user is IMMUTABLE
var createRequiredFields = []string{"id", "email", "newsletter_subscription"}
var outputOnlyFields = []string{"name", "type", "create_time", "update_time"}
var immutableFields = []string{"uid", "id"}

type publicHandler struct {
	mgmtPB.UnimplementedUserPublicServiceServer
	service service.Service
	usg     usage.Usage
}

// NewPublicHandler initiates a public handler instance
func NewPublicHandler(s service.Service, u usage.Usage) mgmtPB.UserPublicServiceServer {
	return &publicHandler{
		service: s,
		usg:     u,
	}
}

// Liveness checks the liveness of the server
func (h *publicHandler) Liveness(ctx context.Context, in *mgmtPB.LivenessRequest) (*mgmtPB.LivenessResponse, error) {
	return &mgmtPB.LivenessResponse{
		HealthCheckResponse: &healthcheckPB.HealthCheckResponse{
			Status: healthcheckPB.HealthCheckResponse_SERVING_STATUS_SERVING,
		},
	}, nil
}

// Readiness checks the readiness of the server
func (h *publicHandler) Readiness(ctx context.Context, in *mgmtPB.ReadinessRequest) (*mgmtPB.ReadinessResponse, error) {
	return &mgmtPB.ReadinessResponse{
		HealthCheckResponse: &healthcheckPB.HealthCheckResponse{
			Status: healthcheckPB.HealthCheckResponse_SERVING_STATUS_SERVING,
		},
	}, nil
}

// ========== Public API

// GetAuthenticatedUser gets the authenticated user.
// Note: this endpoint is hard-coded, assuming the ID of the authenticated user is the default user.
func (h *publicHandler) GetAuthenticatedUser(ctx context.Context, req *mgmtPB.GetAuthenticatedUserRequest) (*mgmtPB.GetAuthenticatedUserResponse, error) {
	logger, _ := logger.GetZapLogger()

	id := config.DefaultUserID

	dbUser, err := h.service.GetUserByID(id)
	if err != nil {
		sta := status.Convert(err)
		switch sta.Code() {
		case codes.InvalidArgument:
			st, e := sterr.CreateErrorBadRequest(
				"get user error", []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "GetAuthenticatedUser",
						Description: sta.Message(),
					},
				})
			if e != nil {
				logger.Error(e.Error())
			}
			return &mgmtPB.GetAuthenticatedUserResponse{}, st.Err()
		default:
			st, e := sterr.CreateErrorResourceInfoStatus(
				sta.Code(),
				"get user error",
				"user",
				fmt.Sprintf("id %s", id),
				"",
				sta.Message(),
			)
			if e != nil {
				logger.Error(e.Error())
			}
			return &mgmtPB.GetAuthenticatedUserResponse{}, st.Err()
		}
	}

	pbUser, err := datamodel.DBUser2PBUser(dbUser)
	if err != nil {
		logger.Error(err.Error())
		st, e := sterr.CreateErrorResourceInfoStatus(
			codes.Internal,
			"get user error",
			"user",
			fmt.Sprintf("id %s", dbUser.ID),
			"",
			err.Error(),
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return &mgmtPB.GetAuthenticatedUserResponse{}, st.Err()
	}

	resp := mgmtPB.GetAuthenticatedUserResponse{
		User: pbUser,
	}
	return &resp, nil

}

// UpdateAuthenticatedUser updates the authenticated user.
// Note: this endpoint is hard-coded, assuming the ID of the authenticated user is the default user.
func (h *publicHandler) UpdateAuthenticatedUser(ctx context.Context, req *mgmtPB.UpdateAuthenticatedUserRequest) (*mgmtPB.UpdateAuthenticatedUserResponse, error) {
	logger, _ := logger.GetZapLogger()

	reqUser := req.GetUser()

	// Validate the field mask
	if !req.GetUpdateMask().IsValid(reqUser) {
		st, e := sterr.CreateErrorBadRequest(
			"update user invalid fieidmask error", []*errdetails.BadRequest_FieldViolation{
				{
					Field:       "UpdateAuthenticatedUserRequest.update_mask",
					Description: "invalid",
				},
			},
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return &mgmtPB.UpdateAuthenticatedUserResponse{}, st.Err()
	}

	reqFieldMask, err := checkfield.CheckUpdateOutputOnlyFields(req.GetUpdateMask(), outputOnlyFields)
	if err != nil {
		st, e := sterr.CreateErrorBadRequest(
			"update user update OUTPUT_ONLY fields error", []*errdetails.BadRequest_FieldViolation{
				{
					Field:       "UpdateAuthenticatedUserRequest OUTPUT_ONLY fields",
					Description: err.Error(),
				},
			},
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return &mgmtPB.UpdateAuthenticatedUserResponse{}, st.Err()
	}

	mask, err := fieldmask_utils.MaskFromProtoFieldMask(reqFieldMask, strcase.ToCamel)
	if err != nil {
		logger.Error(err.Error())
		st, e := sterr.CreateErrorBadRequest(
			"update user update mask error", []*errdetails.BadRequest_FieldViolation{
				{
					Field:       "UpdateAuthenticatedUserRequest.update_mask",
					Description: err.Error(),
				},
			},
		)
		if e != nil {
			logger.Error(e.Error())
		}

		return &mgmtPB.UpdateAuthenticatedUserResponse{}, st.Err()
	}

	// Get current authenticated user
	GResp, err := h.GetAuthenticatedUser(ctx, &mgmtPB.GetAuthenticatedUserRequest{})

	if err != nil {
		return &mgmtPB.UpdateAuthenticatedUserResponse{}, err
	}
	pbUserToUpdate := GResp.GetUser()

	if mask.IsEmpty() {
		// return the un-changed user `pbUserToUpdate`
		resp := mgmtPB.UpdateAuthenticatedUserResponse{
			User: pbUserToUpdate,
		}
		return &resp, nil
	}

	// the current user `pbUserToUpdate`: a struct to copy to
	uid, err := uuid.FromString(pbUserToUpdate.GetUid())
	if err != nil {
		logger.Error(err.Error())
		st, e := sterr.CreateErrorResourceInfoStatus(
			codes.Internal,
			"update authenticated user error",
			"user",
			fmt.Sprintf("user %v", pbUserToUpdate),
			"",
			err.Error(),
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return &mgmtPB.UpdateAuthenticatedUserResponse{}, st.Err()
	}

	// Handle immutable fields from the update mask
	err = checkfield.CheckUpdateImmutableFields(reqUser, pbUserToUpdate, immutableFields)
	if err != nil {
		st, e := sterr.CreateErrorBadRequest(
			"update authenticated user update IMMUTABLE fields error", []*errdetails.BadRequest_FieldViolation{
				{
					Field:       "UpdateAuthenticatedUserRequest IMMUTABLE fields",
					Description: err.Error(),
				},
			},
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return &mgmtPB.UpdateAuthenticatedUserResponse{}, st.Err()
	}

	// Only the fields mentioned in the field mask will be copied to `pbUserToUpdate`, other fields are left intact
	err = fieldmask_utils.StructToStruct(mask, reqUser, pbUserToUpdate)
	if err != nil {
		logger.Error(err.Error())
		st, e := sterr.CreateErrorResourceInfoStatus(
			codes.Internal,
			"update authenticated user error", "user", fmt.Sprintf("uid %s", *reqUser.Uid),
			"",
			err.Error(),
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return &mgmtPB.UpdateAuthenticatedUserResponse{}, st.Err()
	}

	dbUserToUpd, err := datamodel.PBUser2DBUser(pbUserToUpdate)
	if err != nil {
		logger.Error(err.Error())
		st, e := sterr.CreateErrorResourceInfoStatus(
			codes.Internal,
			"update authenticated user error",
			"user",
			fmt.Sprintf("id %s", pbUserToUpdate.GetId()),
			"",
			err.Error(),
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return &mgmtPB.UpdateAuthenticatedUserResponse{}, st.Err()
	}

	dbUserUpdated, err := h.service.UpdateUser(uid, dbUserToUpd)
	if err != nil {
		sta := status.Convert(err)
		switch sta.Code() {
		case codes.InvalidArgument:
			st, e := sterr.CreateErrorBadRequest(
				"update authenticated user error", []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "UpdateAuthenticatedUserRequest",
						Description: sta.Message(),
					},
				})
			if e != nil {
				logger.Error(e.Error())
			}
			return &mgmtPB.UpdateAuthenticatedUserResponse{}, st.Err()
		default:
			st, e := sterr.CreateErrorResourceInfoStatus(
				sta.Code(),
				"update authenticated user error",
				"user",
				fmt.Sprintf("uid %s", uid.String()),
				"",
				sta.Message(),
			)
			if e != nil {
				logger.Error(e.Error())
			}
			return &mgmtPB.UpdateAuthenticatedUserResponse{}, st.Err()
		}
	}

	pbUserUpdated, err := datamodel.DBUser2PBUser(dbUserUpdated)
	if err != nil {
		logger.Error(err.Error())
		st, e := sterr.CreateErrorResourceInfoStatus(
			codes.Internal,
			"get authenticated user error",
			"user",
			fmt.Sprintf("uid %s", dbUserUpdated.UID),
			"",
			err.Error(),
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return &mgmtPB.UpdateAuthenticatedUserResponse{}, st.Err()
	}
	resp := mgmtPB.UpdateAuthenticatedUserResponse{
		User: pbUserUpdated,
	}

	// Trigger single reporter
	if !config.Config.Server.DisableUsage && h.usg != nil {
		h.usg.TriggerSingleReporter(ctx)
	}

	return &resp, nil
}

// ExistUsername verifies if a username (ID) has been occupied
func (h *publicHandler) ExistUsername(ctx context.Context, req *mgmtPB.ExistUsernameRequest) (*mgmtPB.ExistUsernameResponse, error) {
	logger, _ := logger.GetZapLogger()

	id := strings.TrimPrefix(req.GetName(), "users/")

	_, err := h.service.GetUserByID(id)
	if err != nil {
		sta := status.Convert(err)
		switch sta.Code() {
		// user not exist - username not occupied
		case codes.NotFound:
			resp := mgmtPB.ExistUsernameResponse{
				Exists: false,
			}
			return &resp, nil
		// invalid username
		case codes.InvalidArgument:
			st, e := sterr.CreateErrorBadRequest(
				"verify whether username is occupied error", []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "ExistUsernameRequest.name",
						Description: sta.Message(),
					},
				})
			if e != nil {
				logger.Error(e.Error())
			}
			return &mgmtPB.ExistUsernameResponse{}, st.Err()
		default:
			st, e := sterr.CreateErrorResourceInfoStatus(
				sta.Code(),
				"verify whether username is occupied error",
				"user",
				fmt.Sprintf("id %s", id),
				"",
				sta.Message(),
			)
			if e != nil {
				logger.Error(e.Error())
			}
			return &mgmtPB.ExistUsernameResponse{}, st.Err()
		}
	}

	resp := mgmtPB.ExistUsernameResponse{
		Exists: true,
	}
	return &resp, nil
}
