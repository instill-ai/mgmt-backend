package handler

import (
	"database/sql"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/instill-ai/mgmt-backend/pkg/datamodel"
	mgmtPB "github.com/instill-ai/protogen-go/mgmt/v1alpha"
	"google.golang.org/protobuf/types/known/timestamppb"
)

//
func DBUser2PBUser(dbUser *datamodel.User) *mgmtPB.User {

	login := dbUser.Login

	return &mgmtPB.User{
		Name:                   fmt.Sprintf("users/%s", login),
		Id:                     dbUser.Id.String(),
		Email:                  &dbUser.Email.String,
		Login:                  login,
		CompanyName:            &dbUser.CompanyName.String,
		Role:                   &dbUser.Role.String,
		UsageDataCollection:    dbUser.UsageDataCollection,
		NewsletterSubscription: dbUser.NewsletterSubscription,
		CreatedAt:              timestamppb.New(dbUser.Base.CreatedAt),
		UpdatedAt:              timestamppb.New(dbUser.Base.UpdatedAt),
	}
}

//
func PBUser2DBUser(pbUser *mgmtPB.User) (*datamodel.User, error) {
	id, err := uuid.FromString(pbUser.Id)
	if err != nil {
		return nil, err
	}
	email := pbUser.GetEmail()
	companyName := pbUser.GetCompanyName()
	role := pbUser.GetRole()

	return &datamodel.User{
		Base: datamodel.Base{
			Id:        id,
			CreatedAt: pbUser.GetCreatedAt().AsTime(),
			UpdatedAt: pbUser.GetUpdatedAt().AsTime(),
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
