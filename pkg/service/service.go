package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/redis/go-redis/v9"
	"go.einride.tech/aip/filtering"
	"golang.org/x/crypto/bcrypt"

	"github.com/instill-ai/mgmt-backend/internal/resource"
	"github.com/instill-ai/mgmt-backend/pkg/acl"
	"github.com/instill-ai/mgmt-backend/pkg/constant"
	"github.com/instill-ai/mgmt-backend/pkg/datamodel"
	"github.com/instill-ai/mgmt-backend/pkg/repository"

	mgmtpb "github.com/instill-ai/protogen-go/mgmt/v1beta"
	pipelinepb "github.com/instill-ai/protogen-go/pipeline/v1beta"
	errorsx "github.com/instill-ai/x/errors"
)

// Service interface
type Service interface {
	ExtractCtxUser(ctx context.Context, allowVisitor bool) (userUID uuid.UUID, err error)

	CreateAuthenticatedUser(ctx context.Context, ctxUserUID uuid.UUID, user *mgmtpb.AuthenticatedUser) (*mgmtpb.AuthenticatedUser, error)
	GetAuthenticatedUser(ctx context.Context, ctxUserUID uuid.UUID) (*mgmtpb.AuthenticatedUser, error)
	UpdateAuthenticatedUser(ctx context.Context, ctxUserUID uuid.UUID, user *mgmtpb.AuthenticatedUser) (*mgmtpb.AuthenticatedUser, error)

	ListUsers(ctx context.Context, ctxUserUID uuid.UUID, pageSize int, pageToken string, filter filtering.Filter) ([]*mgmtpb.User, int64, string, error)
	GetUser(ctx context.Context, ctxUserUID uuid.UUID, id string) (*mgmtpb.User, error)
	DeleteUser(ctx context.Context, ctxUserUID uuid.UUID, id string) error

	ListUsersAdmin(ctx context.Context, pageSize int, pageToken string, filter filtering.Filter) ([]*mgmtpb.User, int64, string, error)
	ListAuthenticatedUsersAdmin(ctx context.Context, pageSize int, pageToken string, filter filtering.Filter) ([]*mgmtpb.AuthenticatedUser, int64, string, error)
	GetUserAdmin(ctx context.Context, id string) (*mgmtpb.User, error)
	GetUserByUIDAdmin(ctx context.Context, uid uuid.UUID) (*mgmtpb.User, error)
	GetUserUIDByID(ctx context.Context, id string) (uuid.UUID, error)

	CreateToken(ctx context.Context, ctxUserUID uuid.UUID, token *mgmtpb.ApiToken) error
	ListTokens(ctx context.Context, ctxUserUID uuid.UUID, pageSize int64, pageToken string) ([]*mgmtpb.ApiToken, int64, string, error)
	GetToken(ctx context.Context, ctxUserUID uuid.UUID, id string) (*mgmtpb.ApiToken, error)
	DeleteToken(ctx context.Context, ctxUserUID uuid.UUID, id string) error
	ValidateToken(ctx context.Context, accessToken string) (string, error)
	UpdateTokenLastUseTime(ctx context.Context, accessToken string) error

	CheckUserPassword(ctx context.Context, uid uuid.UUID, password string) error
	UpdateUserPassword(ctx context.Context, uid uuid.UUID, newPassword string) error

	ListPipelineTriggerRecords(ctx context.Context, owner *mgmtpb.User, pageSize int64, pageToken string, filter filtering.Filter) ([]*mgmtpb.PipelineTriggerRecord, int64, string, error)
	ListPipelineTriggerTableRecords(ctx context.Context, owner *mgmtpb.User, pageSize int64, pageToken string, filter filtering.Filter) ([]*mgmtpb.PipelineTriggerTableRecord, int64, string, error)
	ListPipelineTriggerChartRecords(_ context.Context, _ *mgmtpb.ListPipelineTriggerChartRecordsRequest, ctxUserUID uuid.UUID) (*mgmtpb.ListPipelineTriggerChartRecordsResponse, error)
	ListPipelineTriggerChartRecordsV0(ctx context.Context, owner *mgmtpb.User, aggregationWindow int64, filter filtering.Filter) ([]*mgmtpb.PipelineTriggerChartRecordV0, error)
	GetPipelineTriggerCount(_ context.Context, _ *mgmtpb.GetPipelineTriggerCountRequest, ctxUserUID uuid.UUID) (*mgmtpb.GetPipelineTriggerCountResponse, error)
	GetModelTriggerCount(_ context.Context, _ *mgmtpb.GetModelTriggerCountRequest, ctxUserUID uuid.UUID) (*mgmtpb.GetModelTriggerCountResponse, error)
	ListModelTriggerChartRecords(ctx context.Context, req *mgmtpb.ListModelTriggerChartRecordsRequest, ctxUserUID uuid.UUID) (*mgmtpb.ListModelTriggerChartRecordsResponse, error)

	DBUser2PBUser(ctx context.Context, dbUser *datamodel.Owner) (*mgmtpb.User, error)
	DBUsers2PBUsers(ctx context.Context, dbUsers []*datamodel.Owner) ([]*mgmtpb.User, error)
	PBAuthenticatedUser2DBUser(ctx context.Context, pbUser *mgmtpb.AuthenticatedUser) (*datamodel.Owner, error)
	DBUsers2PBAuthenticatedUsers(ctx context.Context, dbUsers []*datamodel.Owner) ([]*mgmtpb.AuthenticatedUser, error)

	DBToken2PBToken(ctx context.Context, dbToken *datamodel.Token) (*mgmtpb.ApiToken, error)
	DBTokens2PBTokens(ctx context.Context, dbTokens []*datamodel.Token) ([]*mgmtpb.ApiToken, error)
	PBToken2DBToken(ctx context.Context, pbToken *mgmtpb.ApiToken) (*datamodel.Token, error)

	GetRedisClient() *redis.Client
	GetInfluxClient() repository.InfluxDB
	GetACLClient() *acl.ACLClient

	GrantedNamespaceUID(_ context.Context, namespaceID string, authenticatedUserUID uuid.UUID) (uuid.UUID, error)
}

