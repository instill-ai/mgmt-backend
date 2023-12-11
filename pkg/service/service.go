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
	"google.golang.org/grpc/metadata"

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
	AuthenticateUser(ctx context.Context, allowVisitor bool) (userID string, userUID uuid.UUID, err error)

	ListRole() []string
	CreateUser(ctx context.Context, ctxUserUID uuid.UUID, user *mgmtPB.User) (*mgmtPB.User, error)
	ListUsers(ctx context.Context, ctxUserUID uuid.UUID, pageSize int, pageToken string, filter filtering.Filter) ([]*mgmtPB.User, int64, string, error)
	GetUser(ctx context.Context, ctxUserUID uuid.UUID, id string) (*mgmtPB.User, error)
	UpdateUser(ctx context.Context, ctxUserUID uuid.UUID, id string, user *mgmtPB.User) (*mgmtPB.User, error)
	DeleteUser(ctx context.Context, ctxUserUID uuid.UUID, id string) error

	ListUsersAdmin(ctx context.Context, pageSize int, pageToken string, filter filtering.Filter) ([]*mgmtPB.User, int64, string, error)
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
	ValidateToken(accessToken string) (string, error)

	CheckUserPassword(ctx context.Context, uid uuid.UUID, password string) error
	UpdateUserPassword(ctx context.Context, uid uuid.UUID, newPassword string) error

	ListUserPipelines(ctx context.Context, id string) ([]*pipelinePB.Pipeline, error)
	ListOrganizationPipelines(ctx context.Context, id string) ([]*pipelinePB.Pipeline, error)
	ListPipelineTriggerRecords(ctx context.Context, owner *mgmtPB.User, pageSize int64, pageToken string, filter filtering.Filter) ([]*mgmtPB.PipelineTriggerRecord, int64, string, error)
	ListPipelineTriggerTableRecords(ctx context.Context, owner *mgmtPB.User, pageSize int64, pageToken string, filter filtering.Filter) ([]*mgmtPB.PipelineTriggerTableRecord, int64, string, error)
	ListPipelineTriggerChartRecords(ctx context.Context, owner *mgmtPB.User, aggregationWindow int64, filter filtering.Filter) ([]*mgmtPB.PipelineTriggerChartRecord, error)
	ListConnectorExecuteRecords(ctx context.Context, owner *mgmtPB.User, pageSize int64, pageToken string, filter filtering.Filter) ([]*mgmtPB.ConnectorExecuteRecord, int64, string, error)
	ListConnectorExecuteTableRecords(ctx context.Context, owner *mgmtPB.User, pageSize int64, pageToken string, filter filtering.Filter) ([]*mgmtPB.ConnectorExecuteTableRecord, int64, string, error)
	ListConnectorExecuteChartRecords(ctx context.Context, owner *mgmtPB.User, aggregationWindow int64, filter filtering.Filter) ([]*mgmtPB.ConnectorExecuteChartRecord, error)

	DBUser2PBUser(ctx context.Context, dbUser *datamodel.Owner) (*mgmtPB.User, error)
	DBUsers2PBUsers(ctx context.Context, dbUsers []*datamodel.Owner) ([]*mgmtPB.User, error)
	PBUser2DBUser(pbUser *mgmtPB.User) (*datamodel.Owner, error)

	DBToken2PBToken(ctx context.Context, dbToken *datamodel.Token) (*mgmtPB.ApiToken, error)
	DBTokens2PBTokens(ctx context.Context, dbTokens []*datamodel.Token) ([]*mgmtPB.ApiToken, error)
	PBToken2DBToken(ctx context.Context, pbToken *mgmtPB.ApiToken) (*datamodel.Token, error)

	GetRedisClient() *redis.Client
}

type service struct {
	repository                  repository.Repository
	influxDB                    repository.InfluxDB
	pipelinePublicServiceClient pipelinePB.PipelinePublicServiceClient
	redisClient                 *redis.Client
	aclClient                   *acl.ACLClient
}

// NewService initiates a service instance
func NewService(r repository.Repository, rc *redis.Client, i repository.InfluxDB, p pipelinePB.PipelinePublicServiceClient, acl *acl.ACLClient) Service {
	return &service{
		repository:                  r,
		influxDB:                    i,
		pipelinePublicServiceClient: p,
		redisClient:                 rc,
		aclClient:                   acl,
	}
}

func (s *service) GetRedisClient() *redis.Client {
	return s.redisClient
}

