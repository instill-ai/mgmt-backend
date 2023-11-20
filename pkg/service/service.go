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
	"github.com/instill-ai/mgmt-backend/pkg/acl"
	"github.com/instill-ai/mgmt-backend/pkg/constant"
	"github.com/instill-ai/mgmt-backend/pkg/datamodel"
	"github.com/instill-ai/mgmt-backend/pkg/repository"

	mgmtPB "github.com/instill-ai/protogen-go/core/mgmt/v1alpha"
	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1alpha"
)

// Service interface
type Service interface {
	GetCtxUser(ctx context.Context) (string, uuid.UUID, error)

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

	GetUserPasswordHash(ctx context.Context, uid uuid.UUID) (string, time.Time, error)
	UpdateUserPasswordHash(ctx context.Context, uid uuid.UUID, newPassword string) error

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
}

type service struct {
	repository                   repository.Repository
	influxDB                     repository.InfluxDB
	connectorPublicServiceClient connectorPB.ConnectorPublicServiceClient
	pipelinePublicServiceClient  pipelinePB.PipelinePublicServiceClient
	redisClient                  *redis.Client
	aclClient                    *acl.ACLClient
}

// NewService initiates a service instance
func NewService(r repository.Repository, rc *redis.Client, i repository.InfluxDB, c connectorPB.ConnectorPublicServiceClient, p pipelinePB.PipelinePublicServiceClient, acl *acl.ACLClient) Service {
	return &service{
		repository:                   r,
		influxDB:                     i,
		connectorPublicServiceClient: c,
		pipelinePublicServiceClient:  p,
		redisClient:                  rc,
		aclClient:                    acl,
	}
}

// GetUser returns the api user
func (s *service) GetCtxUser(ctx context.Context) (string, uuid.UUID, error) {
	// Verify if "jwt-sub" is in the header
	headerctxUserUID := resource.GetRequestSingleHeader(ctx, constant.HeaderUserUIDKey)

	if headerctxUserUID != "" {
		_, err := uuid.FromString(headerctxUserUID)
		if err != nil {
			return "", uuid.Nil, status.Errorf(codes.Unauthenticated, "Unauthorized")
		}
		user, err := s.repository.GetUserByUID(ctx, uuid.FromStringOrNil(headerctxUserUID))
		if err != nil {
			return "", uuid.Nil, status.Errorf(codes.Unauthenticated, "Unauthorized")
		}
		return user.ID, uuid.FromStringOrNil(headerctxUserUID), nil
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
func (s *service) ListUsers(ctx context.Context, ctxUserUID uuid.UUID, pageSize int, pageToken string, filter filtering.Filter) ([]*mgmtPB.User, int64, string, error) {
	dbUsers, totalSize, nextPageToken, err := s.repository.ListUsers(ctx, pageSize, pageToken, filter)
	if err != nil {
		return nil, 0, "", err
	}
	pbUsers, err := s.DBUsers2PBUsers(ctx, dbUsers)
	return pbUsers, totalSize, nextPageToken, err
}

// CreateUser creates an user instance
// Return error types
//   - codes.InvalidArgument
//   - codes.NotFound
//   - codes.Internal
func (s *service) CreateUser(ctx context.Context, ctxUserUID uuid.UUID, user *mgmtPB.User) (*mgmtPB.User, error) {

	dbUser, err := s.PBUser2DBUser(user)
	if err != nil {
		return nil, err
	}
	if dbUser.Role.Valid {
		if r := Role(dbUser.Role.String); !ValidateRole(r) {
			return nil, status.Errorf(codes.InvalidArgument, "`role` %s in the body is not valid. Please choose from: [ %v ]", r.GetName(), strings.Join(s.ListRole(), ", "))
		}
	}
	if err := s.repository.CreateUser(ctx, dbUser); err != nil {
		return nil, err
	}

	dbCreatedUser, err := s.repository.GetUser(ctx, dbUser.ID)
	if err != nil {
		return nil, err
	}

	return s.DBUser2PBUser(ctx, dbCreatedUser)
}

// GetUser gets a user by ID
// Return error types
//   - codes.InvalidArgument
//   - codes.NotFound
func (s *service) GetUser(ctx context.Context, ctxUserUID uuid.UUID, id string) (*mgmtPB.User, error) {
	// Validation: Required field
	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "the required field `id` is not specified")
	}

	dbUser, err := s.repository.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.DBUser2PBUser(ctx, dbUser)
}

