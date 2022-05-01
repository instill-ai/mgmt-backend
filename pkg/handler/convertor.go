package handler

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/gofrs/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/instill-ai/mgmt-backend/pkg/datamodel"

	mgmtPB "github.com/instill-ai/protogen-go/mgmt/v1alpha"
)

// DBUser2PBUser converts a database user instance to proto user
func DBUser2PBUser(dbUser *datamodel.User) (*mgmtPB.User, error) {
	if dbUser == nil {
		return nil, errors.New("can't convert a nil user")
	}

	login := dbUser.Login

	return &mgmtPB.User{
		Name:                   fmt.Sprintf("users/%s", login),
		Id:                     dbUser.ID.String(),
		Email:                  &dbUser.Email.String,
		Login:                  login,
		CompanyName:            &dbUser.CompanyName.String,
		Role:                   &dbUser.Role.String,
		UsageDataCollection:    dbUser.UsageDataCollection,
		NewsletterSubscription: dbUser.NewsletterSubscription,
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

	id, err := uuid.FromString(pbUser.Id)
	if err != nil {
		return nil, err
	}
	email := pbUser.GetEmail()
	companyName := pbUser.GetCompanyName()
	role := pbUser.GetRole()

	return &datamodel.User{
		Base: datamodel.Base{
			ID:         id,
			CreateTime: pbUser.GetCreateTime().AsTime(),
			UpdateTime: pbUser.GetUpdateTime().AsTime(),
		},
		Email: sql.NullString{
			String: email,
			Valid:  len(email) > 0,
		},
		Login: pbUser.GetLogin(),
		CompanyName: sql.NullString{
			String: companyName,
			Valid:  len(companyName) > 0,
		},
		Role: sql.NullString{
			String: role,
			Valid:  len(role) > 0,
		},
		UsageDataCollection:    pbUser.GetUsageDataCollection(),
		NewsletterSubscription: pbUser.GetNewsletterSubscription(),
	}, nil
}