type service struct {
	repository                  repository.Repository
	influxDB                    repository.InfluxDB
	pipelinePublicServiceClient pipelinepb.PipelinePublicServiceClient
	redisClient                 *redis.Client
	aclClient                   *acl.ACLClient
	instillCoreHost             string
}

// NewService initiates a service instance
func NewService(p pipelinepb.PipelinePublicServiceClient, r repository.Repository, rc *redis.Client, i repository.InfluxDB, acl *acl.ACLClient, h string) Service {
	return &service{
		pipelinePublicServiceClient: p,
		repository:                  r,
		influxDB:                    i,
		redisClient:                 rc,
		aclClient:                   acl,
		instillCoreHost:             h,
	}
}

func (s *service) GetRedisClient() *redis.Client {
	return s.redisClient
}

func (s *service) GetInfluxClient() repository.InfluxDB {
	return s.influxDB
}

func (s *service) GetACLClient() *acl.ACLClient {
	return s.aclClient
}

func (s *service) convertUserIDAlias(ctx context.Context, ctxUserUID uuid.UUID, id string) (string, error) {

	ctx = context.WithValue(ctx, repository.UserUIDCtxKey, ctxUserUID)

	if id == "me" {
		user, err := s.GetUserByUIDAdmin(ctx, ctxUserUID)
		if err != nil {
			return "", errorsx.ErrUnauthenticated
		}
		return user.Id, nil
	}
	return id, nil
}

// GetUser returns the api user
func (s *service) ExtractCtxUser(ctx context.Context, allowVisitor bool) (userUID uuid.UUID, err error) {
	// Verify if "instill-user-uid" is in the header
	authType := resource.GetRequestSingleHeader(ctx, constant.HeaderAuthType)
	if authType == "user" {
		headerCtxUserUID := resource.GetRequestSingleHeader(ctx, constant.HeaderUserUIDKey)
		if headerCtxUserUID == "" {
			return uuid.Nil, errorsx.ErrUnauthenticated
		}
		return uuid.FromStringOrNil(headerCtxUserUID), nil
	} else {
		if !allowVisitor {
			return uuid.Nil, errorsx.ErrUnauthenticated
		}
		headerCtxVisitorUID := resource.GetRequestSingleHeader(ctx, constant.HeaderVisitorUIDKey)
		if headerCtxVisitorUID == "" {
			return uuid.Nil, errorsx.ErrUnauthenticated
		}

		return uuid.FromStringOrNil(headerCtxVisitorUID), nil
	}

}

