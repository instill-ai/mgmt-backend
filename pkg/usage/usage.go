package usage

import (
	"context"
	"fmt"
	"time"

	"github.com/instill-ai/mgmt-backend/pkg/logger"
	"github.com/instill-ai/mgmt-backend/pkg/service"
	"github.com/instill-ai/x/repo"
	"go.einride.tech/aip/filtering"

	mgmtPB "github.com/instill-ai/protogen-go/core/mgmt/v1beta"
	usagePB "github.com/instill-ai/protogen-go/core/usage/v1beta"
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
	service       service.Service
	reporter      usageReporter.Reporter
	edition       string
	version       string
	defaultUserID string
}

// NewUsage initiates a usage instance
func NewUsage(ctx context.Context, s service.Service, usc usagePB.UsageServiceClient, edition string, defaultUserID string) Usage {
	logger, _ := logger.GetZapLogger(ctx)

	version, err := repo.ReadReleaseManifest("release-please/manifest.json")
	if err != nil {
		logger.Error(err.Error())
		return nil
	}

	var defaultOwnerUID string
	if user, err := s.GetUserAdmin(ctx, defaultUserID); err == nil {
		defaultOwnerUID = *user.Uid
	} else {
		logger.Error(err.Error())
	}

	reporter, err := usageClient.InitReporter(ctx, usc, usagePB.Session_SERVICE_MGMT, edition, version, defaultOwnerUID)
	if err != nil {
		logger.Error(err.Error())
		return nil
	}

	return &usage{
		service:       s,
		reporter:      reporter,
		edition:       edition,
		version:       version,
		defaultUserID: defaultUserID,
	}
}

// RetrieveUsageData retrieves the server's usage data
func (u *usage) RetrieveUsageData() interface{} {
	ctx := context.Background()
	logger, _ := logger.GetZapLogger(ctx)
	logger.Debug("[mgmt-backend] retrieve usage data...")

	allUsers := []*mgmtPB.User{}
	pageToken := ""
	for {
		users, _, token, err := u.service.ListUsersAdmin(ctx, 100, pageToken, filtering.Filter{})
		if err != nil {
			logger.Error(fmt.Sprintf("%s", err))
			break
		}

		pageToken = token
		allUsers = append(allUsers, users...)
		if token == "" {
			break
		}
	}

	logger.Debug(fmt.Sprintf("[mgmt-backend] usage data length: %v", len(allUsers)))

	logger.Debug("[mgmt-backend] send usage data...")

	return &usagePB.SessionReport_MgmtUsageData{
		MgmtUsageData: &usagePB.MgmtUsageData{
			Usages: allUsers,
		},
	}
}

func (u *usage) StartReporter(ctx context.Context) {
	if u.reporter == nil {
		return
	}

	logger, _ := logger.GetZapLogger(ctx)

	var defaultOwnerUID string
	if user, err := u.service.GetUserAdmin(ctx, u.defaultUserID); err == nil {
		defaultOwnerUID = *user.Uid
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
	if user, err := u.service.GetUserAdmin(ctx, u.defaultUserID); err == nil {
		defaultOwnerUID = *user.Uid
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
