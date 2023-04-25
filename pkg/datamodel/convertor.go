package datamodel

import (
	"database/sql"
	"fmt"

	"github.com/gofrs/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	mgmtPB "github.com/instill-ai/protogen-go/vdp/mgmt/v1alpha"
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