func (s *service) ListUsers(ctx context.Context, ctxUserUID uuid.UUID, pageSize int, pageToken string, filter filtering.Filter) (users []*mgmtpb.User, totalSize int64, nextPageToken string, err error) {
	ctx = context.WithValue(ctx, repository.UserUIDCtxKey, ctxUserUID)
	dbUsers, totalSize, nextPageToken, err := s.repository.ListUsers(ctx, pageSize, pageToken, filter)
	if err != nil {
		return nil, 0, "", fmt.Errorf("users/ with page_size=%d page_token=%s: %w", pageSize, pageToken, err)
	}
	users, err = s.DBUsers2PBUsers(ctx, dbUsers)
	return users, totalSize, nextPageToken, err
}

func (s *service) CreateAuthenticatedUser(ctx context.Context, ctxUserUID uuid.UUID, user *mgmtpb.AuthenticatedUser) (*mgmtpb.AuthenticatedUser, error) {
	ctx = context.WithValue(ctx, repository.UserUIDCtxKey, ctxUserUID)
	dbUser, err := s.PBAuthenticatedUser2DBUser(ctx, user)
	if err != nil {
		return nil, err
	}

	if err := s.repository.CreateUser(ctx, dbUser); err != nil {
		return nil, err
	}

	return s.GetAuthenticatedUser(ctx, ctxUserUID)
}

func (s *service) GetUser(ctx context.Context, ctxUserUID uuid.UUID, id string) (*mgmtpb.User, error) {
	ctx = context.WithValue(ctx, repository.UserUIDCtxKey, ctxUserUID)
	id, err := s.convertUserIDAlias(ctx, ctxUserUID, id)
	if err != nil {
		return nil, err
	}

	if pbUser := s.getUserFromCacheByID(ctx, id); pbUser != nil {
		return pbUser, nil
	}

	dbUser, err := s.repository.GetUser(ctx, id, false)
	if err != nil {
		return nil, fmt.Errorf("users/%s: %w", id, err)
	}

	pbUser, err := s.DBUser2PBUser(ctx, dbUser)
	if err != nil {
		return nil, err
	}

	err = s.setUserToCache(ctx, pbUser)
	if err != nil {
		return nil, err
	}

	return pbUser, nil
}

func (s *service) GetUserAdmin(ctx context.Context, id string) (*mgmtpb.User, error) {

	if pbUser := s.getUserFromCacheByID(ctx, id); pbUser != nil {
		return pbUser, nil
	}
	dbUser, err := s.repository.GetUser(ctx, id, false)
	if err != nil {
		return nil, fmt.Errorf("users/%s: %w", id, err)
	}
	pbUser, err := s.DBUser2PBUser(ctx, dbUser)
	if err != nil {
		return nil, err
	}
	err = s.setUserToCache(ctx, pbUser)
	if err != nil {
		return nil, err
	}
	return pbUser, nil
}

func (s *service) GetUserByUIDAdmin(ctx context.Context, uid uuid.UUID) (*mgmtpb.User, error) {

	if pbUser := s.getUserFromCacheByUID(ctx, uid); pbUser != nil {
		return pbUser, nil
	}
	dbUser, err := s.repository.GetUserByUID(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("users/%s: %w", uid, err)
	}
	pbUser, err := s.DBUser2PBUser(ctx, dbUser)
	if err != nil {
		return nil, err
	}
	err = s.setUserToCache(ctx, pbUser)
	if err != nil {
		return nil, err
	}
	return pbUser, nil
}

// GetUserUIDByID returns the internal UUID for a user given their public ID.
// This is used internally when the UID is needed but the public protobuf doesn't contain it.
func (s *service) GetUserUIDByID(ctx context.Context, id string) (uuid.UUID, error) {
	dbUser, err := s.repository.GetUser(ctx, id, false)
	if err != nil {
		return uuid.Nil, fmt.Errorf("users/%s: %w", id, err)
	}
	return dbUser.UID, nil
}

func (s *service) ListUsersAdmin(ctx context.Context, pageSize int, pageToken string, filter filtering.Filter) ([]*mgmtpb.User, int64, string, error) {
	dbUsers, totalSize, nextPageToken, err := s.repository.ListUsers(ctx, pageSize, pageToken, filter)
	if err != nil {
		return nil, 0, "", fmt.Errorf("users/ with page_size=%d page_token=%s: %w", pageSize, pageToken, err)
	}
	pbUsers, err := s.DBUsers2PBUsers(ctx, dbUsers)
	return pbUsers, totalSize, nextPageToken, err
}