// GetUser returns the api user
func (s *service) AuthenticateUser(ctx context.Context, allowVisitor bool) (userID string, userUID uuid.UUID, err error) {
	// Verify if "jwt-sub" is in the header
	headerCtxUserUID := resource.GetRequestSingleHeader(ctx, constant.HeaderUserUIDKey)

	if headerCtxUserUID != "" {
		if allowVisitor && strings.HasPrefix(headerCtxUserUID, "visitor:") {
			_, err := uuid.FromString(strings.Split(headerCtxUserUID, ":")[1])
			if err != nil {
				return "", uuid.Nil, ErrUnauthenticated
			}
			return "", uuid.FromStringOrNil(strings.Split(headerCtxUserUID, ":")[1]), nil
		} else {
			_, err := uuid.FromString(headerCtxUserUID)
			if err != nil {
				return "", uuid.Nil, ErrUnauthenticated
			}
			user, err := s.repository.GetUserByUID(ctx, uuid.FromStringOrNil(headerCtxUserUID))
			if err != nil {
				return "", uuid.Nil, ErrUnauthenticated
			}
			return user.ID, uuid.FromStringOrNil(headerCtxUserUID), nil
		}

	}

	return "", uuid.Nil, ErrUnauthenticated

}

// ListRole lists names of all roles
func (s *service) ListRole() []string {
	return ListAllowedRoleName()
}

func (s *service) ListUsers(ctx context.Context, ctxUserUID uuid.UUID, pageSize int, pageToken string, filter filtering.Filter) (users []*mgmtPB.User, totalSize int64, nextPageToken string, err error) {
	dbUsers, totalSize, nextPageToken, err := s.repository.ListUsers(ctx, pageSize, pageToken, filter)
	if err != nil {
		return nil, 0, "", fmt.Errorf("users/ with page_size=%d page_token=%s: %w", pageSize, pageToken, err)
	}
	users, err = s.DBUsers2PBUsers(ctx, dbUsers)
	return users, totalSize, nextPageToken, err
}

func (s *service) CreateUser(ctx context.Context, ctxUserUID uuid.UUID, user *mgmtPB.User) (*mgmtPB.User, error) {

	dbUser, err := s.PBUser2DBUser(user)
	if err != nil {
		return nil, err
	}
	if dbUser.Role.Valid {
		if r := Role(dbUser.Role.String); !ValidateRole(r) {
			return nil, ErrInvalidRole
		}
	}
	if err := s.repository.CreateUser(ctx, dbUser); err != nil {
		return nil, err
	}

	dbCreatedUser, err := s.repository.GetUser(ctx, dbUser.ID)
	if err != nil {
		return nil, fmt.Errorf("users/%s: %w", dbUser.ID, err)
	}

	return s.DBUser2PBUser(ctx, dbCreatedUser)
}

func (s *service) GetUser(ctx context.Context, ctxUserUID uuid.UUID, id string) (*mgmtPB.User, error) {

	dbUser, err := s.repository.GetUser(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("users/%s: %w", id, err)
	}
	return s.DBUser2PBUser(ctx, dbUser)
}

func (s *service) GetUserAdmin(ctx context.Context, id string) (*mgmtPB.User, error) {

	dbUser, err := s.repository.GetUser(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("users/%s: %w", id, err)
	}
	return s.DBUser2PBUser(ctx, dbUser)
}

func (s *service) GetUserByUIDAdmin(ctx context.Context, uid uuid.UUID) (*mgmtPB.User, error) {

	dbUser, err := s.repository.GetUserByUID(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("users/%s: %w", uid, err)
	}
	return s.DBUser2PBUser(ctx, dbUser)
}

func (s *service) ListUsersAdmin(ctx context.Context, pageSize int, pageToken string, filter filtering.Filter) ([]*mgmtPB.User, int64, string, error) {
	dbUsers, totalSize, nextPageToken, err := s.repository.ListUsers(ctx, pageSize, pageToken, filter)
	if err != nil {
		return nil, 0, "", fmt.Errorf("users/ with page_size=%d page_token=%s: %w", pageSize, pageToken, err)
	}
	pbUsers, err := s.DBUsers2PBUsers(ctx, dbUsers)
	return pbUsers, totalSize, nextPageToken, err
}

