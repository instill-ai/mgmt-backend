package main

import (
	"gorm.io/gorm"

	"github.com/instill-ai/mgmt-backend/pkg/handler"
	"github.com/instill-ai/mgmt-backend/pkg/repository"

	mgmtv1alpha "github.com/instill-ai/protogen-go/vdp/mgmt/v1alpha"
	usagev1alpha "github.com/instill-ai/protogen-go/vdp/usage/v1alpha"
)

// retrieveUsageData retrieves usage data of the server
func retrieveUsageData(db *gorm.DB) (interface{}, error) {
	r := repository.NewRepository(db)
	dbUsers, err := r.GetAllUsers()
	if err != nil {
		return &usagev1alpha.SessionReport_MgmtUsageData{}, err
	}

	pbUsers := []*mgmtv1alpha.User{}
	for _, v := range dbUsers {
		pbUser, err := handler.DBUser2PBUser(&v)
		if err != nil {
			return &usagev1alpha.SessionReport_MgmtUsageData{}, err
		}
		pbUsers = append(pbUsers, pbUser)
	}

	sessionReport := usagev1alpha.SessionReport_MgmtUsageData{
		MgmtUsageData: &usagev1alpha.MgmtUsageData{
			Usages: pbUsers,
		},
	}
	return &sessionReport, nil
}
