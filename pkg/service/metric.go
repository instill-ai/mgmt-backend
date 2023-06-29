package service

import (
	"context"

	"go.einride.tech/aip/filtering"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/instill-ai/mgmt-backend/config"

	mgmtPB "github.com/instill-ai/protogen-go/base/mgmt/v1alpha"
)

func (s *service) ListPipielineTriggerRecord(ctx context.Context, owner *mgmtPB.User, pageSize int64, pageToken string, filter filtering.Filter) ([]*mgmtPB.PipelineTriggerRecord, int64, string, error) {

	if !config.Config.Log.External {
		return nil, 0, "", status.Errorf(codes.Internal, "[influxdb] service not found")
	}

	pipelineTriggerDataPoints, ps, pt, err := s.influxDB.QueryPipelineTriggerDataPoint(ctx, *owner.Uid, pageSize, pageToken, filter)
	if err != nil {
		return nil, 0, "", err
	}

	return pipelineTriggerDataPoints, ps, pt, nil
}
