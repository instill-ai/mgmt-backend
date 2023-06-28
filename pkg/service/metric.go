package service

import (
	"context"

	"go.einride.tech/aip/filtering"

	mgmtPB "github.com/instill-ai/protogen-go/base/mgmt/v1alpha"
)

func (s *service) ListPipielineTriggerRecord(ctx context.Context, owner *mgmtPB.User, pageSize int64, pageToken string, filter filtering.Filter) ([]*mgmtPB.PipelineTriggerRecord, int64, string, error) {

	pipelineTriggerDataPoints, ps, pt, err := s.influxDB.QueryPipelineTriggerDataPoint(ctx, *owner.Uid, pageSize, pageToken, filter)
	if err != nil {
		return nil, 0, "", err
	}

	return pipelineTriggerDataPoints, ps, pt, nil
}