// UpdateUser updates a user by UUID
func (s *service) UpdateUser(ctx context.Context, ctxUserUID uuid.UUID, id string, user *mgmtPB.User) (*mgmtPB.User, error) {

	// Check if the user exists
	if _, err := s.repository.GetUser(ctx, id); err != nil {
		return nil, fmt.Errorf("users/%s: %w", id, err)
	}

	// Update the user
	dbUser, err := s.PBUser2DBUser(user)
	if err != nil {
		return nil, err
	}
	//Validation: role field
	if dbUser.Role.Valid {
		if r := Role(dbUser.Role.String); !ValidateRole(r) {
			return nil, ErrInvalidRole
		}
	}
	if err := s.repository.UpdateUser(ctx, id, dbUser); err != nil {
		return nil, fmt.Errorf("users/%s: %w", id, err)
	}

	dbUserUpdated, err := s.repository.GetUser(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("users/%s: %w", id, err)
	}
	return s.DBUser2PBUser(ctx, dbUserUpdated)

}

// DeleteUser deletes a user by ID
func (s *service) DeleteUser(ctx context.Context, ctxUserUID uuid.UUID, id string) error {

	err := s.repository.DeleteUser(ctx, id)
	if err != nil {
		return fmt.Errorf("users/%s: %w", id, err)
	}
	return nil
}

func (s *service) ListOrganizations(ctx context.Context, ctxUserUID uuid.UUID, pageSize int, pageToken string, filter filtering.Filter) ([]*mgmtPB.Organization, int64, string, error) {
	dbOrgs, totalSize, nextPageToken, err := s.repository.ListOrganizations(ctx, pageSize, pageToken, filter)
	if err != nil {
		return nil, 0, "", fmt.Errorf("organizations/ with page_size=%d page_token=%s: %w", pageSize, pageToken, err)
	}
	pbOrgs, err := s.DBOrgs2PBOrgs(ctx, dbOrgs)
	return pbOrgs, totalSize, nextPageToken, err
}

func (s *service) CreateOrganization(ctx context.Context, ctxUserUID uuid.UUID, org *mgmtPB.Organization) (*mgmtPB.Organization, error) {

	uid, _ := uuid.NewV4()
	uidStr := uid.String()
	org.Uid = uidStr
	dbOrg, err := s.PBOrg2DBOrg(org)
	if err != nil {
		return nil, err
	}

	if err := s.repository.CreateOrganization(ctx, dbOrg); err != nil {
		return nil, err
	}

	dbCreatedOrg, err := s.repository.GetOrganization(ctx, dbOrg.ID)
	if err != nil {
		return nil, fmt.Errorf("organizations/%s: %w", dbOrg.ID, err)
	}

	err = s.aclClient.SetOrganizationUserMembership(dbOrg.UID, ctxUserUID, "owner")
	if err != nil {
		return nil, err
	}

	return s.DBOrg2PBOrg(ctx, dbCreatedOrg)
}

func (s *service) GetOrganization(ctx context.Context, ctxUserUID uuid.UUID, id string) (*mgmtPB.Organization, error) {

	dbOrg, err := s.repository.GetOrganization(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("organizations/%s: %w", id, err)
	}
	return s.DBOrg2PBOrg(ctx, dbOrg)
}

func (s *service) UpdateOrganization(ctx context.Context, ctxUserUID uuid.UUID, id string, org *mgmtPB.Organization) (*mgmtPB.Organization, error) {

	oriOrg, err := s.repository.GetOrganization(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("organizations/%s: %w", id, err)
	}
	canUpdateOrganization, err := s.aclClient.CheckOrganizationUserMembership(oriOrg.UID, ctxUserUID, "can_update_organization")
	if err != nil {
		return nil, err
	}
	if !canUpdateOrganization {
		return nil, ErrNoPermission
	}

	// Update the user
	dbOrg, err := s.PBOrg2DBOrg(org)
	if err != nil {
		return nil, err
	}

	if err := s.repository.UpdateOrganization(ctx, id, dbOrg); err != nil {
		return nil, fmt.Errorf("organizations/%s: %w", id, err)
	}

	dbOrgUpdated, err := s.repository.GetOrganization(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.DBOrg2PBOrg(ctx, dbOrgUpdated)

}

func (s *service) DeleteOrganization(ctx context.Context, ctxUserUID uuid.UUID, id string) error {
	org, err := s.repository.GetOrganization(ctx, id)
	if err != nil {
		return fmt.Errorf("organizations/%s: %w", id, err)
	}

	canDeleteOrganization, err := s.aclClient.CheckOrganizationUserMembership(org.UID, ctxUserUID, "can_delete_organization")
	if err != nil {
		return err
	}
	if !canDeleteOrganization {
		return ErrNoPermission
	}

	err = s.repository.DeleteOrganization(ctx, id)
	if err != nil {
		return fmt.Errorf("organizations/%s: %w", id, err)
	}
	return nil
}

func (s *service) GetOrganizationAdmin(ctx context.Context, id string) (*mgmtPB.Organization, error) {

	dbOrganization, err := s.repository.GetOrganization(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("organizations/%s: %w", id, err)
	}
	return s.DBOrg2PBOrg(ctx, dbOrganization)
}

func (s *service) GetOrganizationByUIDAdmin(ctx context.Context, uid uuid.UUID) (*mgmtPB.Organization, error) {

	dbOrganization, err := s.repository.GetOrganizationByUID(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("organizations/%s: %w", uid, err)
	}
	return s.DBOrg2PBOrg(ctx, dbOrganization)
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
	passwordHash, _, err := s.repository.GetUserPasswordHash(ctx, uid)
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		return ErrPasswordNotMatch
	}
	return nil
}

