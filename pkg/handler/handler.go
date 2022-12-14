package handler

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/iancoleman/strcase"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/instill-ai/mgmt-backend/config"
	"github.com/instill-ai/mgmt-backend/internal/logger"
	"github.com/instill-ai/mgmt-backend/pkg/datamodel"
	"github.com/instill-ai/mgmt-backend/pkg/service"
	"github.com/instill-ai/mgmt-backend/pkg/usage"

	"github.com/instill-ai/x/sterr"

	fieldmask_utils "github.com/mennanov/fieldmask-utils"

	healthcheckPB "github.com/instill-ai/protogen-go/vdp/healthcheck/v1alpha"
	mgmtPB "github.com/instill-ai/protogen-go/vdp/mgmt/v1alpha"
	checkfield "github.com/instill-ai/x/checkfield"
)

// TODO: Validate mask based on the field behavior.
// Currently, the OUTPUT_ONLY fields are hard-coded.
var createRequiredFields = []string{"email"}
var outputOnlyFields = []string{"name", "uid", "type", "create_time", "update_time"}
var immutableFields = []string{"id"}

const defaultPageSize = int64(10)
const maxPageSize = int64(100)

type handler struct {
	mgmtPB.UnimplementedUserServiceServer
	service service.Service
	usg     usage.Usage
}

// NewHandler initiates a handler instance
func NewHandler(s service.Service, u usage.Usage) mgmtPB.UserServiceServer {
	return &handler{
		service: s,
		usg:     u,
	}
}

// Liveness checks the liveness of the server
func (h *handler) Liveness(ctx context.Context, in *mgmtPB.LivenessRequest) (*mgmtPB.LivenessResponse, error) {
	return &mgmtPB.LivenessResponse{
		HealthCheckResponse: &healthcheckPB.HealthCheckResponse{
			Status: healthcheckPB.HealthCheckResponse_SERVING_STATUS_SERVING,
		},
	}, nil
}

// Readiness checks the readiness of the server
func (h *handler) Readiness(ctx context.Context, in *mgmtPB.ReadinessRequest) (*mgmtPB.ReadinessResponse, error) {
	return &mgmtPB.ReadinessResponse{
		HealthCheckResponse: &healthcheckPB.HealthCheckResponse{
			Status: healthcheckPB.HealthCheckResponse_SERVING_STATUS_SERVING,
		},
	}, nil
}

// ListUser lists all users
func (h *handler) ListUser(ctx context.Context, req *mgmtPB.ListUserRequest) (*mgmtPB.ListUserResponse, error) {
	logger, _ := logger.GetZapLogger()

	pageSize := req.GetPageSize()
	if pageSize == 0 {
		pageSize = defaultPageSize
	} else if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	dbUsers, nextPageToken, totalSize, err := h.service.ListUser(int(pageSize), req.GetPageToken())
	if err != nil {
		sta := status.Convert(err)
		switch sta.Code() {
		case codes.InvalidArgument:
			st, e := sterr.CreateErrorBadRequest(
				"list user error", []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "ListUserRequest.page_token",
						Description: sta.Message(),
					},
				})
			if e != nil {
				logger.Error(e.Error())
			}
			return &mgmtPB.ListUserResponse{}, st.Err()
		default:
			st, e := sterr.CreateErrorResourceInfoStatus(
				sta.Code(),
				"list user error",
				"user",
				"",
				"",
				sta.Message(),
			)
			if e != nil {
				logger.Error(e.Error())
			}
			return &mgmtPB.ListUserResponse{}, st.Err()
		}
	}

	pbUsers := []*mgmtPB.User{}
	for _, dbUser := range dbUsers {
		pbUser, err := datamodel.DBUser2PBUser(&dbUser)
		if err != nil {
			logger.Error(err.Error())
			st, e := sterr.CreateErrorResourceInfoStatus(
				codes.Internal,
				"list user error",
				"user",
				fmt.Sprintf("id %s", dbUser.ID),
				"",
				err.Error(),
			)
			if e != nil {
				logger.Error(e.Error())
			}
			return &mgmtPB.ListUserResponse{}, st.Err()
		}
		pbUsers = append(pbUsers, pbUser)
	}

	resp := mgmtPB.ListUserResponse{
		Users: pbUsers,
		NextPageToken: nextPageToken,
		TotalSize: totalSize,
	}
	return &resp, nil
}