func (s *service) GetUserAdmin(ctx context.Context, id string) (*mgmtPB.User, error) {

	dbUser, err := s.repository.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.DBUser2PBUser(ctx, dbUser)
}

func (s *service) GetUserByUIDAdmin(ctx context.Context, uid uuid.UUID) (*mgmtPB.User, error) {

	dbUser, err := s.repository.GetUserByUID(ctx, uid)
	if err != nil {
		return nil, err
	}
	return s.DBUser2PBUser(ctx, dbUser)
}

func (s *service) ListUsersAdmin(ctx context.Context, pageSize int, pageToken string, filter filtering.Filter) ([]*mgmtPB.User, int64, string, error) {
	dbUsers, totalSize, nextPageToken, err := s.repository.ListUsers(ctx, pageSize, pageToken, filter)
	if err != nil {
		return nil, 0, "", err
	}
	pbUsers, err := s.DBUsers2PBUsers(ctx, dbUsers)
	return pbUsers, totalSize, nextPageToken, err
}

// UpdateUser updates a user by UUID
// Return error types
//   - codes.InvalidArgument
//   - codes.NotFound
//   - codes.Internal
func (s *service) UpdateUser(ctx context.Context, ctxUserUID uuid.UUID, id string, user *mgmtPB.User) (*mgmtPB.User, error) {

	// Check if the user exists
	if _, err := s.repository.GetUser(ctx, id); err != nil {
		return nil, err
	}

	// Update the user
	dbUser, err := s.PBUser2DBUser(user)
	if err != nil {
		return nil, err
	}
	//Validation: role field
	if dbUser.Role.Valid {
		if r := Role(dbUser.Role.String); !ValidateRole(r) {
			return nil, status.Errorf(codes.InvalidArgument, "`role` %s in the body is not valid. Please choose from: [ %v ]", r.GetName(), strings.Join(s.ListRole(), ", "))
		}
	}
	if err := s.repository.UpdateUser(ctx, id, dbUser); err != nil {
		return nil, err
	}

	dbUserUpdated, err := s.repository.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.DBUser2PBUser(ctx, dbUserUpdated)

}

// DeleteUser deletes a user by ID
// Return error types
//   - codes.InvalidArgument
//   - codes.NotFound
//   - codes.Internal
func (s *service) DeleteUser(ctx context.Context, ctxUserUID uuid.UUID, id string) error {
	// Validation: Required field
	if id == "" {
		return status.Error(codes.InvalidArgument, "the required field `id` is not specified")
	}

	return s.repository.DeleteUser(ctx, id)
}

func (s *service) ListOrganizations(ctx context.Context, ctxUserUID uuid.UUID, pageSize int, pageToken string, filter filtering.Filter) ([]*mgmtPB.Organization, int64, string, error) {
	dbOrgs, totalSize, nextPageToken, err := s.repository.ListOrganizations(ctx, pageSize, pageToken, filter)
	if err != nil {
		return nil, 0, "", err
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
		return nil, err
	}

	err = s.aclClient.SetOrganizationUserMembership(dbOrg.UID, ctxUserUID, "owner")
	if err != nil {
		return nil, err
	}

	return s.DBOrg2PBOrg(ctx, dbCreatedOrg)
}

func (s *service) GetOrganization(ctx context.Context, ctxUserUID uuid.UUID, id string) (*mgmtPB.Organization, error) {
	// Validation: Required field
	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "the required field `id` is not specified")
	}

	dbOrg, err := s.repository.GetOrganization(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.DBOrg2PBOrg(ctx, dbOrg)
}