func (s *service) UpdateUserPassword(ctx context.Context, uid uuid.UUID, newPassword string) error {
	passwordBytes, err := bcrypt.GenerateFromPassword([]byte(newPassword), 10)
	if err != nil {
		return err
	}
	return s.repository.UpdateUserPasswordHash(ctx, uid, string(passwordBytes), time.Now())
}

func (s *service) CreateToken(ctx context.Context, ctxUserUID uuid.UUID, token *mgmtPB.ApiToken) error {

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
	// TODO: should be more robust
	s.redisClient.Set(context.Background(), fmt.Sprintf(constant.AccessTokenKeyFormat, dbToken.AccessToken), dbToken.Owner, 0)
	s.redisClient.ExpireAt(context.Background(), fmt.Sprintf(constant.AccessTokenKeyFormat, dbToken.AccessToken), dbToken.ExpireTime)

	return nil
}
func (s *service) ListTokens(ctx context.Context, ctxUserUID uuid.UUID, pageSize int64, pageToken string) ([]*mgmtPB.ApiToken, int64, string, error) {
	ownerPermlink := fmt.Sprintf("users/%s", ctxUserUID.String())
	dbTokens, pageSize, pageToken, err := s.repository.ListTokens(ctx, ownerPermlink, pageSize, pageToken)
	if err != nil {
		return nil, 0, "", fmt.Errorf("tokens/ with page_size=%d page_token=%s: %w", pageSize, pageToken, err)
	}

	pbTokens, err := s.DBTokens2PBTokens(ctx, dbTokens)
	return pbTokens, pageSize, pageToken, err

}
func (s *service) GetToken(ctx context.Context, ctxUserUID uuid.UUID, id string) (*mgmtPB.ApiToken, error) {

	ownerPermlink := fmt.Sprintf("users/%s", ctxUserUID.String())
	dbToken, err := s.repository.GetToken(ctx, ownerPermlink, id)
	if err != nil {
		return nil, fmt.Errorf("tokens/%s: %w", id, err)
	}

	return s.DBToken2PBToken(ctx, dbToken)

}
func (s *service) DeleteToken(ctx context.Context, ctxUserUID uuid.UUID, id string) error {

	ownerPermlink := fmt.Sprintf("users/%s", ctxUserUID.String())
	token, err := s.repository.GetToken(ctx, ownerPermlink, id)
	if err != nil {
		return fmt.Errorf("tokens/%s: %w", id, err)
	}
	accessToken := token.AccessToken

	// TODO: should be more robust
	s.redisClient.Del(context.Background(), fmt.Sprintf(constant.AccessTokenKeyFormat, accessToken))
	err = s.repository.DeleteToken(ctx, ownerPermlink, id)
	if err != nil {
		return fmt.Errorf("tokens/%s: %w", id, err)
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

func (s *service) ListUserMemberships(ctx context.Context, ctxUserUID uuid.UUID, userID string) ([]*mgmtPB.UserMembership, error) {
	user, err := s.repository.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("users/%s: %w", userID, err)
	}
	if ctxUserUID != user.UID {
		return nil, ErrNoPermission
	}

	orgRelations, err := s.aclClient.GetUserOrganizations(user.UID)
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
		if err != nil {
			return nil, fmt.Errorf("organizations/%s: %w", org.ID, err)
		}
		pbOrg, err := s.DBOrg2PBOrg(ctx, org)
		if err != nil {
			return nil, err
		}

		memberships = append(memberships, &mgmtPB.UserMembership{
			Name:         fmt.Sprintf("users/%s/memberships/%s", user.ID, org.ID),
			Role:         orgRelation.Relation,
			User:         pbUser,
			Organization: pbOrg,
			State:        mgmtPB.MembershipState_MEMBERSHIP_STATE_ACTIVE,
		})
	}
	return memberships, nil
}

