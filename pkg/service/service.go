package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/redis/go-redis/v9"
	"go.einride.tech/aip/filtering"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/instill-ai/mgmt-backend/internal/resource"
	"github.com/instill-ai/mgmt-backend/pkg/constant"
	"github.com/instill-ai/mgmt-backend/pkg/datamodel"
	"github.com/instill-ai/mgmt-backend/pkg/repository"

	mgmtPB "github.com/instill-ai/protogen-go/core/mgmt/v1alpha"
	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1alpha"
)

// Service interface
type Service interface {
	ListRole() []string
	ListUser(ctx context.Context, pageSize int, pageToken string) ([]datamodel.User, string, int64, error)
	CreateUser(ctx context.Context, user *datamodel.User) (*datamodel.User, error)
	GetUser(ctx context.Context, uid uuid.UUID) (*datamodel.User, error)
	GetUserByID(ctx context.Context, id string) (*datamodel.User, error)
	UpdateUser(ctx context.Context, uid uuid.UUID, user *datamodel.User) (*datamodel.User, error)
	DeleteUser(ctx context.Context, uid uuid.UUID) error
	DeleteUserByID(ctx context.Context, id string) error

	GetUserPasswordHash(ctx context.Context, uid uuid.UUID) (string, time.Time, error)
	UpdateUserPasswordHash(ctx context.Context, uid uuid.UUID, newPassword string) error

	ListPipelineTriggerRecords(ctx context.Context, owner *mgmtPB.User, pageSize int64, pageToken string, filter filtering.Filter) ([]*mgmtPB.PipelineTriggerRecord, int64, string, error)
	ListPipelineTriggerTableRecords(ctx context.Context, owner *mgmtPB.User, pageSize int64, pageToken string, filter filtering.Filter) ([]*mgmtPB.PipelineTriggerTableRecord, int64, string, error)
	ListPipelineTriggerChartRecords(ctx context.Context, owner *mgmtPB.User, aggregationWindow int64, filter filtering.Filter) ([]*mgmtPB.PipelineTriggerChartRecord, error)
	ListConnectorExecuteRecords(ctx context.Context, owner *mgmtPB.User, pageSize int64, pageToken string, filter filtering.Filter) ([]*mgmtPB.ConnectorExecuteRecord, int64, string, error)
	ListConnectorExecuteTableRecords(ctx context.Context, owner *mgmtPB.User, pageSize int64, pageToken string, filter filtering.Filter) ([]*mgmtPB.ConnectorExecuteTableRecord, int64, string, error)
	ListConnectorExecuteChartRecords(ctx context.Context, owner *mgmtPB.User, aggregationWindow int64, filter filtering.Filter) ([]*mgmtPB.ConnectorExecuteChartRecord, error)

	GetCtxUser(ctx context.Context) (string, uuid.UUID, error)

	CreateToken(ctx context.Context, token *datamodel.Token) error
	ListTokens(ctx context.Context, pageSize int64, pageToken string, owner string) ([]datamodel.Token, int64, string, error)
	GetToken(ctx context.Context, id string, owner string) (*datamodel.Token, error)
	DeleteToken(ctx context.Context, id string, owner string) error
	ValidateToken(accessToken string) (string, error)
}

type service struct {
	repository                   repository.Repository
	influxDB                     repository.InfluxDB
	connectorPublicServiceClient connectorPB.ConnectorPublicServiceClient
	pipelinePublicServiceClient  pipelinePB.PipelinePublicServiceClient
	redisClient                  *redis.Client
}

// NewService initiates a service instance
func NewService(r repository.Repository, rc *redis.Client, i repository.InfluxDB, c connectorPB.ConnectorPublicServiceClient, p pipelinePB.PipelinePublicServiceClient) Service {
	return &service{
		repository:                   r,
		influxDB:                     i,
		connectorPublicServiceClient: c,
		pipelinePublicServiceClient:  p,
		redisClient:                  rc,
	}
}

// GetUser returns the api user
func (s *service) GetCtxUser(ctx context.Context) (string, uuid.UUID, error) {
	// Verify if "jwt-sub" is in the header
	headerUserUId := resource.GetRequestSingleHeader(ctx, constant.HeaderUserUIDKey)

	if headerUserUId != "" {
		_, err := uuid.FromString(headerUserUId)
		if err != nil {
			return "", uuid.Nil, status.Errorf(codes.Unauthenticated, "Unauthorized")
		}
		user, err := s.GetUser(ctx, uuid.FromStringOrNil(headerUserUId))
		if err != nil {
			return "", uuid.Nil, status.Errorf(codes.Unauthenticated, "Unauthorized")
		}
		return user.ID, uuid.FromStringOrNil(headerUserUId), nil
	}

	return "", uuid.Nil, status.Errorf(codes.Unauthenticated, "Unauthorized")

}

// ListRole lists names of all roles
func (s *service) ListRole() []string {
	return ListAllowedRoleName()
}

// ListUser lists all users
// Return error types
//   - codes.InvalidArgument
//   - codes.Internal
func (s *service) ListUser(ctx context.Context, pageSize int, pageToken string) ([]datamodel.User, string, int64, error) {
	return s.repository.ListUser(ctx, pageSize, pageToken)
}

