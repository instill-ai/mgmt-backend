package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/instill-ai/mgmt-backend/pkg/datamodel"
	"github.com/instill-ai/mgmt-backend/pkg/logger"

	mgmtPB "github.com/instill-ai/protogen-go/core/mgmt/v1alpha"
)

// maps for user owner type
var (
	PBUserType2DBUserType = map[mgmtPB.OwnerType]string{
		mgmtPB.OwnerType_OWNER_TYPE_UNSPECIFIED:  "unspecified",
		mgmtPB.OwnerType_OWNER_TYPE_USER:         "user",
		mgmtPB.OwnerType_OWNER_TYPE_ORGANIZATION: "organization",
	}
	DBUserType2PBUserType = map[string]mgmtPB.OwnerType{
		"unspecified":  mgmtPB.OwnerType_OWNER_TYPE_UNSPECIFIED,
		"user":         mgmtPB.OwnerType_OWNER_TYPE_USER,
		"organization": mgmtPB.OwnerType_OWNER_TYPE_ORGANIZATION,
	}
)

// DBUser2PBUser converts a database user instance to proto user
func (s *service) DBUser2PBUser(ctx context.Context, dbUser *datamodel.Owner) (*mgmtPB.User, error) {
	if dbUser == nil {
		return nil, status.Error(codes.Internal, "can't convert a nil user")
	}

	id := dbUser.ID
	uid := dbUser.Base.UID.String()

	return &mgmtPB.User{
		Name:                   fmt.Sprintf("users/%s", id),
		Uid:                    &uid,
		Id:                     id,
		CreateTime:             timestamppb.New(dbUser.Base.CreateTime),
		UpdateTime:             timestamppb.New(dbUser.Base.UpdateTime),
		Email:                  dbUser.Email,
		CustomerId:             dbUser.CustomerId,
		FirstName:              &dbUser.FirstName.String,
		LastName:               &dbUser.LastName.String,
		OrgName:                &dbUser.OrgName.String,
		Role:                   &dbUser.Role.String,
		NewsletterSubscription: dbUser.NewsletterSubscription,
		CookieToken:            &dbUser.CookieToken.String,
	}, nil
}

// PBUser2DBUser converts a proto user instance to database user
func (s *service) PBUser2DBUser(pbUser *mgmtPB.User) (*datamodel.Owner, error) {
	if pbUser == nil {
		return nil, status.Error(codes.Internal, "can't convert a nil user")
	}

	uid, err := uuid.FromString(pbUser.GetUid())
	if err != nil {
		return nil, err
	}

	userType := "user"
	email := pbUser.GetEmail()
	customerId := pbUser.GetCustomerId()
	firstName := pbUser.GetFirstName()
	lastName := pbUser.GetLastName()
	orgName := pbUser.GetOrgName()
	role := pbUser.GetRole()
	cookieToken := pbUser.GetCookieToken()

	return &datamodel.Owner{
		Base: datamodel.Base{
			UID: uid,
		},
		ID: pbUser.GetId(),
		OwnerType: sql.NullString{
			String: userType,
			Valid:  len(userType) > 0,
		},
		Email:      email,
		CustomerId: customerId,
		FirstName: sql.NullString{
			String: firstName,
			Valid:  len(firstName) > 0,
		},
		LastName: sql.NullString{
			String: lastName,
			Valid:  len(lastName) > 0,
		},
		OrgName: sql.NullString{
			String: orgName,
			Valid:  len(orgName) > 0,
		},
		Role: sql.NullString{
			String: role,
			Valid:  len(role) > 0,
		},
		NewsletterSubscription: pbUser.GetNewsletterSubscription(),
		CookieToken: sql.NullString{
			String: cookieToken,
			Valid:  len(cookieToken) > 0,
		},
	}, nil
}

func (s *service) DBUsers2PBUsers(ctx context.Context, dbUsers []*datamodel.Owner) ([]*mgmtPB.User, error) {
	var err error
	pbUsers := make([]*mgmtPB.User, len(dbUsers))
	for idx := range dbUsers {
		pbUsers[idx], err = s.DBUser2PBUser(
			ctx,
			dbUsers[idx],
		)
		if err != nil {
			return nil, err
		}

	}
	return pbUsers, nil
}

// DBUser2PBUser converts a database user instance to proto user
func (s *service) DBOrg2PBOrg(ctx context.Context, dbOrg *datamodel.Owner) (*mgmtPB.Organization, error) {
	if dbOrg == nil {
		return nil, status.Error(codes.Internal, "can't convert a nil organization")
	}

	id := dbOrg.ID
	uid := dbOrg.Base.UID.String()

	return &mgmtPB.Organization{
		Name:       fmt.Sprintf("organizations/%s", id),
		Uid:        uid,
		Id:         id,
		CreateTime: timestamppb.New(dbOrg.Base.CreateTime),
		UpdateTime: timestamppb.New(dbOrg.Base.UpdateTime),
		CustomerId: dbOrg.CustomerId,
		OrgName:    &dbOrg.OrgName.String,
	}, nil
}

