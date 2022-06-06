package handler

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/gofrs/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/instill-ai/mgmt-backend/pkg/datamodel"

	mgmtPB "github.com/instill-ai/protogen-go/vdp/mgmt/v1alpha"
)

// DBUser2PBUser converts a database user instance to proto user
func DBUser2PBUser(dbUser *datamodel.User) (*mgmtPB.User, error) {
	if dbUser == nil {
		return nil, errors.New("can't convert a nil user")
	}

	id := dbUser.ID

	return &mgmtPB.User{
		Name:                   fmt.Sprintf("users/%s", id),
		Uid:                    dbUser.Base.UID.String(),
		Email:                  &dbUser.Email.String,
		Id:                     id,
		CompanyName:            &dbUser.CompanyName.String,
		Role:                   &dbUser.Role.String,
		NewsletterSubscription: dbUser.NewsletterSubscription,
		CookieToken:            &dbUser.CookieToken.String,
		Type:                   mgmtPB.OwnerType_OWNER_TYPE_USER,
		CreateTime:             timestamppb.New(dbUser.Base.CreateTime),
		UpdateTime:             timestamppb.New(dbUser.Base.UpdateTime),
	}, nil
}

// PBUser2DBUser converts a proto user instance to database user
func PBUser2DBUser(pbUser *mgmtPB.User) (*datamodel.User, error) {
	if pbUser == nil {
		return nil, errors.New("can't convert a nil user")
	}

	uid, err := uuid.FromString(pbUser.Uid)
	if err != nil {
		return nil, err
	}
	email := pbUser.GetEmail()
	companyName := pbUser.GetCompanyName()
	role := pbUser.GetRole()
	cookieToken := pbUser.GetCookieToken()

	return &datamodel.User{
		Base: datamodel.Base{
			UID:        uid,
			CreateTime: pbUser.GetCreateTime().AsTime(),
			UpdateTime: pbUser.GetUpdateTime().AsTime(),
		},
		Email: sql.NullString{
			String: email,
			Valid:  len(email) > 0,
		},
		ID: pbUser.GetId(),
		CompanyName: sql.NullString{
			String: companyName,
			Valid:  len(companyName) > 0,
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