// CreateUser creates an user instance
// Return error types
//   - codes.InvalidArgument
//   - codes.NotFound
//   - codes.Internal
func (s *service) CreateUser(ctx context.Context, user *datamodel.User) (*datamodel.User, error) {
	//TODO: validate spec JSON schema

	//Validation: role field
	if user.Role.Valid {
		if r := Role(user.Role.String); !ValidateRole(r) {
			return nil, status.Errorf(codes.InvalidArgument, "`role` %s in the body is not valid. Please choose from: [ %v ]", r.GetName(), strings.Join(s.ListRole(), ", "))
		}
	}

	if err := s.repository.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return s.repository.GetUserByID(user.ID)
}

// GetUserByID gets a user by ID
// Return error types
//   - codes.InvalidArgument
//   - codes.NotFound
func (s *service) GetUserByID(ctx context.Context, id string) (*datamodel.User, error) {
	// Validation: Required field
	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "the required field `id` is not specified")
	}

	return s.repository.GetUserByID(id)
}

// GetUser gets a user by UUID
// Return error types
//   - codes.InvalidArgument
//   - codes.NotFound
func (s *service) GetUser(ctx context.Context, uid uuid.UUID) (*datamodel.User, error) {
	// Validation: Required field
	if uid.IsNil() {
		return nil, status.Error(codes.InvalidArgument, "the required field `uid` is not specified")
	}
	return s.repository.GetUser(uid)
}

// UpdateUser updates a user by UUID
// Return error types
//   - codes.InvalidArgument
//   - codes.NotFound
//   - codes.Internal
func (s *service) UpdateUser(ctx context.Context, uid uuid.UUID, user *datamodel.User) (*datamodel.User, error) {
	// Validation: Required field
	if uid.IsNil() {
		return nil, status.Error(codes.InvalidArgument, "the required field `uid` is not specified")
	}

	//Validation: role field
	if user.Role.Valid {
		if r := Role(user.Role.String); !ValidateRole(r) {
			return nil, status.Errorf(codes.InvalidArgument, "`role` %s in the body is not valid. Please choose from: [ %v ]", r.GetName(), strings.Join(s.ListRole(), ", "))
		}
	}

	// Check if the user exists
	if _, err := s.repository.GetUser(uid); err != nil {
		return nil, err
	}

	// Update the user
	if err := s.repository.UpdateUser(ctx, uid, user); err != nil {
		return nil, err
	}

	// Get the updated user
	return s.repository.GetUser(uid)
}

// DeleteUser deletes a user by UUID
// Return error types
//   - codes.InvalidArgument
//   - codes.NotFound
//   - codes.Internal
func (s *service) DeleteUser(ctx context.Context, uid uuid.UUID) error {
	// Validation: Required field
	if uid.IsNil() {
		return status.Error(codes.InvalidArgument, "the required field `uid` is not specified")
	}
	return s.repository.DeleteUser(ctx, uid)
}

// DeleteUserByID deletes a user by ID
// Return error types
//   - codes.InvalidArgument
//   - codes.NotFound
//   - codes.Internal
func (s *service) DeleteUserByID(ctx context.Context, id string) error {
	// Validation: Required field
	if id == "" {
		return status.Error(codes.InvalidArgument, "the required field `id` is not specified")
	}

	return s.repository.DeleteUserByID(ctx, id)
}

func (s *service) GetUserPasswordHash(ctx context.Context, uid uuid.UUID) (string, time.Time, error) {
	return s.repository.GetUserPasswordHash(ctx, uid)
}

func (s *service) UpdateUserPasswordHash(ctx context.Context, uid uuid.UUID, newPassword string) error {

	return s.repository.UpdateUserPasswordHash(ctx, uid, newPassword, time.Now())
}

func (s *service) CreateToken(ctx context.Context, token *datamodel.Token) error {

	err := s.repository.CreateToken(ctx, token)
	if err != nil {
		return err
	}
	// TODO: should be more robust
	s.redisClient.Set(context.Background(), fmt.Sprintf(constant.AccessTokenKeyFormat, token.AccessToken), token.Owner, 0)
	s.redisClient.ExpireAt(context.Background(), fmt.Sprintf(constant.AccessTokenKeyFormat, token.AccessToken), token.ExpireTime)

	return nil
}
func (s *service) ListTokens(ctx context.Context, pageSize int64, pageToken string, owner string) ([]datamodel.Token, int64, string, error) {
	return s.repository.ListTokens(ctx, pageSize, pageToken, owner)

}
func (s *service) GetToken(ctx context.Context, id string, owner string) (*datamodel.Token, error) {
	return s.repository.GetToken(ctx, id, owner)

}
func (s *service) DeleteToken(ctx context.Context, id string, owner string) error {

	token, err := s.GetToken(ctx, id, owner)
	if err != nil {
		return err
	}
	accessToken := token.AccessToken

	// TODO: should be more robust
	s.redisClient.Del(context.Background(), fmt.Sprintf(constant.AccessTokenKeyFormat, accessToken))
	delErr := s.repository.DeleteToken(ctx, id, owner)
	if delErr != nil {
		return delErr
	}

	return nil
}
func (s *service) ValidateToken(accessToken string) (string, error) {
	ownerPermalink, err := s.redisClient.Get(context.Background(), fmt.Sprintf(constant.AccessTokenKeyFormat, accessToken)).Result()
	if err != nil {
		return "", err
	}
	return strings.Split(ownerPermalink, "/")[1], nil
}
