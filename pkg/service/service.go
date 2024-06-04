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

	mgmtPB "github.com/instill-ai/protogen-go/core/mgmt/v1beta"
	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1beta"
)

// Service interface
type Service interface {
	ExtractCtxUser(ctx context.Context, allowVisitor bool) (userUID uuid.UUID, err error)

	CreateAuthenticatedUser(ctx context.Context, ctxUserUID uuid.UUID, user *mgmtPB.AuthenticatedUser) (*mgmtPB.AuthenticatedUser, error)
	GetAuthenticatedUser(ctx context.Context, ctxUserUID uuid.UUID) (*mgmtPB.AuthenticatedUser, error)
	UpdateAuthenticatedUser(ctx context.Context, ctxUserUID uuid.UUID, user *mgmtPB.AuthenticatedUser) (*mgmtPB.AuthenticatedUser, error)

	ListUsers(ctx context.Context, ctxUserUID uuid.UUID, pageSize int, pageToken string, filter filtering.Filter) ([]*mgmtPB.User, int64, string, error)
	GetUser(ctx context.Context, ctxUserUID uuid.UUID, id string) (*mgmtPB.User, error)
	DeleteUser(ctx context.Context, ctxUserUID uuid.UUID, id string) error

	ListUsersAdmin(ctx context.Context, pageSize int, pageToken string, filter filtering.Filter) ([]*mgmtPB.User, int64, string, error)
	ListAuthenticatedUsersAdmin(ctx context.Context, pageSize int, pageToken string, filter filtering.Filter) ([]*mgmtPB.AuthenticatedUser, int64, string, error)
	GetUserAdmin(ctx context.Context, id string) (*mgmtPB.User, error)
	GetUserByUIDAdmin(ctx context.Context, uid uuid.UUID) (*mgmtPB.User, error)

	CreateOrganization(ctx context.Context, ctxUserUID uuid.UUID, org *mgmtPB.Organization) (*mgmtPB.Organization, error)
	ListOrganizations(ctx context.Context, ctxUserUID uuid.UUID, pageSize int, pageToken string, filter filtering.Filter) ([]*mgmtPB.Organization, int64, string, error)
	GetOrganization(ctx context.Context, ctxUserUID uuid.UUID, id string) (*mgmtPB.Organization, error)
	UpdateOrganization(ctx context.Context, ctxUserUID uuid.UUID, id string, org *mgmtPB.Organization) (*mgmtPB.Organization, error)
	DeleteOrganization(ctx context.Context, ctxUserUID uuid.UUID, id string) error

	ListOrganizationsAdmin(ctx context.Context, pageSize int, pageToken string, filter filtering.Filter) ([]*mgmtPB.Organization, int64, string, error)
	GetOrganizationAdmin(ctx context.Context, id string) (*mgmtPB.Organization, error)
	GetOrganizationByUIDAdmin(ctx context.Context, uid uuid.UUID) (*mgmtPB.Organization, error)

	ListUserMemberships(ctx context.Context, ctxUserUID uuid.UUID, userID string) ([]*mgmtPB.UserMembership, error)
	GetUserMembership(ctx context.Context, ctxUserUID uuid.UUID, userID string, orgID string) (*mgmtPB.UserMembership, error)
	UpdateUserMembership(ctx context.Context, ctxUserUID uuid.UUID, userID string, orgID string, membership *mgmtPB.UserMembership) (*mgmtPB.UserMembership, error)
	DeleteUserMembership(ctx context.Context, ctxUserUID uuid.UUID, userID string, orgID string) error

	ListOrganizationMemberships(ctx context.Context, ctxUserUID uuid.UUID, orgID string) ([]*mgmtPB.OrganizationMembership, error)
	GetOrganizationMembership(ctx context.Context, ctxUserUID uuid.UUID, orgID string, userID string) (*mgmtPB.OrganizationMembership, error)
	UpdateOrganizationMembership(ctx context.Context, ctxUserUID uuid.UUID, orgID string, userID string, membership *mgmtPB.OrganizationMembership) (*mgmtPB.OrganizationMembership, error)
	DeleteOrganizationMembership(ctx context.Context, ctxUserUID uuid.UUID, orgID string, userID string) error

	CreateToken(ctx context.Context, ctxUserUID uuid.UUID, token *mgmtPB.ApiToken) error
	ListTokens(ctx context.Context, ctxUserUID uuid.UUID, pageSize int64, pageToken string) ([]*mgmtPB.ApiToken, int64, string, error)
	GetToken(ctx context.Context, ctxUserUID uuid.UUID, id string) (*mgmtPB.ApiToken, error)
	DeleteToken(ctx context.Context, ctxUserUID uuid.UUID, id string) error
	ValidateToken(ctx context.Context, accessToken string) (string, error)
	UpdateTokenLastUseTime(ctx context.Context, accessToken string) error

	CheckUserPassword(ctx context.Context, uid uuid.UUID, password string) error
	UpdateUserPassword(ctx context.Context, uid uuid.UUID, newPassword string) error

	ListPipelineTriggerRecords(ctx context.Context, owner *mgmtPB.User, pageSize int64, pageToken string, filter filtering.Filter) ([]*mgmtPB.PipelineTriggerRecord, int64, string, error)
	ListPipelineTriggerTableRecords(ctx context.Context, owner *mgmtPB.User, pageSize int64, pageToken string, filter filtering.Filter) ([]*mgmtPB.PipelineTriggerTableRecord, int64, string, error)
	ListPipelineTriggerChartRecords(ctx context.Context, owner *mgmtPB.User, aggregationWindow int64, filter filtering.Filter) ([]*mgmtPB.PipelineTriggerChartRecord, error)

	DBUser2PBUser(ctx context.Context, dbUser *datamodel.Owner) (*mgmtPB.User, error)
	DBUsers2PBUsers(ctx context.Context, dbUsers []*datamodel.Owner) ([]*mgmtPB.User, error)
	PBAuthenticatedUser2DBUser(ctx context.Context, pbUser *mgmtPB.AuthenticatedUser) (*datamodel.Owner, error)
	DBUsers2PBAuthenticatedUsers(ctx context.Context, dbUsers []*datamodel.Owner) ([]*mgmtPB.AuthenticatedUser, error)

	DBToken2PBToken(ctx context.Context, dbToken *datamodel.Token) (*mgmtPB.ApiToken, error)
	DBTokens2PBTokens(ctx context.Context, dbTokens []*datamodel.Token) ([]*mgmtPB.ApiToken, error)
	PBToken2DBToken(ctx context.Context, pbToken *mgmtPB.ApiToken) (*datamodel.Token, error)

	GetRedisClient() *redis.Client
	GetInfluxClient() repository.InfluxDB
	GetACLClient() *acl.ACLClient
}

