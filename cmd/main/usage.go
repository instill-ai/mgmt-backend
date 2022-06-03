package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/instill-ai/mgmt-backend/pkg/handler"
	"github.com/instill-ai/mgmt-backend/pkg/repository"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"gorm.io/gorm"

	mgmtPB "github.com/instill-ai/protogen-go/vdp/mgmt/v1alpha"
	usagePB "github.com/instill-ai/protogen-go/vdp/usage/v1alpha"
	usageclient "github.com/instill-ai/usage-client/usage"
)

func readReleaseTag() (string, error) {
	type Release struct {
		Tag string `json:".,"` // field appears in JSON as key "."
	}

	content, err := ioutil.ReadFile("release-please/manifest.json")
	if err != nil {
		return "", fmt.Errorf("can't read release manifest.json file: %v", err)
	}
	release := Release{}
	err = json.Unmarshal(content, &release)
	if err != nil {
		return "", err
	}
	return release.Tag, nil
}

// retrieveUsageData retrieves usage data of the server
func retrieveUsageData(db *gorm.DB) (usageclient.IsReportUsageData, error) {
	r := repository.NewRepository(db)
	dbUsers, err := r.GetAllUsers()
	if err != nil {
		return &usageclient.SessionReportMgmtUsageData{}, err
	}

	pbUsers := []*mgmtPB.User{}
	for _, v := range dbUsers {
		pbUser, err := handler.DBUser2PBUser(&v)
		if err != nil {
			return &usageclient.SessionReportMgmtUsageData{}, err
		}
		pbUsers = append(pbUsers, pbUser)
	}

	sessionReport := usageclient.SessionReportMgmtUsageData{
		MgmtUsageData: &usagePB.MgmtUsageData{
			Usages: pbUsers,
		},
	}
	return &sessionReport, nil
}

// startReporter starts the usage collection for a server session
func startReporter(ctx context.Context, db *gorm.DB, conn *grpc.ClientConn, logger *zap.Logger, url, env, version string) {
	usageDelay := 5 * time.Second
	service := usagePB.Session_SERVICE_MGMT
	time.Sleep(usageDelay)
	err := usageclient.StartReporter(ctx, db, conn, service, url, env, version, retrieveUsageData)
	if err != nil {
		logger.Error(fmt.Sprintf("unable to start reporter: %v\n", err))
	}
}