// PBUser2DBUser converts a proto user instance to database user
func (s *service) PBOrg2DBOrg(pbOrg *mgmtPB.Organization) (*datamodel.Owner, error) {
	if pbOrg == nil {
		return nil, status.Error(codes.Internal, "can't convert a nil organization")
	}

	uid, err := uuid.FromString(pbOrg.GetUid())
	if err != nil {
		return nil, err
	}

	userType := "organization"
	customerId := pbOrg.GetCustomerId()
	orgName := pbOrg.GetOrgName()

	return &datamodel.Owner{
		Base: datamodel.Base{
			UID: uid,
		},
		ID: pbOrg.GetId(),
		OwnerType: sql.NullString{
			String: userType,
			Valid:  len(userType) > 0,
		},
		CustomerId: customerId,
		OrgName: sql.NullString{
			String: orgName,
			Valid:  len(orgName) > 0,
		},
	}, nil
}

func (s *service) DBOrgs2PBOrgs(ctx context.Context, dbOrgs []*datamodel.Owner) ([]*mgmtPB.Organization, error) {
	var err error
	pbOrgs := make([]*mgmtPB.Organization, len(dbOrgs))
	for idx := range dbOrgs {
		pbOrgs[idx], err = s.DBOrg2PBOrg(
			ctx,
			dbOrgs[idx],
		)
		if err != nil {
			return nil, err
		}

	}
	return pbOrgs, nil
}

// DBToken2PBToken converts a database user instance to proto user
func (s *service) DBToken2PBToken(ctx context.Context, dbToken *datamodel.Token) (*mgmtPB.ApiToken, error) {
	id := dbToken.ID
	state := mgmtPB.ApiToken_State(dbToken.State)
	if dbToken.ExpireTime.Before(time.Now()) {
		state = mgmtPB.ApiToken_State(mgmtPB.ApiToken_STATE_EXPIRED)
	}

	return &mgmtPB.ApiToken{
		Name:        fmt.Sprintf("tokens/%s", id),
		Uid:         dbToken.Base.UID.String(),
		Id:          id,
		State:       state,
		AccessToken: dbToken.AccessToken,
		TokenType:   dbToken.TokenType,
		Expiration:  &mgmtPB.ApiToken_ExpireTime{ExpireTime: timestamppb.New(dbToken.ExpireTime)},
		CreateTime:  timestamppb.New(dbToken.Base.CreateTime),
		UpdateTime:  timestamppb.New(dbToken.Base.UpdateTime),
	}, nil
}

// PBToken2DBToken converts a proto user instance to database user
func (s *service) PBToken2DBToken(ctx context.Context, pbToken *mgmtPB.ApiToken) (*datamodel.Token, error) {

	logger, _ := logger.GetZapLogger(ctx)

	r := &datamodel.Token{
		Base: datamodel.Base{
			UID: func() uuid.UUID {
				if pbToken.GetUid() == "" {
					return uuid.UUID{}
				}
				id, err := uuid.FromString(pbToken.GetUid())
				if err != nil {
					logger.Error(err.Error())
				}
				return id
			}(),

			CreateTime: func() time.Time {
				if pbToken.GetCreateTime() != nil {
					return pbToken.GetCreateTime().AsTime()
				}
				return time.Time{}
			}(),

			UpdateTime: func() time.Time {
				if pbToken.GetUpdateTime() != nil {
					return pbToken.GetUpdateTime().AsTime()
				}
				return time.Time{}
			}(),
		},
		ID:          pbToken.GetId(),
		State:       datamodel.TokenState(pbToken.GetState()),
		AccessToken: pbToken.AccessToken,
		TokenType:   pbToken.TokenType,
		ExpireTime:  pbToken.GetExpireTime().AsTime(),
	}
	return r, nil

}

func (s *service) DBTokens2PBTokens(ctx context.Context, dbTokens []*datamodel.Token) ([]*mgmtPB.ApiToken, error) {
	var err error
	pbUsers := make([]*mgmtPB.ApiToken, len(dbTokens))
	for idx := range dbTokens {
		pbUsers[idx], err = s.DBToken2PBToken(
			ctx,
			dbTokens[idx],
		)
		if err != nil {
			return nil, err
		}

	}
	return pbUsers, nil
}
