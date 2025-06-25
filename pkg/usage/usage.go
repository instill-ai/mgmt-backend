package usage

import (
	"context"
	"fmt"
	"time"

	"github.com/instill-ai/mgmt-backend/config"
	"github.com/instill-ai/mgmt-backend/pkg/constant"
	"github.com/instill-ai/mgmt-backend/pkg/logger"
	"github.com/instill-ai/mgmt-backend/pkg/service"
	"go.einride.tech/aip/filtering"

	mgmtpb "github.com/instill-ai/protogen-go/core/mgmt/v1beta"
	usagepb "github.com/instill-ai/protogen-go/core/usage/v1beta"
	usageclient "github.com/instill-ai/usage-client/client"
	usagereporter "github.com/instill-ai/usage-client/reporter"
)

// Usage interface
type Usage interface {
	RetrieveUsageData() any
	StartReporter(ctx context.Context)
	TriggerSingleReporter(ctx context.Context)
}

type usage struct {
	service        service.Service
	reporter       usagereporter.Reporter
	serviceVersion string
}

// NewUsage initiates a usage instance
func NewUsage(ctx context.Context, s service.Service, usc usagepb.UsageServiceClient, serviceVersion string) Usage {
	logger, _ := logger.GetZapLogger(ctx)

	var defaultOwnerUID string
	if user, err := s.GetUserAdmin(ctx, constant.DefaultUserID); err == nil {
		defaultOwnerUID = *user.Uid
	} else {
		logger.Error(err.Error())
	}

	reporter, err := usageclient.InitReporter(ctx, usc, usagepb.Session_SERVICE_MGMT, config.Config.Server.Edition, serviceVersion, defaultOwnerUID)
	if err != nil {
		logger.Error(err.Error())
		return nil
	}

	return &usage{
		service:        s,
		reporter:       reporter,
		serviceVersion: serviceVersion,
	}
}

// RetrieveUsageData retrieves the server's usage data
func (u *usage) RetrieveUsageData() any {
	ctx := context.Background()
	logger, _ := logger.GetZapLogger(ctx)
	logger.Debug("[mgmt-backend] retrieve usage data...")

	allUsers := []*mgmtpb.AuthenticatedUser{}
	pageToken := ""
	for {
		users, _, token, err := u.service.ListAuthenticatedUsersAdmin(ctx, 100, pageToken, filtering.Filter{})
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

	allOrgs := []*mgmtpb.Organization{}
	pageToken = ""
	for {
		orgs, _, token, err := u.service.ListOrganizationsAdmin(ctx, 100, pageToken, filtering.Filter{})
		if err != nil {
			logger.Error(fmt.Sprintf("%s", err))
			break
		}

		pageToken = token
		allOrgs = append(allOrgs, orgs...)
		if token == "" {
			break
		}
	}

	logger.Debug(fmt.Sprintf("[mgmt-backend] user usage data length: %v", len(allUsers)))
	logger.Debug(fmt.Sprintf("[mgmt-backend] org usage data length: %v", len(allOrgs)))

	logger.Debug("[mgmt-backend] send usage data...")

	return &usagepb.SessionReport_MgmtUsageData{
		MgmtUsageData: &usagepb.MgmtUsageData{
			UserUsages: allUsers,
			OrgUsages:  allOrgs,
		},
	}
}

func (u *usage) StartReporter(ctx context.Context) {
	if u.reporter == nil {
		return
	}

	logger, _ := logger.GetZapLogger(ctx)

	var defaultOwnerUID string
	if user, err := u.service.GetUserAdmin(ctx, constant.DefaultUserID); err == nil {
		defaultOwnerUID = *user.Uid
	} else {
		logger.Error(err.Error())
	}

	go func() {
		time.Sleep(5 * time.Second)
		err := usageclient.StartReporter(ctx, u.reporter, usagepb.Session_SERVICE_MGMT, config.Config.Server.Edition, u.serviceVersion, defaultOwnerUID, u.RetrieveUsageData)
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
	if user, err := u.service.GetUserAdmin(ctx, constant.DefaultUserID); err == nil {
		defaultOwnerUID = *user.Uid
	} else {
		logger.Error(err.Error())
	}

	err := usageclient.SingleReporter(ctx, u.reporter, usagepb.Session_SERVICE_MGMT, config.Config.Server.Edition, u.serviceVersion, defaultOwnerUID, u.RetrieveUsageData())
	if err != nil {
		logger.Error(fmt.Sprintf("unable to trigger single reporter: %v\n", err))
	} else {
		logger.Debug("[mgmt-backend] trigger single reporter...")
	}
}