type service struct {
	repository                  repository.Repository
	influxDB                    repository.InfluxDB
	pipelinePublicServiceClient pipelinePB.PipelinePublicServiceClient
	redisClient                 *redis.Client
	aclClient                   *acl.ACLClient
	instillCoreHost             string
}

// NewService initiates a service instance
func NewService(r repository.Repository, rc *redis.Client, i repository.InfluxDB, p pipelinePB.PipelinePublicServiceClient, acl *acl.ACLClient, h string) Service {
	return &service{
		repository:                  r,
		influxDB:                    i,
		pipelinePublicServiceClient: p,
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
			return "", ErrUnauthenticated
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
			return uuid.Nil, ErrUnauthenticated
		}
		return uuid.FromStringOrNil(headerCtxUserUID), nil
	} else {
		if !allowVisitor {
			return uuid.Nil, ErrUnauthenticated
		}
		headerCtxVisitorUID := resource.GetRequestSingleHeader(ctx, constant.HeaderVisitorUIDKey)
		if headerCtxVisitorUID == "" {
			return uuid.Nil, ErrUnauthenticated
		}

		return uuid.FromStringOrNil(headerCtxVisitorUID), nil
	}

}

func (s *service) ListUsers(ctx context.Context, ctxUserUID uuid.UUID, pageSize int, pageToken string, filter filtering.Filter) (users []*mgmtPB.User, totalSize int64, nextPageToken string, err error) {
	ctx = context.WithValue(ctx, repository.UserUIDCtxKey, ctxUserUID)
	dbUsers, totalSize, nextPageToken, err := s.repository.ListUsers(ctx, pageSize, pageToken, filter)
	if err != nil {
		return nil, 0, "", fmt.Errorf("users/ with page_size=%d page_token=%s: %w", pageSize, pageToken, err)
	}
	users, err = s.DBUsers2PBUsers(ctx, dbUsers)
	return users, totalSize, nextPageToken, err
}