func (s *service) GetUserMembership(ctx context.Context, ctxUserUID uuid.UUID, userID string, orgID string) (*mgmtPB.UserMembership, error) {
	user, err := s.repository.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("users/%s: %w", userID, err)
	}
	if ctxUserUID != user.UID {
		return nil, ErrNoPermission
	}
	org, err := s.repository.GetOrganization(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("organizations/%s: %w", orgID, err)
	}
	role, err := s.aclClient.GetOrganizationUserMembership(org.UID, user.UID)
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

	membership := &mgmtPB.UserMembership{
		Name:         fmt.Sprintf("users/%s/memberships/%s", user.ID, org.ID),
		Role:         role,
		User:         pbUser,
		Organization: pbOrg,
		State:        mgmtPB.MembershipState_MEMBERSHIP_STATE_ACTIVE,
	}
	return membership, nil
}

func (s *service) UpdateUserMembership(ctx context.Context, ctxUserUID uuid.UUID, userID string, orgID string, membership *mgmtPB.UserMembership) (*mgmtPB.UserMembership, error) {
	user, err := s.repository.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("users/%s: %w", userID, err)
	}
	if ctxUserUID != user.UID {
		return nil, ErrNoPermission
	}
	org, err := s.repository.GetOrganization(ctx, orgID)
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
		curRole, err := s.aclClient.GetOrganizationUserMembership(org.UID, user.UID)
		if err != nil {
			return nil, err
		}

		curRoleSplits := strings.Split(curRole, "_")
		if len(curRoleSplits) == 2 {
			curRole = curRoleSplits[1]
		}
		err = s.aclClient.SetOrganizationUserMembership(org.UID, user.UID, curRole)
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
	user, err := s.repository.GetUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("users/%s: %w", userID, err)
	}
	if ctxUserUID != user.UID {
		return ErrNoPermission
	}
	org, err := s.repository.GetOrganization(ctx, orgID)
	if err != nil {
		return fmt.Errorf("organizations/%s: %w", orgID, err)
	}
	err = s.aclClient.DeleteOrganizationUserMembership(org.UID, user.UID)
	if err != nil {
		return err
	}
	return nil
}

func (s *service) ListOrganizationMemberships(ctx context.Context, ctxUserUID uuid.UUID, orgID string) ([]*mgmtPB.OrganizationMembership, error) {
	org, err := s.repository.GetOrganization(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("organizations/%s: %w", orgID, err)
	}

	canGetMembership, err := s.aclClient.CheckOrganizationUserMembership(org.UID, ctxUserUID, "can_get_membership")
	if err != nil {
		return nil, err
	}
	if !canGetMembership {
		return nil, ErrNoPermission
	}

	userRelations, err := s.aclClient.GetOrganizationUsers(org.UID)
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

		memberships = append(memberships, &mgmtPB.OrganizationMembership{
			Name:         fmt.Sprintf("organizations/%s/memberships/%s", user.ID, org.ID),
			Role:         userRelation.Relation,
			User:         pbUser,
			Organization: pbOrg,
			State:        mgmtPB.MembershipState_MEMBERSHIP_STATE_ACTIVE,
		})
	}
	return memberships, nil
}

func (s *service) GetOrganizationMembership(ctx context.Context, ctxUserUID uuid.UUID, orgID string, userID string) (*mgmtPB.OrganizationMembership, error) {
	user, err := s.repository.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("users/%s: %w", userID, err)
	}
	org, err := s.repository.GetOrganization(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("organizations/%s: %w", orgID, err)
	}

	canGetMembership, err := s.aclClient.CheckOrganizationUserMembership(org.UID, ctxUserUID, "can_get_membership")
	if err != nil {
		return nil, err
	}
	if !canGetMembership {
		return nil, ErrNoPermission
	}

	role, err := s.aclClient.GetOrganizationUserMembership(org.UID, user.UID)
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

	membership := &mgmtPB.OrganizationMembership{
		Name:         fmt.Sprintf("organizations/%s/memberships/%s", user.ID, org.ID),
		Role:         role,
		User:         pbUser,
		Organization: pbOrg,
		State:        mgmtPB.MembershipState_MEMBERSHIP_STATE_ACTIVE,
	}
	return membership, nil
}