func (s *service) UpdateOrganization(ctx context.Context, ctxUserUID uuid.UUID, id string, org *mgmtPB.Organization) (*mgmtPB.Organization, error) {

	// Check if the org exists
	if _, err := s.repository.GetOrganization(ctx, id); err != nil {
		return nil, err
	}

	// Update the user
	dbOrg, err := s.PBOrg2DBOrg(org)
	if err != nil {
		return nil, err
	}

	if err := s.repository.UpdateOrganization(ctx, id, dbOrg); err != nil {
		return nil, err
	}

	dbOrgUpdated, err := s.repository.GetOrganization(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.DBOrg2PBOrg(ctx, dbOrgUpdated)

}

func (s *service) DeleteOrganization(ctx context.Context, ctxUserUID uuid.UUID, id string) error {
	// Validation: Required field
	if id == "" {
		return status.Error(codes.InvalidArgument, "the required field `id` is not specified")
	}

	return s.repository.DeleteOrganization(ctx, id)
}

func (s *service) GetUserPasswordHash(ctx context.Context, uid uuid.UUID) (string, time.Time, error) {
	return s.repository.GetUserPasswordHash(ctx, uid)
}

func (s *service) UpdateUserPasswordHash(ctx context.Context, uid uuid.UUID, newPassword string) error {

	return s.repository.UpdateUserPasswordHash(ctx, uid, newPassword, time.Now())
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
			return status.Errorf(codes.InvalidArgument, "ttl should >= -1")
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
		return nil, 0, "", err
	}

	pbTokens, err := s.DBTokens2PBTokens(ctx, dbTokens)
	return pbTokens, pageSize, pageToken, err

}
func (s *service) GetToken(ctx context.Context, ctxUserUID uuid.UUID, id string) (*mgmtPB.ApiToken, error) {

	ownerPermlink := fmt.Sprintf("users/%s", ctxUserUID.String())
	dbToken, err := s.repository.GetToken(ctx, ownerPermlink, id)
	if err != nil {
		return nil, err
	}

	return s.DBToken2PBToken(ctx, dbToken)

}
func (s *service) DeleteToken(ctx context.Context, ctxUserUID uuid.UUID, id string) error {

	ownerPermlink := fmt.Sprintf("users/%s", ctxUserUID.String())
	token, err := s.repository.GetToken(ctx, ownerPermlink, id)
	if err != nil {
		return err
	}
	accessToken := token.AccessToken

	// TODO: should be more robust
	s.redisClient.Del(context.Background(), fmt.Sprintf(constant.AccessTokenKeyFormat, accessToken))
	delErr := s.repository.DeleteToken(ctx, ownerPermlink, id)
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

func (s *service) ListUserMemberships(ctx context.Context, ctxUserUID uuid.UUID, userID string) ([]*mgmtPB.UserMembership, error) {
	user, err := s.repository.GetUser(ctx, userID)
	if err != nil {
		return nil, err
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
			return nil, err
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
		return nil, err
	}
	org, err := s.repository.GetOrganization(ctx, orgID)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	if ctxUserUID != user.UID {
		return nil, status.Errorf(codes.PermissionDenied, "Permission Denied")
	}
	org, err := s.repository.GetOrganization(ctx, orgID)
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

	return nil, fmt.Errorf("state can only be 'active'")
}

func (s *service) DeleteUserMembership(ctx context.Context, ctxUserUID uuid.UUID, userID string, orgID string) error {
	user, err := s.repository.GetUser(ctx, userID)
	if err != nil {
		return err
	}
	if ctxUserUID != user.UID {
		return status.Errorf(codes.PermissionDenied, "Permission Denied")
	}
	org, err := s.repository.GetOrganization(ctx, orgID)
	if err != nil {
		return err
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
		return nil, err
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
			return nil, err
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
		return nil, err
	}
	org, err := s.repository.GetOrganization(ctx, orgID)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	org, err := s.repository.GetOrganization(ctx, orgID)
	if err != nil {
		return nil, err
	}

	role, err := s.aclClient.GetOrganizationUserMembership(org.UID, ctxUserUID)
	if err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "Permission Denied")
	}
	if role != "owner" {
		return nil, status.Errorf(codes.PermissionDenied, "Permission Denied")
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
		return err
	}
	org, err := s.repository.GetOrganization(ctx, orgID)
	if err != nil {
		return err
	}
	role, err := s.aclClient.GetOrganizationUserMembership(org.UID, ctxUserUID)
	if err != nil {
		return status.Errorf(codes.PermissionDenied, "Permission Denied")
	}
	if role != "owner" {
		return status.Errorf(codes.PermissionDenied, "Permission Denied")
	}
	err = s.aclClient.DeleteOrganizationUserMembership(org.UID, user.UID)
	if err != nil {
		return err
	}
	return nil
}