func (s *service) ListAuthenticatedUsersAdmin(ctx context.Context, pageSize int, pageToken string, filter filtering.Filter) ([]*mgmtpb.AuthenticatedUser, int64, string, error) {
	dbUsers, totalSize, nextPageToken, err := s.repository.ListUsers(ctx, pageSize, pageToken, filter)
	if err != nil {
		return nil, 0, "", fmt.Errorf("users/ with page_size=%d page_token=%s: %w", pageSize, pageToken, err)
	}
	pbUsers, err := s.DBUsers2PBAuthenticatedUsers(ctx, dbUsers)
	return pbUsers, totalSize, nextPageToken, err
}

func (s *service) GetAuthenticatedUser(ctx context.Context, ctxUserUID uuid.UUID) (*mgmtpb.AuthenticatedUser, error) {
	ctx = context.WithValue(ctx, repository.UserUIDCtxKey, ctxUserUID)
	dbUser, err := s.repository.GetUserByUID(ctx, ctxUserUID)
	if err != nil {
		return nil, err
	}

	pbUser, err := s.DBUser2PBAuthenticatedUser(ctx, dbUser)
	if err != nil {
		return nil, err
	}

	return pbUser, nil
}

// UpdateUser updates a user by UUID
func (s *service) UpdateAuthenticatedUser(ctx context.Context, ctxUserUID uuid.UUID, user *mgmtpb.AuthenticatedUser) (*mgmtpb.AuthenticatedUser, error) {
	ctx = context.WithValue(ctx, repository.UserUIDCtxKey, ctxUserUID)

	// Check if the user exists
	if _, err := s.repository.GetUserByUID(ctx, ctxUserUID); err != nil {
		return nil, err
	}

	// Update the user
	dbUser, err := s.PBAuthenticatedUser2DBUser(ctx, user)
	if err != nil {
		return nil, err
	}

	err = s.deleteUserFromCacheByID(ctx, dbUser.ID)
	if err != nil {
		return nil, err
	}
	if err := s.repository.UpdateUser(ctx, dbUser.ID, dbUser); err != nil {
		return nil, fmt.Errorf("users/%s: %w", dbUser.ID, err)
	}

	return s.GetAuthenticatedUser(ctx, ctxUserUID)

}

// DeleteUser deletes a user by ID
func (s *service) DeleteUser(ctx context.Context, ctxUserUID uuid.UUID, id string) error {
	ctx = context.WithValue(ctx, repository.UserUIDCtxKey, ctxUserUID)

	err := s.deleteUserFromCacheByID(ctx, id)
	if err != nil {
		return err
	}
	err = s.repository.DeleteUser(ctx, id)
	if err != nil {
		return fmt.Errorf("users/%s: %w", id, err)
	}
	return nil
}

func (s *service) CheckUserPassword(ctx context.Context, uid uuid.UUID, password string) error {

	ctx = context.WithValue(ctx, repository.UserUIDCtxKey, uid)

	var err error
	passwordHash := s.getUserPasswordHashFromCache(ctx, uid)
	if passwordHash == "" {
		passwordHash, _, err = s.repository.GetUserPasswordHash(ctx, uid)
		if err != nil {
			return err
		}
		_ = s.setUserPasswordHashToCache(ctx, uid, passwordHash)
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		return errorsx.ErrPasswordNotMatch
	}
	return nil
}

func (s *service) UpdateUserPassword(ctx context.Context, uid uuid.UUID, newPassword string) error {

	ctx = context.WithValue(ctx, repository.UserUIDCtxKey, uid)

	passwordBytes, err := bcrypt.GenerateFromPassword([]byte(newPassword), 10)
	if err != nil {
		return err
	}
	_ = s.deleteUserPasswordHashFromCache(ctx, uid)
	return s.repository.UpdateUserPasswordHash(ctx, uid, string(passwordBytes), time.Now())
}