func (s *service) UpdateOrganizationMembership(ctx context.Context, ctxUserUID uuid.UUID, orgID string, userID string, membership *mgmtPB.OrganizationMembership) (*mgmtPB.OrganizationMembership, error) {
	user, err := s.repository.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("users/%s: %w", userID, err)
	}
	org, err := s.repository.GetOrganization(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("organizations/%s: %w", orgID, err)
	}

	canSetMembership, err := s.aclClient.CheckOrganizationUserMembership(org.UID, ctxUserUID, "can_set_membership")
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

	curRole, err := s.aclClient.GetOrganizationUserMembership(org.UID, user.UID)
	if err == nil && !strings.HasPrefix(curRole, "pending") {
		err = s.aclClient.SetOrganizationUserMembership(org.UID, user.UID, membership.Role)
		if err != nil {
			return nil, err
		}

		updatedMembership := &mgmtPB.OrganizationMembership{
			Name:         fmt.Sprintf("organizations/%s/memberships/%s", user.ID, org.ID),
			Role:         membership.Role,
			User:         pbUser,
			Organization: pbOrg,
			State:        mgmtPB.MembershipState_MEMBERSHIP_STATE_ACTIVE,
		}
		return updatedMembership, nil
	} else {
		err = s.aclClient.SetOrganizationUserMembership(org.UID, user.UID, "pending_"+membership.Role)
		if err != nil {
			return nil, err
		}

		updatedMembership := &mgmtPB.OrganizationMembership{
			Name:         fmt.Sprintf("organizations/%s/memberships/%s", user.ID, org.ID),
			Role:         membership.Role,
			User:         pbUser,
			Organization: pbOrg,
			State:        mgmtPB.MembershipState_MEMBERSHIP_STATE_PENDING,
		}
		return updatedMembership, nil
	}

}

func (s *service) DeleteOrganizationMembership(ctx context.Context, ctxUserUID uuid.UUID, orgID string, userID string) error {

	user, err := s.repository.GetUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("users/%s: %w", userID, err)
	}
	org, err := s.repository.GetOrganization(ctx, orgID)
	if err != nil {
		return fmt.Errorf("organizations/%s: %w", orgID, err)
	}

	canRemoveMembership, err := s.aclClient.CheckOrganizationUserMembership(org.UID, ctxUserUID, "can_remove_membership")
	if err != nil {
		return err
	}
	if !canRemoveMembership {
		return ErrNoPermission
	}
	if canRemoveMembership && ctxUserUID == user.UID {
		return ErrCanNotRemoveOwnerFromOrganization
	}
	err = s.aclClient.DeleteOrganizationUserMembership(org.UID, user.UID)
	if err != nil {
		return err
	}
	return nil
}
func (s *service) ListUserPipelines(ctx context.Context, id string) ([]*pipelinePB.Pipeline, error) {

	pageToken := ""
	pageSize := int32(100)

	pipelines := []*pipelinePB.Pipeline{}
	for {
		resp, err := s.pipelinePublicServiceClient.ListUserPipelines(
			metadata.AppendToOutgoingContext(ctx, "Jwt-Sub", resource.GetRequestSingleHeader(ctx, constant.HeaderUserUIDKey)),
			&pipelinePB.ListUserPipelinesRequest{
				PageSize:  &pageSize,
				PageToken: &pageToken,
				Parent:    fmt.Sprintf("users/%s", id),
			},
		)
		if err != nil {
			return nil, err
		}
		pipelines = append(pipelines, resp.Pipelines...)
		pageToken = resp.NextPageToken
		if pageToken == "" {
			break
		}
	}

	return pipelines, nil
}

func (s *service) ListOrganizationPipelines(ctx context.Context, id string) ([]*pipelinePB.Pipeline, error) {

	pageToken := ""
	pageSize := int32(100)

	pipelines := []*pipelinePB.Pipeline{}
	for {
		resp, err := s.pipelinePublicServiceClient.ListOrganizationPipelines(
			metadata.AppendToOutgoingContext(ctx, "Jwt-Sub", resource.GetRequestSingleHeader(ctx, constant.HeaderUserUIDKey)),
			&pipelinePB.ListOrganizationPipelinesRequest{
				PageSize:  &pageSize,
				PageToken: &pageToken,
				Parent:    fmt.Sprintf("organizations/%s", id),
			},
		)
		if err != nil {
			return nil, err
		}
		pipelines = append(pipelines, resp.Pipelines...)
		pageToken = resp.NextPageToken
		if pageToken == "" {
			break
		}
	}

	return pipelines, nil
}
