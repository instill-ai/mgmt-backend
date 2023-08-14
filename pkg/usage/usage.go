package usage

import (
	"context"
	"fmt"
	"time"

	"github.com/instill-ai/mgmt-backend/pkg/constant"
	"github.com/instill-ai/mgmt-backend/pkg/datamodel"
	"github.com/instill-ai/mgmt-backend/pkg/logger"
	"github.com/instill-ai/mgmt-backend/pkg/repository"
	"github.com/instill-ai/x/repo"

	mgmtPB "github.com/instill-ai/protogen-go/base/mgmt/v1alpha"
	usagePB "github.com/instill-ai/protogen-go/base/usage/v1alpha"
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
	logger, _ := logger.GetZapLogger(ctx)

	version, err := repo.ReadReleaseManifest("release-please/manifest.json")
	if err != nil {
		logger.Error(err.Error())
		return nil
	}

	var defaultOwnerUID string
	if user, err := r.GetUserByID(constant.DefaultUserID); err == nil {
		defaultOwnerUID = user.UID.String()
	} else {
		logger.Error(err.Error())
	}

	reporter, err := usageClient.InitReporter(ctx, usc, usagePB.Session_SERVICE_MGMT, edition, version, defaultOwnerUID)
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
	ctx := context.Background()
	logger, _ := logger.GetZapLogger(ctx)
	logger.Debug("[mgmt-backend] retrieve usage data...")

	dbUsers, err := u.repository.GetAllUsers(ctx)
	if err != nil {
		logger.Error(fmt.Sprintf("%s", err))
	}

	pbUsers := []*mgmtPB.User{}
	for idx := range dbUsers {
		pbUser, err := datamodel.DBUser2PBUser(&dbUsers[idx])
		if err != nil {
			logger.Error(fmt.Sprintf("%s", err))
		}
		pbUsers = append(pbUsers, pbUser)
	}

	logger.Debug(fmt.Sprintf("[mgmt-backend] usage data length: %v", len(pbUsers)))

	logger.Debug("[mgmt-backend] send usage data...")

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

	logger, _ := logger.GetZapLogger(ctx)

	var defaultOwnerUID string
	if user, err := u.repository.GetUserByID(constant.DefaultUserID); err == nil {
		defaultOwnerUID = user.UID.String()
	} else {
		logger.Error(err.Error())
	}

	go func() {
		time.Sleep(5 * time.Second)
		err := usageClient.StartReporter(ctx, u.reporter, usagePB.Session_SERVICE_MGMT, u.edition, u.version, defaultOwnerUID, u.RetrieveUsageData)
		if err != nil {
			logger.Error(fmt.Sprintf("unable to start reporter: %v\n", err))
		}
	}()
}

func (u *usage) TriggerSingleReporter(ctx context.Context) {
	if u.reporter == nil {
		return
	}

	logger, _ := logger.GetZapLogger(ctx)

	var defaultOwnerUID string
	if user, err := u.repository.GetUserByID(constant.DefaultUserID); err == nil {
		defaultOwnerUID = user.UID.String()
	} else {
		logger.Error(err.Error())
	}

	err := usageClient.SingleReporter(ctx, u.reporter, usagePB.Session_SERVICE_MGMT, u.edition, u.version, defaultOwnerUID, u.RetrieveUsageData())
	if err != nil {
		logger.Error(fmt.Sprintf("unable to trigger single reporter: %v\n", err))
	} else {
		logger.Debug("[mgmt-backend] trigger single reporter...")
	}
}
