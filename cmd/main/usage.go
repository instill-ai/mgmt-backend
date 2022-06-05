package main

import (
	"gorm.io/gorm"

	"github.com/instill-ai/mgmt-backend/pkg/handler"
	"github.com/instill-ai/mgmt-backend/pkg/repository"

	mgmtPB "github.com/instill-ai/protogen-go/vdp/mgmt/v1alpha"
	usagePB "github.com/instill-ai/protogen-go/vdp/usage/v1alpha"
)

// retrieveUsageData retrieves usage data of the server
func retrieveUsageData(db *gorm.DB) (interface{}, error) {
	r := repository.NewRepository(db)
	dbUsers, err := r.GetAllUsers()
	if err != nil {
		return &usagePB.SessionReport_MgmtUsageData{}, err
	}

	pbUsers := []*mgmtPB.User{}
	for _, v := range dbUsers {
		pbUser, err := handler.DBUser2PBUser(&v)
		if err != nil {
			return &usagePB.SessionReport_MgmtUsageData{}, err
		}
		pbUsers = append(pbUsers, pbUser)
	}

	sessionReport := usagePB.SessionReport_MgmtUsageData{
		MgmtUsageData: &usagePB.MgmtUsageData{
			Usages: pbUsers,
		},
	}
	return &sessionReport, nil
}
