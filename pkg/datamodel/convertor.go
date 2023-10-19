package datamodel

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

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
func DBUser2PBUser(dbUser *User) (*mgmtPB.User, error) {
	if dbUser == nil {
		return nil, status.Error(codes.Internal, "can't convert a nil user")
	}

	id := dbUser.ID
	uid := dbUser.Base.UID.String()

	return &mgmtPB.User{
		Name:                   fmt.Sprintf("users/%s", id),
		Uid:                    &uid,
		Id:                     id,
		Type:                   DBUserType2PBUserType[dbUser.OwnerType.String],
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
func PBUser2DBUser(pbUser *mgmtPB.User) (*User, error) {
	if pbUser == nil {
		return nil, status.Error(codes.Internal, "can't convert a nil user")
	}

	uid, err := uuid.FromString(pbUser.GetUid())
	if err != nil {
		return nil, err
	}

	userType := PBUserType2DBUserType[pbUser.GetType()]
	email := pbUser.GetEmail()
	customerId := pbUser.GetCustomerId()
	firstName := pbUser.GetFirstName()
	lastName := pbUser.GetLastName()
	orgName := pbUser.GetOrgName()
	role := pbUser.GetRole()
	cookieToken := pbUser.GetCookieToken()

	return &User{
		Base: Base{
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

// DBToken2PBToken converts a database user instance to proto user
func DBToken2PBToken(dbToken *Token) *mgmtPB.ApiToken {
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
	}
}

// PBToken2DBToken converts a proto user instance to database user
func PBToken2DBToken(ctx context.Context, pbToken *mgmtPB.ApiToken) *Token {

	logger, _ := logger.GetZapLogger(ctx)

	r := &Token{
		Base: Base{
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
		State:       TokenState(pbToken.GetState()),
		AccessToken: pbToken.AccessToken,
		TokenType:   pbToken.TokenType,
		ExpireTime:  pbToken.GetExpireTime().AsTime(),
	}
	return r

}
