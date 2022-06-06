package usage

import (
	"fmt"

	"github.com/instill-ai/mgmt-backend/internal/logger"
	"github.com/instill-ai/mgmt-backend/pkg/handler"
	"github.com/instill-ai/mgmt-backend/pkg/repository"

	mgmtPB "github.com/instill-ai/protogen-go/vdp/mgmt/v1alpha"
	usagePB "github.com/instill-ai/protogen-go/vdp/usage/v1alpha"
)

// Usage interface
type Usage interface {
	RetrieveUsageData() interface{}
}

type usage struct {
	repository repository.Repository
}

// NewUsage initiates a usage instance
func NewUsage(r repository.Repository) Usage {
	return &usage{
		repository: r,
	}
}

// RetrieveUsageData retrieves the server's usage data
func (u *usage) RetrieveUsageData() interface{} {

	logger, _ := logger.GetZapLogger()
	logger.Debug("Retrieve usage data...")

	dbUsers, err := u.repository.GetAllUsers()
	if err != nil {
		logger.Error(fmt.Sprintf("%s", err))
	}

	pbUsers := []*mgmtPB.User{}
	for _, v := range dbUsers {
		pbUser, err := handler.DBUser2PBUser(&v)
		if err != nil {
			logger.Error(fmt.Sprintf("%s", err))
		}
		pbUsers = append(pbUsers, pbUser)
	}

	logger.Debug("Send retrieved usage data...")

	return &usagePB.SessionReport_MgmtUsageData{
		MgmtUsageData: &usagePB.MgmtUsageData{
			Usages: pbUsers,
		},
	}
}