func (s *service) CreateAuthenticatedUser(ctx context.Context, ctxUserUID uuid.UUID, user *mgmtPB.AuthenticatedUser) (*mgmtPB.AuthenticatedUser, error) {
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

func (s *service) GetUser(ctx context.Context, ctxUserUID uuid.UUID, id string) (*mgmtPB.User, error) {
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

func (s *service) GetUserAdmin(ctx context.Context, id string) (*mgmtPB.User, error) {

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

func (s *service) GetUserByUIDAdmin(ctx context.Context, uid uuid.UUID) (*mgmtPB.User, error) {

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

func (s *service) ListUsersAdmin(ctx context.Context, pageSize int, pageToken string, filter filtering.Filter) ([]*mgmtPB.User, int64, string, error) {
	dbUsers, totalSize, nextPageToken, err := s.repository.ListUsers(ctx, pageSize, pageToken, filter)
	if err != nil {
		return nil, 0, "", fmt.Errorf("users/ with page_size=%d page_token=%s: %w", pageSize, pageToken, err)
	}
	pbUsers, err := s.DBUsers2PBUsers(ctx, dbUsers)
	return pbUsers, totalSize, nextPageToken, err
}

func (s *service) ListAuthenticatedUsersAdmin(ctx context.Context, pageSize int, pageToken string, filter filtering.Filter) ([]*mgmtPB.AuthenticatedUser, int64, string, error) {
	dbUsers, totalSize, nextPageToken, err := s.repository.ListUsers(ctx, pageSize, pageToken, filter)
	if err != nil {
		return nil, 0, "", fmt.Errorf("users/ with page_size=%d page_token=%s: %w", pageSize, pageToken, err)
	}
	pbUsers, err := s.DBUsers2PBAuthenticatedUsers(ctx, dbUsers)
	return pbUsers, totalSize, nextPageToken, err
}

func (s *service) GetAuthenticatedUser(ctx context.Context, ctxUserUID uuid.UUID) (*mgmtPB.AuthenticatedUser, error) {
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
func (s *service) UpdateAuthenticatedUser(ctx context.Context, ctxUserUID uuid.UUID, user *mgmtPB.AuthenticatedUser) (*mgmtPB.AuthenticatedUser, error) {
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

func (s *service) ListOrganizations(ctx context.Context, ctxUserUID uuid.UUID, pageSize int, pageToken string, filter filtering.Filter) ([]*mgmtPB.Organization, int64, string, error) {
	ctx = context.WithValue(ctx, repository.UserUIDCtxKey, ctxUserUID)
	dbOrgs, totalSize, nextPageToken, err := s.repository.ListOrganizations(ctx, pageSize, pageToken, filter)
	if err != nil {
		return nil, 0, "", fmt.Errorf("organizations/ with page_size=%d page_token=%s: %w", pageSize, pageToken, err)
	}
	pbOrgs, err := s.DBOrgs2PBOrgs(ctx, dbOrgs)
	return pbOrgs, totalSize, nextPageToken, err
}

func (s *service) CreateOrganization(ctx context.Context, ctxUserUID uuid.UUID, org *mgmtPB.Organization) (*mgmtPB.Organization, error) {
	ctx = context.WithValue(ctx, repository.UserUIDCtxKey, ctxUserUID)

	uid, _ := uuid.NewV4()
	uidStr := uid.String()
	org.Uid = uidStr
	dbOrg, err := s.PBOrg2DBOrg(ctx, org)
	if err != nil {
		return nil, err
	}

	if err := s.repository.CreateOrganization(ctx, dbOrg); err != nil {
		return nil, err
	}

	err = s.aclClient.SetOrganizationUserMembership(ctx, dbOrg.UID, ctxUserUID, "owner")
	if err != nil {
		return nil, err
	}
	return s.GetOrganization(ctx, ctxUserUID, org.Id)
}

func (s *service) GetOrganization(ctx context.Context, ctxUserUID uuid.UUID, id string) (*mgmtPB.Organization, error) {

	ctx = context.WithValue(ctx, repository.UserUIDCtxKey, ctxUserUID)

	if pbOrg := s.getOrganizationFromCacheByID(ctx, id); pbOrg != nil {
		return pbOrg, nil
	}

	dbOrg, err := s.repository.GetOrganization(ctx, id, false)
	if err != nil {
		return nil, fmt.Errorf("organizations/%s: %w", id, err)
	}
	pbOrg, err := s.DBOrg2PBOrg(ctx, dbOrg)
	if err != nil {
		return nil, err
	}
	err = s.setOrganizationToCache(ctx, pbOrg)
	if err != nil {
		return nil, err
	}
	return pbOrg, nil
}

func (s *service) UpdateOrganization(ctx context.Context, ctxUserUID uuid.UUID, id string, org *mgmtPB.Organization) (*mgmtPB.Organization, error) {

	ctx = context.WithValue(ctx, repository.UserUIDCtxKey, ctxUserUID)

	oriOrg, err := s.repository.GetOrganization(ctx, id, false)
	if err != nil {
		return nil, fmt.Errorf("organizations/%s: %w", id, err)
	}
	canUpdateOrganization, err := s.aclClient.CheckOrganizationUserMembership(ctx, oriOrg.UID, ctxUserUID, "can_update_organization")
	if err != nil {
		return nil, err
	}
	if !canUpdateOrganization {
		return nil, ErrNoPermission
	}

	// Update the user
	dbOrg, err := s.PBOrg2DBOrg(ctx, org)
	if err != nil {
		return nil, err
	}
	err = s.deleteOrganizationFromCacheByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.repository.UpdateOrganization(ctx, id, dbOrg); err != nil {
		return nil, fmt.Errorf("organizations/%s: %w", id, err)
	}

	return s.GetOrganization(ctx, ctxUserUID, org.Id)

}

func (s *service) DeleteOrganization(ctx context.Context, ctxUserUID uuid.UUID, id string) error {

	ctx = context.WithValue(ctx, repository.UserUIDCtxKey, ctxUserUID)

	org, err := s.repository.GetOrganization(ctx, id, false)
	if err != nil {
		return fmt.Errorf("organizations/%s: %w", id, err)
	}

	canDeleteOrganization, err := s.aclClient.CheckOrganizationUserMembership(ctx, org.UID, ctxUserUID, "can_delete_organization")
	if err != nil {
		return err
	}
	if !canDeleteOrganization {
		return ErrNoPermission
	}

	err = s.deleteOrganizationFromCacheByID(ctx, id)
	if err != nil {
		return err
	}

	// TODO: optimize this
	pageToken := ""
	pipelineIDList := []string{}
	for {
		// TODO: optimize this cascade delete
		resp, err := s.pipelinePublicServiceClient.ListOrganizationPipelines(InjectOwnerToContext(ctx, ctxUserUID.String()),
			&pipelinePB.ListOrganizationPipelinesRequest{
				Parent:    fmt.Sprintf("organizations/%s", id),
				PageToken: &pageToken})
		if err != nil {
			break
		}
		for _, connector := range resp.Pipelines {
			pipelineIDList = append(pipelineIDList, connector.Id)
		}
		pageToken = resp.NextPageToken
		if pageToken == "" {
			break
		}
	}

	for _, pipelineID := range pipelineIDList {
		_, _ = s.pipelinePublicServiceClient.DeleteOrganizationPipeline(InjectOwnerToContext(ctx, ctxUserUID.String()),
			&pipelinePB.DeleteOrganizationPipelineRequest{
				Name: fmt.Sprintf("organizations/%s/pipelines/%s", id, pipelineID),
			})
	}

	memberships, _ := s.aclClient.GetOrganizationUsers(ctx, org.UID)

	for _, membership := range memberships {
		_ = s.aclClient.DeleteOrganizationUserMembership(ctx, org.UID, membership.UID)
	}

	err = s.repository.DeleteOrganization(ctx, id)
	if err != nil {
		return fmt.Errorf("organizations/%s: %w", id, err)
	}
	return nil
}

func (s *service) GetOrganizationAdmin(ctx context.Context, id string) (*mgmtPB.Organization, error) {

	if pbOrg := s.getOrganizationFromCacheByID(ctx, id); pbOrg != nil {
		return pbOrg, nil
	}

	dbOrganization, err := s.repository.GetOrganization(ctx, id, false)
	if err != nil {
		return nil, fmt.Errorf("organizations/%s: %w", id, err)
	}

	pbOrg, err := s.DBOrg2PBOrg(ctx, dbOrganization)
	if err != nil {
		return nil, err
	}
	err = s.setOrganizationToCache(ctx, pbOrg)
	if err != nil {
		return nil, err
	}
	return pbOrg, nil
}

func (s *service) GetOrganizationByUIDAdmin(ctx context.Context, uid uuid.UUID) (*mgmtPB.Organization, error) {

	if pbOrg := s.getOrganizationFromCacheByUID(ctx, uid); pbOrg != nil {
		return pbOrg, nil
	}

	dbOrganization, err := s.repository.GetOrganizationByUID(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("organizations/%s: %w", uid, err)
	}
	pbOrg, err := s.DBOrg2PBOrg(ctx, dbOrganization)
	if err != nil {
		return nil, err
	}
	err = s.setOrganizationToCache(ctx, pbOrg)
	if err != nil {
		return nil, err
	}
	return pbOrg, nil
}

func (s *service) ListOrganizationsAdmin(ctx context.Context, pageSize int, pageToken string, filter filtering.Filter) ([]*mgmtPB.Organization, int64, string, error) {

	dbOrganizations, totalSize, nextPageToken, err := s.repository.ListOrganizations(ctx, pageSize, pageToken, filter)
	if err != nil {
		return nil, 0, "", fmt.Errorf("organizations/ with page_size=%d page_token=%s: %w", pageSize, pageToken, err)
	}
	pbOrganizations, err := s.DBOrgs2PBOrgs(ctx, dbOrganizations)
	return pbOrganizations, totalSize, nextPageToken, err
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
		return ErrPasswordNotMatch
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

func (s *service) CreateToken(ctx context.Context, ctxUserUID uuid.UUID, token *mgmtPB.ApiToken) error {

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
	dbToken.State = datamodel.TokenState(mgmtPB.ApiToken_STATE_ACTIVE)

	switch token.GetExpiration().(type) {
	case *mgmtPB.ApiToken_Ttl:
		if token.GetTtl() >= 0 {
			dbToken.ExpireTime = curTime.Add(time.Second * time.Duration(token.GetTtl()))
		} else if token.GetTtl() == -1 {
			dbToken.ExpireTime = time.Date(2099, 12, 31, 0, 0, 0, 0, time.Now().UTC().Location())
		} else {
			return ErrInvalidTokenTTL
		}
	case *mgmtPB.ApiToken_ExpireTime:
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
func (s *service) ListTokens(ctx context.Context, ctxUserUID uuid.UUID, pageSize int64, pageToken string) ([]*mgmtPB.ApiToken, int64, string, error) {

	ctx = context.WithValue(ctx, repository.UserUIDCtxKey, ctxUserUID)

	ownerPermlink := fmt.Sprintf("users/%s", ctxUserUID.String())
	dbTokens, pageSize, pageToken, err := s.repository.ListTokens(ctx, ownerPermlink, pageSize, pageToken)
	if err != nil {
		return nil, 0, "", fmt.Errorf("tokens/ with page_size=%d page_token=%s: %w", pageSize, pageToken, err)
	}

	pbTokens, err := s.DBTokens2PBTokens(ctx, dbTokens)
	return pbTokens, pageSize, pageToken, err

}
func (s *service) GetToken(ctx context.Context, ctxUserUID uuid.UUID, id string) (*mgmtPB.ApiToken, error) {

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
			return "", ErrUnauthenticated
		}
		uid = uuid.FromStringOrNil(strings.Split(dbToken.Owner, "/")[1])
		_ = s.setAPITokenToCache(ctx, dbToken.AccessToken, uid, dbToken.ExpireTime)

	}
	if uid == uuid.Nil {
		return "", ErrUnauthenticated
	}
	return uid.String(), nil
}

func (s *service) ListUserMemberships(ctx context.Context, ctxUserUID uuid.UUID, userID string) ([]*mgmtPB.UserMembership, error) {

	ctx = context.WithValue(ctx, repository.UserUIDCtxKey, ctxUserUID)

	userID, err := s.convertUserIDAlias(ctx, ctxUserUID, userID)
	if err != nil {
		return nil, err
	}

	user, err := s.repository.GetUser(ctx, userID, false)
	if err != nil {
		return nil, fmt.Errorf("users/%s: %w", userID, err)
	}
	if ctxUserUID != user.UID {
		return nil, ErrNoPermission
	}

	orgRelations, err := s.aclClient.GetUserOrganizations(ctx, user.UID)
	if err != nil {
		return nil, err
	}

	pbUser, err := s.DBUser2PBUser(ctx, user)
	if err != nil {
		return nil, err
	}

	memberships := []*mgmtPB.UserMembership{}
	for _, orgRelation := range orgRelations {
		org, err := s.repository.GetOrganizationByUID(ctx, orgRelation.UID)
		if err == nil {
			pbOrg, err := s.DBOrg2PBOrg(ctx, org)
			if err != nil {
				return nil, err
			}
			role := orgRelation.Relation
			state := mgmtPB.MembershipState_MEMBERSHIP_STATE_ACTIVE
			if strings.HasPrefix(role, "pending") {
				role = strings.Replace(role, "pending_", "", -1)
				state = mgmtPB.MembershipState_MEMBERSHIP_STATE_PENDING
			}

			memberships = append(memberships, &mgmtPB.UserMembership{
				Name:         fmt.Sprintf("users/%s/memberships/%s", user.ID, org.ID),
				Role:         role,
				User:         pbUser,
				Organization: pbOrg,
				State:        state,
			})
		}
	}
	return memberships, nil
}

func (s *service) GetUserMembership(ctx context.Context, ctxUserUID uuid.UUID, userID string, orgID string) (*mgmtPB.UserMembership, error) {

	ctx = context.WithValue(ctx, repository.UserUIDCtxKey, ctxUserUID)

	userID, err := s.convertUserIDAlias(ctx, ctxUserUID, userID)
	if err != nil {
		return nil, err
	}

	user, err := s.repository.GetUser(ctx, userID, false)
	if err != nil {
		return nil, fmt.Errorf("users/%s: %w", userID, err)
	}
	if ctxUserUID != user.UID {
		return nil, ErrNoPermission
	}
	org, err := s.repository.GetOrganization(ctx, orgID, false)
	if err != nil {
		return nil, fmt.Errorf("organizations/%s: %w", orgID, err)
	}
	role, err := s.aclClient.GetOrganizationUserMembership(ctx, org.UID, user.UID)
	if err != nil {
		return nil, err
	}
	pbUser, err := s.DBUser2PBUser(ctx, user)
	if err != nil {
		return nil, err
	}

	pbOrg, err := s.DBOrg2PBOrg(ctx, org)
	if err != nil {
		return nil, err
	}

	state := mgmtPB.MembershipState_MEMBERSHIP_STATE_ACTIVE
	if strings.HasPrefix(role, "pending") {
		role = strings.Replace(role, "pending_", "", -1)
		state = mgmtPB.MembershipState_MEMBERSHIP_STATE_PENDING
	}
	membership := &mgmtPB.UserMembership{
		Name:         fmt.Sprintf("users/%s/memberships/%s", user.ID, org.ID),
		Role:         role,
		User:         pbUser,
		Organization: pbOrg,
		State:        state,
	}
	return membership, nil
}

func (s *service) UpdateUserMembership(ctx context.Context, ctxUserUID uuid.UUID, userID string, orgID string, membership *mgmtPB.UserMembership) (*mgmtPB.UserMembership, error) {

	ctx = context.WithValue(ctx, repository.UserUIDCtxKey, ctxUserUID)

	userID, err := s.convertUserIDAlias(ctx, ctxUserUID, userID)
	if err != nil {
		return nil, err
	}

	user, err := s.repository.GetUser(ctx, userID, false)
	if err != nil {
		return nil, fmt.Errorf("users/%s: %w", userID, err)
	}
	if ctxUserUID != user.UID {
		return nil, ErrNoPermission
	}
	org, err := s.repository.GetOrganization(ctx, orgID, false)
	if err != nil {
		return nil, fmt.Errorf("organizations/%s: %w", orgID, err)
	}
	pbUser, err := s.DBUser2PBUser(ctx, user)
	if err != nil {
		return nil, err
	}

	pbOrg, err := s.DBOrg2PBOrg(ctx, org)
	if err != nil {
		return nil, err
	}

	if membership.State == mgmtPB.MembershipState_MEMBERSHIP_STATE_ACTIVE {
		curRole, err := s.aclClient.GetOrganizationUserMembership(ctx, org.UID, user.UID)
		if err != nil {
			return nil, err
		}

		curRoleSplits := strings.Split(curRole, "_")
		if len(curRoleSplits) == 2 {
			curRole = curRoleSplits[1]
		}
		err = s.aclClient.SetOrganizationUserMembership(ctx, org.UID, user.UID, curRole)
		if err != nil {
			return nil, err
		}

		updatedMembership := &mgmtPB.UserMembership{
			Name:         fmt.Sprintf("users/%s/memberships/%s", user.ID, org.ID),
			Role:         curRole,
			User:         pbUser,
			Organization: pbOrg,
			State:        mgmtPB.MembershipState_MEMBERSHIP_STATE_ACTIVE,
		}
		return updatedMembership, nil
	}

	return nil, ErrStateCanOnlyBeActive
}

func (s *service) DeleteUserMembership(ctx context.Context, ctxUserUID uuid.UUID, userID string, orgID string) error {

	ctx = context.WithValue(ctx, repository.UserUIDCtxKey, ctxUserUID)

	userID, err := s.convertUserIDAlias(ctx, ctxUserUID, userID)
	if err != nil {
		return err
	}

	user, err := s.repository.GetUser(ctx, userID, false)
	if err != nil {
		return fmt.Errorf("users/%s: %w", userID, err)
	}
	if ctxUserUID != user.UID {
		return ErrNoPermission
	}
	org, err := s.repository.GetOrganization(ctx, orgID, false)
	if err != nil {
		return fmt.Errorf("organizations/%s: %w", orgID, err)
	}
	err = s.aclClient.DeleteOrganizationUserMembership(ctx, org.UID, user.UID)
	if err != nil {
		return err
	}
	return nil
}

func (s *service) ListOrganizationMemberships(ctx context.Context, ctxUserUID uuid.UUID, orgID string) ([]*mgmtPB.OrganizationMembership, error) {

	ctx = context.WithValue(ctx, repository.UserUIDCtxKey, ctxUserUID)

	org, err := s.repository.GetOrganization(ctx, orgID, false)
	if err != nil {
		return nil, fmt.Errorf("organizations/%s: %w", orgID, err)
	}

	canGetMembership, err := s.aclClient.CheckOrganizationUserMembership(ctx, org.UID, ctxUserUID, "can_get_membership")
	if err != nil {
		return nil, err
	}
	if !canGetMembership {
		return nil, ErrNoPermission
	}

	canSetMembership, err := s.aclClient.CheckOrganizationUserMembership(ctx, org.UID, ctxUserUID, "can_set_membership")
	if err != nil {
		return nil, err
	}

	userRelations, err := s.aclClient.GetOrganizationUsers(ctx, org.UID)
	if err != nil {
		return nil, err
	}

	pbOrg, err := s.DBOrg2PBOrg(ctx, org)
	if err != nil {
		return nil, err
	}

	memberships := []*mgmtPB.OrganizationMembership{}
	for _, userRelation := range userRelations {
		user, err := s.repository.GetUserByUID(ctx, userRelation.UID)
		if err != nil {
			return nil, fmt.Errorf("users/%s: %w", user.ID, err)
		}
		pbUser, err := s.DBUser2PBUser(ctx, user)
		if err != nil {
			return nil, err
		}
		role := userRelation.Relation
		state := mgmtPB.MembershipState_MEMBERSHIP_STATE_ACTIVE
		if strings.HasPrefix(role, "pending") {
			role = strings.Replace(role, "pending_", "", -1)
			state = mgmtPB.MembershipState_MEMBERSHIP_STATE_PENDING
		}
		if state != mgmtPB.MembershipState_MEMBERSHIP_STATE_PENDING || canSetMembership {
			memberships = append(memberships, &mgmtPB.OrganizationMembership{
				Name:         fmt.Sprintf("organizations/%s/memberships/%s", org.ID, user.ID),
				Role:         role,
				User:         pbUser,
				Organization: pbOrg,
				State:        state,
			})
		}
	}
	return memberships, nil
}

func (s *service) GetOrganizationMembership(ctx context.Context, ctxUserUID uuid.UUID, orgID string, userID string) (*mgmtPB.OrganizationMembership, error) {

	ctx = context.WithValue(ctx, repository.UserUIDCtxKey, ctxUserUID)

	userID, err := s.convertUserIDAlias(ctx, ctxUserUID, userID)
	if err != nil {
		return nil, err
	}

	user, err := s.repository.GetUser(ctx, userID, false)
	if err != nil {
		return nil, fmt.Errorf("users/%s: %w", userID, err)
	}
	org, err := s.repository.GetOrganization(ctx, orgID, false)
	if err != nil {
		return nil, fmt.Errorf("organizations/%s: %w", orgID, err)
	}

	canGetMembership, err := s.aclClient.CheckOrganizationUserMembership(ctx, org.UID, ctxUserUID, "can_get_membership")
	if err != nil {
		return nil, err
	}
	if !canGetMembership {
		return nil, ErrNoPermission
	}
	canSetMembership, err := s.aclClient.CheckOrganizationUserMembership(ctx, org.UID, ctxUserUID, "can_set_membership")
	if err != nil {
		return nil, err
	}

	role, err := s.aclClient.GetOrganizationUserMembership(ctx, org.UID, user.UID)
	if err != nil {
		return nil, err
	}
	pbUser, err := s.DBUser2PBUser(ctx, user)
	if err != nil {
		return nil, err
	}

	pbOrg, err := s.DBOrg2PBOrg(ctx, org)
	if err != nil {
		return nil, err
	}

	state := mgmtPB.MembershipState_MEMBERSHIP_STATE_ACTIVE
	if strings.HasPrefix(role, "pending") {
		role = strings.Replace(role, "pending_", "", -1)
		state = mgmtPB.MembershipState_MEMBERSHIP_STATE_PENDING
	}

	if state == mgmtPB.MembershipState_MEMBERSHIP_STATE_PENDING && ctxUserUID != uuid.FromStringOrNil(*pbUser.Uid) {
		if !canSetMembership {
			return nil, ErrNoPermission
		}
	}
	membership := &mgmtPB.OrganizationMembership{
		Name:         fmt.Sprintf("organizations/%s/memberships/%s", org.ID, user.ID),
		Role:         role,
		User:         pbUser,
		Organization: pbOrg,
		State:        state,
	}
	return membership, nil
}

func (s *service) UpdateOrganizationMembership(ctx context.Context, ctxUserUID uuid.UUID, orgID string, userID string, membership *mgmtPB.OrganizationMembership) (*mgmtPB.OrganizationMembership, error) {

	ctx = context.WithValue(ctx, repository.UserUIDCtxKey, ctxUserUID)

	userID, err := s.convertUserIDAlias(ctx, ctxUserUID, userID)
	if err != nil {
		return nil, err
	}

	user, err := s.repository.GetUser(ctx, userID, false)
	if err != nil {
		return nil, fmt.Errorf("users/%s: %w", userID, err)
	}
	org, err := s.repository.GetOrganization(ctx, orgID, false)
	if err != nil {
		return nil, fmt.Errorf("organizations/%s: %w", orgID, err)
	}

	canSetMembership, err := s.aclClient.CheckOrganizationUserMembership(ctx, org.UID, ctxUserUID, "can_set_membership")
	if err != nil {
		return nil, err
	}
	if !canSetMembership {
		return nil, ErrNoPermission
	}

	if membership.Role == "owner" {
		return nil, ErrCanNotSetAnotherOwner
	}

	pbUser, err := s.DBUser2PBUser(ctx, user)
	if err != nil {
		return nil, err
	}

	pbOrg, err := s.DBOrg2PBOrg(ctx, org)
	if err != nil {
		return nil, err
	}

	curRole, err := s.aclClient.GetOrganizationUserMembership(ctx, org.UID, user.UID)
	if err == nil && !strings.HasPrefix(curRole, "pending") {
		err = s.aclClient.SetOrganizationUserMembership(ctx, org.UID, user.UID, membership.Role)
		if err != nil {
			return nil, err
		}

		updatedMembership := &mgmtPB.OrganizationMembership{
			Name:         fmt.Sprintf("organizations/%s/memberships/%s", org.ID, user.ID),
			Role:         membership.Role,
			User:         pbUser,
			Organization: pbOrg,
			State:        mgmtPB.MembershipState_MEMBERSHIP_STATE_ACTIVE,
		}
		return updatedMembership, nil
	} else {
		err = s.aclClient.SetOrganizationUserMembership(ctx, org.UID, user.UID, "pending_"+membership.Role)
		if err != nil {
			return nil, err
		}

		updatedMembership := &mgmtPB.OrganizationMembership{
			Name:         fmt.Sprintf("organizations/%s/memberships/%s", org.ID, user.ID),
			Role:         membership.Role,
			User:         pbUser,
			Organization: pbOrg,
			State:        mgmtPB.MembershipState_MEMBERSHIP_STATE_PENDING,
		}
		return updatedMembership, nil
	}

}

func (s *service) DeleteOrganizationMembership(ctx context.Context, ctxUserUID uuid.UUID, orgID string, userID string) error {

	ctx = context.WithValue(ctx, repository.UserUIDCtxKey, ctxUserUID)

	userID, err := s.convertUserIDAlias(ctx, ctxUserUID, userID)
	if err != nil {
		return err
	}

	user, err := s.repository.GetUser(ctx, userID, false)
	if err != nil {
		return fmt.Errorf("users/%s: %w", userID, err)
	}
	org, err := s.repository.GetOrganization(ctx, orgID, false)
	if err != nil {
		return fmt.Errorf("organizations/%s: %w", orgID, err)
	}

	canRemoveMembership, err := s.aclClient.CheckOrganizationUserMembership(ctx, org.UID, ctxUserUID, "can_remove_membership")
	if err != nil {
		return err
	}
	if !canRemoveMembership {
		return ErrNoPermission
	}
	if canRemoveMembership && ctxUserUID == user.UID {
		return ErrCanNotRemoveOwnerFromOrganization
	}
	err = s.aclClient.DeleteOrganizationUserMembership(ctx, org.UID, user.UID)
	if err != nil {
		return err
	}
	return nil
}
