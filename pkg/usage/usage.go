package usage

import (
	"context"
	"fmt"
	"time"

	"github.com/instill-ai/mgmt-backend/pkg/datamodel"
	"github.com/instill-ai/mgmt-backend/pkg/logger"
	"github.com/instill-ai/mgmt-backend/pkg/repository"
	"github.com/instill-ai/x/repo"

	mgmtPB "github.com/instill-ai/protogen-go/vdp/mgmt/v1alpha"
	usagePB "github.com/instill-ai/protogen-go/vdp/usage/v1alpha"
	usageClient "github.com/instill-ai/usage-client/client"
	usageReporter "github.com/instill-ai/usage-client/reporter"
)

// Usage interface
type Usage interface {
	RetrieveUsageData() interface{}
	StartReporter(ctx context.Context)
	TriggerSingleReporter(ctx context.Context)
}

type usage struct {
	repository repository.Repository
	reporter   usageReporter.Reporter
	edition    string
	version    string
}

// NewUsage initiates a usage instance
func NewUsage(ctx context.Context, r repository.Repository, usc usagePB.UsageServiceClient, edition string) Usage {
	logger, _ := logger.GetZapLogger()

	version, err := repo.ReadReleaseManifest("release-please/manifest.json")
	if err != nil {
		logger.Error(err.Error())
		return nil
	}

	reporter, err := usageClient.InitReporter(ctx, usc, usagePB.Session_SERVICE_MGMT, edition, version)
	if err != nil {
		logger.Error(err.Error())
		return nil
	}

	return &usage{
		repository: r,
		reporter:   reporter,
		edition:    edition,
		version:    version,
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
		pbUser, err := datamodel.DBUser2PBUser(&v)
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

func (u *usage) StartReporter(ctx context.Context) {
	if u.reporter == nil {
		return
	}

	logger, _ := logger.GetZapLogger()
	go func() {
		time.Sleep(5 * time.Second)
		err := usageClient.StartReporter(ctx, u.reporter, usagePB.Session_SERVICE_MGMT, u.edition, u.version, u.RetrieveUsageData)
		if err != nil {
			logger.Error(fmt.Sprintf("unable to start reporter: %v\n", err))
		}
	}()
}

func (u *usage) TriggerSingleReporter(ctx context.Context) {
	if u.reporter == nil {
		return
	}
	logger, _ := logger.GetZapLogger()
	err := usageClient.SingleReporter(ctx, u.reporter, usagePB.Session_SERVICE_MGMT, u.edition, u.version, u.RetrieveUsageData())
	if err != nil {
		logger.Error(fmt.Sprintf("unable to trigger single reporter: %v\n", err))
	}
}