func (s *service) CreateToken(ctx context.Context, ctxUserUID uuid.UUID, token *mgmtpb.ApiToken) error {

	ctx = context.WithValue(ctx, repository.UserUIDCtxKey, ctxUserUID)

	dbToken, err := s.PBToken2DBToken(ctx, token)
	if err != nil {
		return err
	}

	dbToken.AccessToken = datamodel.GenerateToken()
	dbToken.Owner = fmt.Sprintf("users/%s", ctxUserUID)
	curTime := time.Now()
	dbToken.CreateTime = curTime
	dbToken.UpdateTime = curTime
	dbToken.State = datamodel.TokenState(mgmtpb.ApiToken_STATE_ACTIVE)

	switch token.GetExpiration().(type) {
	case *mgmtpb.ApiToken_Ttl:
		if token.GetTtl() >= 0 {
			dbToken.ExpireTime = curTime.Add(time.Second * time.Duration(token.GetTtl()))
		} else if token.GetTtl() == -1 {
			dbToken.ExpireTime = time.Date(2099, 12, 31, 0, 0, 0, 0, time.Now().UTC().Location())
		} else {
			return errorsx.ErrInvalidTokenTTL
		}
	case *mgmtpb.ApiToken_ExpireTime:
		dbToken.ExpireTime = token.GetExpireTime().AsTime()
	}

	dbToken.TokenType = constant.DefaultTokenType

	err = s.repository.CreateToken(ctx, dbToken)
	if err != nil {
		return err
	}

	_ = s.setAPITokenToCache(ctx, dbToken.AccessToken, ctxUserUID, dbToken.ExpireTime)

	return nil
}
func (s *service) ListTokens(ctx context.Context, ctxUserUID uuid.UUID, pageSize int64, pageToken string) ([]*mgmtpb.ApiToken, int64, string, error) {

	ctx = context.WithValue(ctx, repository.UserUIDCtxKey, ctxUserUID)

	ownerPermlink := fmt.Sprintf("users/%s", ctxUserUID.String())
	dbTokens, pageSize, pageToken, err := s.repository.ListTokens(ctx, ownerPermlink, pageSize, pageToken)
	if err != nil {
		return nil, 0, "", fmt.Errorf("tokens/ with page_size=%d page_token=%s: %w", pageSize, pageToken, err)
	}

	pbTokens, err := s.DBTokens2PBTokens(ctx, dbTokens)
	return pbTokens, pageSize, pageToken, err

}
func (s *service) GetToken(ctx context.Context, ctxUserUID uuid.UUID, id string) (*mgmtpb.ApiToken, error) {

	ctx = context.WithValue(ctx, repository.UserUIDCtxKey, ctxUserUID)

	ownerPermlink := fmt.Sprintf("users/%s", ctxUserUID.String())
	dbToken, err := s.repository.GetToken(ctx, ownerPermlink, id)
	if err != nil {
		return nil, fmt.Errorf("tokens/%s: %w", id, err)
	}

	return s.DBToken2PBToken(ctx, dbToken)

}
func (s *service) DeleteToken(ctx context.Context, ctxUserUID uuid.UUID, id string) error {

	ctx = context.WithValue(ctx, repository.UserUIDCtxKey, ctxUserUID)

	ownerPermlink := fmt.Sprintf("users/%s", ctxUserUID.String())
	token, err := s.repository.GetToken(ctx, ownerPermlink, id)
	if err != nil {
		return fmt.Errorf("tokens/%s: %w", id, err)
	}
	accessToken := token.AccessToken

	// TODO: should be more robust
	_ = s.deleteAPITokenFromCache(ctx, accessToken)
	err = s.repository.DeleteToken(ctx, ownerPermlink, id)
	if err != nil {
		return fmt.Errorf("tokens/%s: %w", id, err)
	}
	return nil
}

func (s *service) UpdateTokenLastUseTime(ctx context.Context, accessToken string) error {
	return s.repository.UpdateTokenLastUseTime(ctx, accessToken)
}

func (s *service) ValidateToken(ctx context.Context, accessToken string) (string, error) {
	uid := s.getAPITokenFromCache(ctx, accessToken)
	if uid == uuid.Nil {
		dbToken, err := s.repository.LookupToken(ctx, accessToken)
		if err != nil {
			return "", errorsx.ErrUnauthenticated
		}
		uid = uuid.FromStringOrNil(strings.Split(dbToken.Owner, "/")[1])
		_ = s.setAPITokenToCache(ctx, dbToken.AccessToken, uid, dbToken.ExpireTime)

	}
	if uid == uuid.Nil {
		return "", errorsx.ErrUnauthenticated
	}
	return uid.String(), nil
}