// CreateUser creates a user. This endpoint is not supported.
func (h *handler) CreateUser(ctx context.Context, req *mgmtPB.CreateUserRequest) (*mgmtPB.CreateUserResponse, error) {
	logger, _ := logger.GetZapLogger()
	resp := &mgmtPB.CreateUserResponse{}

	// Return error if REQUIRED fields are not provided in the requested payload pipeline resource
	if err := checkfield.CheckRequiredFields(req.GetUser(), append(createRequiredFields, immutableFields...)); err != nil {
		st, e := sterr.CreateErrorBadRequest(
			"create user bad request error", []*errdetails.BadRequest_FieldViolation{
				{
					Field:       fmt.Sprintf("%v, %v", createRequiredFields, immutableFields),
					Description: err.Error(),
				},
			},
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return resp, st.Err()
	}

	// Validate the user id conforms to RFC-1034, which restricts to letters, numbers,
	// and hyphen, with the first character a letter, the last a letter or a
	// number, and a 63 character maximum.
	id := req.GetUser().GetId()
	err := checkfield.CheckResourceID(id)
	if err != nil {
		st, e := sterr.CreateErrorBadRequest(
			"create user bad request error", []*errdetails.BadRequest_FieldViolation{
				{
					Field:       "id",
					Description: err.Error(),
				},
			},
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return resp, st.Err()
	}

	st, err := sterr.CreateErrorResourceInfoStatus(
		codes.Unimplemented,
		"create user not implemented error",
		"endpoint",
		"/users",
		"",
		"not implemented",
	)
	if err != nil {
		logger.Error(err.Error())
	}
	return resp, st.Err()
}

// GetUser gets a user
func (h *handler) GetUser(ctx context.Context, req *mgmtPB.GetUserRequest) (*mgmtPB.GetUserResponse, error) {
	logger, _ := logger.GetZapLogger()

	id := strings.TrimPrefix(req.GetName(), "users/")

	dbUser, err := h.service.GetUserByID(id)
	if err != nil {
		sta := status.Convert(err)
		switch sta.Code() {
		case codes.InvalidArgument:
			st, e := sterr.CreateErrorBadRequest(
				"get user error", []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "GetUserRequest.name",
						Description: sta.Message(),
					},
				})
			if e != nil {
				logger.Error(e.Error())
			}
			return &mgmtPB.GetUserResponse{}, st.Err()
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
			return &mgmtPB.GetUserResponse{}, st.Err()
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
		return &mgmtPB.GetUserResponse{}, st.Err()
	}

	resp := mgmtPB.GetUserResponse{
		User: pbUser,
	}
	return &resp, nil
}

// LookUpUser gets a user by permalink
func (h *handler) LookUpUser(ctx context.Context, req *mgmtPB.LookUpUserRequest) (*mgmtPB.LookUpUserResponse, error) {
	logger, _ := logger.GetZapLogger()

	uidStr := strings.TrimPrefix(req.GetPermalink(), "users/")
	// Validation: `uid` in request is valid
	uid, err := uuid.FromString(uidStr)
	if err != nil {
		st, e := sterr.CreateErrorBadRequest(
			"look up user invalid uuid error", []*errdetails.BadRequest_FieldViolation{
				{
					Field:       "LookUpUserRequest.permalink",
					Description: err.Error(),
				},
			},
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return &mgmtPB.LookUpUserResponse{}, st.Err()
	}

	dbUser, err := h.service.GetUser(uid)
	if err != nil {
		sta := status.Convert(err)
		switch sta.Code() {
		case codes.InvalidArgument:
			st, e := sterr.CreateErrorBadRequest(
				"look up user error", []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "LookUpUserRequest.permalink",
						Description: sta.Message(),
					},
				})
			if e != nil {
				logger.Error(e.Error())
			}
			return &mgmtPB.LookUpUserResponse{}, st.Err()
		default:
			st, e := sterr.CreateErrorResourceInfoStatus(
				sta.Code(),
				"look up user error",
				"user",
				fmt.Sprintf("uid %s", uid),
				"",
				sta.Message(),
			)
			if e != nil {
				logger.Error(e.Error())
			}
			return &mgmtPB.LookUpUserResponse{}, st.Err()
		}
	}

	pbUser, err := datamodel.DBUser2PBUser(dbUser)
	if err != nil {
		logger.Error(err.Error())
		st, e := sterr.CreateErrorResourceInfoStatus(
			codes.Internal,
			"look up user error",
			"user",
			fmt.Sprintf("uid %s", dbUser.UID),
			"",
			err.Error(),
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return &mgmtPB.LookUpUserResponse{}, st.Err()
	}
	resp := mgmtPB.LookUpUserResponse{
		User: pbUser,
	}
	return &resp, nil
}

// UpdateUser updates an existing user
func (h *handler) UpdateUser(ctx context.Context, req *mgmtPB.UpdateUserRequest) (*mgmtPB.UpdateUserResponse, error) {
	logger, _ := logger.GetZapLogger()

	reqUser := req.GetUser()

	// Validate the field mask
	if !req.GetUpdateMask().IsValid(reqUser) {
		st, e := sterr.CreateErrorBadRequest(
			"update user invalid fieidmask error", []*errdetails.BadRequest_FieldViolation{
				{
					Field:       "UpdateUserRequest.update_mask",
					Description: "invalid",
				},
			},
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return &mgmtPB.UpdateUserResponse{}, st.Err()
	}

	reqFieldMask, err := checkfield.CheckUpdateOutputOnlyFields(req.GetUpdateMask(), outputOnlyFields)
	if err != nil {
		st, e := sterr.CreateErrorBadRequest(
			"update user update OUTPUT_ONLY fields error", []*errdetails.BadRequest_FieldViolation{
				{
					Field:       "UpdateUserRequest OUTPUT_ONLY fields",
					Description: err.Error(),
				},
			},
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return &mgmtPB.UpdateUserResponse{}, st.Err()
	}

	mask, err := fieldmask_utils.MaskFromProtoFieldMask(reqFieldMask, strcase.ToCamel)
	if err != nil {
		logger.Error(err.Error())
		st, e := sterr.CreateErrorBadRequest(
			"update user update mask error", []*errdetails.BadRequest_FieldViolation{
				{
					Field:       "UpdateUserRequest.update_mask",
					Description: err.Error(),
				},
			},
		)
		if e != nil {
			logger.Error(e.Error())
		}

		return &mgmtPB.UpdateUserResponse{}, st.Err()
	}

	// Get current user
	GResp, err := h.GetUser(ctx, &mgmtPB.GetUserRequest{Name: reqUser.GetName()})
	if err != nil {
		return &mgmtPB.UpdateUserResponse{}, err
	}
	pbUserToUpdate := GResp.GetUser()

	if mask.IsEmpty() {
		// return the un-changed user `pbUserToUpdate`
		resp := mgmtPB.UpdateUserResponse{
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
			"update user error",
			"user",
			fmt.Sprintf("user %v", pbUserToUpdate),
			"",
			err.Error(),
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return &mgmtPB.UpdateUserResponse{}, st.Err()
	}

	// Handle immutable fields from the update mask
	err = checkfield.CheckUpdateImmutableFields(reqUser, pbUserToUpdate, immutableFields)
	if err != nil {
		st, e := sterr.CreateErrorBadRequest(
			"update user update IMMUTABLE fields error", []*errdetails.BadRequest_FieldViolation{
				{
					Field:       "UpdateUserRequest IMMUTABLE fields",
					Description: err.Error(),
				},
			},
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return &mgmtPB.UpdateUserResponse{}, st.Err()
	}

	// Only the fields mentioned in the field mask will be copied to `pbUserToUpdate`, other fields are left intact
	err = fieldmask_utils.StructToStruct(mask, reqUser, pbUserToUpdate)
	if err != nil {
		logger.Error(err.Error())
		st, e := sterr.CreateErrorResourceInfoStatus(
			codes.Internal,
			"update user error", "user", fmt.Sprintf("uid %s", reqUser.Uid),
			"",
			err.Error(),
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return &mgmtPB.UpdateUserResponse{}, st.Err()
	}

	dbUserToUpd, err := datamodel.PBUser2DBUser(pbUserToUpdate)
	if err != nil {
		logger.Error(err.Error())
		st, e := sterr.CreateErrorResourceInfoStatus(
			codes.Internal,
			"update user error",
			"user",
			fmt.Sprintf("id %s", pbUserToUpdate.GetId()),
			"",
			err.Error(),
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return &mgmtPB.UpdateUserResponse{}, st.Err()
	}

	dbUserUpdated, err := h.service.UpdateUser(uid, dbUserToUpd)
	if err != nil {
		sta := status.Convert(err)
		switch sta.Code() {
		case codes.InvalidArgument:
			st, e := sterr.CreateErrorBadRequest(
				"update user error", []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "UpdateUserRequest",
						Description: sta.Message(),
					},
				})
			if e != nil {
				logger.Error(e.Error())
			}
			return &mgmtPB.UpdateUserResponse{}, st.Err()
		default:
			st, e := sterr.CreateErrorResourceInfoStatus(
				sta.Code(),
				"update user error",
				"user",
				fmt.Sprintf("uid %s", uid.String()),
				"",
				sta.Message(),
			)
			if e != nil {
				logger.Error(e.Error())
			}
			return &mgmtPB.UpdateUserResponse{}, st.Err()
		}
	}

	pbUserUpdated, err := datamodel.DBUser2PBUser(dbUserUpdated)
	if err != nil {
		logger.Error(err.Error())
		st, e := sterr.CreateErrorResourceInfoStatus(
			codes.Internal,
			"get user error",
			"user",
			fmt.Sprintf("uid %s", dbUserUpdated.UID),
			"",
			err.Error(),
		)
		if e != nil {
			logger.Error(e.Error())
		}
		return &mgmtPB.UpdateUserResponse{}, st.Err()
	}
	resp := mgmtPB.UpdateUserResponse{
		User: pbUserUpdated,
	}

	// Trigger single reporter
	if !config.Config.Server.DisableUsage && h.usg != nil {
		h.usg.TriggerSingleReporter(ctx)
	}

	return &resp, nil
}

// DeleteUser deletes a user. This endpoint is not supported.
func (h *handler) DeleteUser(ctx context.Context, req *mgmtPB.DeleteUserRequest) (*mgmtPB.DeleteUserResponse, error) {
	logger, _ := logger.GetZapLogger()

	st, err := sterr.CreateErrorResourceInfoStatus(
		codes.Unimplemented,
		"delete user not implemented error",
		"endpoint",
		"/users/{user}",
		"",
		"not implemented",
	)
	if err != nil {
		logger.Error(err.Error())
	}
	return &mgmtPB.DeleteUserResponse{}, st.Err()
}
