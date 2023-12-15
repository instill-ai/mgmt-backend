package service

import (
	"context"
	"fmt"
	"strings"

	"go.einride.tech/aip/filtering"
	"google.golang.org/grpc/metadata"

	"github.com/gofrs/uuid"
	"github.com/instill-ai/mgmt-backend/internal/resource"
	"github.com/instill-ai/mgmt-backend/pkg/constant"
	"github.com/instill-ai/mgmt-backend/pkg/repository"

	mgmtPB "github.com/instill-ai/protogen-go/core/mgmt/v1beta"
	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1beta"
)

func InjectOwnerToContext(ctx context.Context, owner *mgmtPB.User) context.Context {
	ctx = metadata.AppendToOutgoingContext(ctx, "Jwt-Sub", owner.GetUid())
	return ctx
}

func (s *service) checkOwnership(ctx context.Context, filter filtering.Filter, owner *mgmtPB.User) (filtering.Filter, error) {

	if len(filter.CheckedExpr.GetExpr().GetCallExpr().GetArgs()) > 0 {
		ownerID, _ := repository.ExtractConstExpr(filter.CheckedExpr.GetExpr(), constant.OwnerID, false)

		if strings.HasPrefix(ownerID, "users") {
			if ownerID != fmt.Sprintf("users/%s", owner.Id) {
				return filter, ErrNoPermission
			}
			repository.HijackConstExpr(filter.CheckedExpr.GetExpr(), constant.OwnerID, constant.OwnerUID, *owner.Uid, false)
		} else if strings.HasPrefix(ownerID, "organizations") {
			id, err := resource.GetRscNameID(ownerID)
			if err != nil {
				return filter, err
			}
			org, err := s.GetOrganizationAdmin(ctx, id)
			if err != nil {
				return filter, err
			}
			granted, err := s.GetACLClient().CheckPermission("organization", uuid.FromStringOrNil(org.Uid), "user", uuid.FromStringOrNil(owner.GetUid()), "", "member")
			if err != nil {
				return filter, err
			}
			if !granted {
				return filter, ErrNoPermission
			}
			repository.HijackConstExpr(filter.CheckedExpr.GetExpr(), constant.OwnerID, constant.OwnerUID, org.Uid, false)
		} else {
			return filter, fmt.Errorf("owner_id namepsace format error")
		}
	}
	return filter, nil
}

func (s *service) pipelineUIDLookup(ctx context.Context, filter filtering.Filter, owner *mgmtPB.User) (filtering.Filter, error) {

	ctx = InjectOwnerToContext(ctx, owner)

	// lookup pipeline uid
	if len(filter.CheckedExpr.GetExpr().GetCallExpr().GetArgs()) > 0 {
		pipelineID, _ := repository.ExtractConstExpr(filter.CheckedExpr.GetExpr(), constant.PipelineID, false)

		if pipelineID != "" {
			if respPipeline, err := s.pipelinePublicServiceClient.GetUserPipeline(ctx, &pipelinePB.GetUserPipelineRequest{
				Name: fmt.Sprintf("%s/pipelines/%s", owner.Name, pipelineID),
			}); err != nil {
				return filter, err
			} else {
				repository.HijackConstExpr(filter.CheckedExpr.GetExpr(), constant.PipelineID, constant.PipelineUID, respPipeline.Pipeline.Uid, false)

				// lookup pipeline release uid
				pipelineReleaseID, _ := repository.ExtractConstExpr(filter.CheckedExpr.GetExpr(), constant.PipelineReleaseID, false)

				respPipelineRelease, err := s.pipelinePublicServiceClient.GetUserPipelineRelease(ctx, &pipelinePB.GetUserPipelineReleaseRequest{
					Name: fmt.Sprintf("%s/pipelines/%s/releases/%s", owner.Name, pipelineID, pipelineReleaseID),
				})
				if err == nil {
					repository.HijackConstExpr(filter.CheckedExpr.GetExpr(), constant.PipelineID, constant.PipelineUID, respPipelineRelease.Release.Uid, false)
				}
			}
		}
	}

	return filter, nil
}

func (s *service) connectorUIDLookup(ctx context.Context, filter filtering.Filter, owner *mgmtPB.User) (filtering.Filter, error) {

	ctx = InjectOwnerToContext(ctx, owner)

	// lookup connector uid
	if len(filter.CheckedExpr.GetExpr().GetCallExpr().GetArgs()) > 0 {
		connectorID, _ := repository.ExtractConstExpr(filter.CheckedExpr.GetExpr(), constant.ConnectorID, false)

		if connectorID != "" {
			if respConnector, err := s.pipelinePublicServiceClient.GetUserConnector(ctx, &pipelinePB.GetUserConnectorRequest{
				Name: fmt.Sprintf("%s/connectors/%s", owner.Name, connectorID),
			}); err != nil {
				return filter, err
			} else {
				repository.HijackConstExpr(filter.CheckedExpr.GetExpr(), constant.ConnectorID, constant.ConnectorUID, respConnector.Connector.Uid, false)
			}
		}
	}

	return filter, nil
}

func (s *service) ListPipelineTriggerRecords(ctx context.Context, owner *mgmtPB.User, pageSize int64, pageToken string, filter filtering.Filter) ([]*mgmtPB.PipelineTriggerRecord, int64, string, error) {

	var err error
	filter, err = s.checkOwnership(ctx, filter, owner)
	if err != nil {
		return []*mgmtPB.PipelineTriggerRecord{}, 0, "", err
	}

	filter, err = s.pipelineUIDLookup(ctx, filter, owner)
	if err != nil {
		return []*mgmtPB.PipelineTriggerRecord{}, 0, "", nil
	}

	pipelineTriggerRecords, ps, pt, err := s.influxDB.QueryPipelineTriggerRecords(ctx, *owner.Uid, pageSize, pageToken, filter)
	if err != nil {
		return nil, 0, "", err
	}

	return pipelineTriggerRecords, ps, pt, nil
}

func (s *service) ListPipelineTriggerTableRecords(ctx context.Context, owner *mgmtPB.User, pageSize int64, pageToken string, filter filtering.Filter) ([]*mgmtPB.PipelineTriggerTableRecord, int64, string, error) {

	var err error
	filter, err = s.checkOwnership(ctx, filter, owner)
	if err != nil {
		return []*mgmtPB.PipelineTriggerTableRecord{}, 0, "", err
	}

	filter, err = s.pipelineUIDLookup(ctx, filter, owner)
	if err != nil {
		return []*mgmtPB.PipelineTriggerTableRecord{}, 0, "", nil
	}

	pipelineTriggerTableRecords, ps, pt, err := s.influxDB.QueryPipelineTriggerTableRecords(ctx, *owner.Uid, pageSize, pageToken, filter)
	if err != nil {
		return nil, 0, "", err
	}

	return pipelineTriggerTableRecords, ps, pt, nil
}

func (s *service) ListPipelineTriggerChartRecords(ctx context.Context, owner *mgmtPB.User, aggregationWindow int64, filter filtering.Filter) ([]*mgmtPB.PipelineTriggerChartRecord, error) {

	var err error
	filter, err = s.checkOwnership(ctx, filter, owner)
	if err != nil {
		return []*mgmtPB.PipelineTriggerChartRecord{}, err
	}

	filter, err = s.pipelineUIDLookup(ctx, filter, owner)
	if err != nil {
		return []*mgmtPB.PipelineTriggerChartRecord{}, nil
	}

	pipelineTriggerChartRecords, err := s.influxDB.QueryPipelineTriggerChartRecords(ctx, *owner.Uid, aggregationWindow, filter)
	if err != nil {
		return nil, err
	}

	return pipelineTriggerChartRecords, nil
}

func (s *service) ListConnectorExecuteRecords(ctx context.Context, owner *mgmtPB.User, pageSize int64, pageToken string, filter filtering.Filter) ([]*mgmtPB.ConnectorExecuteRecord, int64, string, error) {

	var err error
	filter, err = s.checkOwnership(ctx, filter, owner)
	if err != nil {
		return []*mgmtPB.ConnectorExecuteRecord{}, 0, "", err
	}

	filter, err = s.pipelineUIDLookup(ctx, filter, owner)
	if err != nil {
		return []*mgmtPB.ConnectorExecuteRecord{}, 0, "", nil
	}

	filter, err = s.connectorUIDLookup(ctx, filter, owner)
	if err != nil {
		return []*mgmtPB.ConnectorExecuteRecord{}, 0, "", nil
	}

	connectorExecuteRecords, ps, pt, err := s.influxDB.QueryConnectorExecuteRecords(ctx, *owner.Uid, pageSize, pageToken, filter)
	if err != nil {
		return nil, 0, "", err
	}

	return connectorExecuteRecords, ps, pt, nil
}

func (s *service) ListConnectorExecuteTableRecords(ctx context.Context, owner *mgmtPB.User, pageSize int64, pageToken string, filter filtering.Filter) ([]*mgmtPB.ConnectorExecuteTableRecord, int64, string, error) {

	var err error
	filter, err = s.checkOwnership(ctx, filter, owner)
	if err != nil {
		return []*mgmtPB.ConnectorExecuteTableRecord{}, 0, "", err
	}

	filter, err = s.connectorUIDLookup(ctx, filter, owner)
	if err != nil {
		return []*mgmtPB.ConnectorExecuteTableRecord{}, 0, "", nil
	}

	connectorExecuteTableRecords, ps, pt, err := s.influxDB.QueryConnectorExecuteTableRecords(ctx, *owner.Uid, pageSize, pageToken, filter)
	if err != nil {
		return nil, 0, "", err
	}

	return connectorExecuteTableRecords, ps, pt, nil
}

func (s *service) ListConnectorExecuteChartRecords(ctx context.Context, owner *mgmtPB.User, aggregationWindow int64, filter filtering.Filter) ([]*mgmtPB.ConnectorExecuteChartRecord, error) {

	var err error
	filter, err = s.checkOwnership(ctx, filter, owner)
	if err != nil {
		return []*mgmtPB.ConnectorExecuteChartRecord{}, err
	}

	filter, err = s.pipelineUIDLookup(ctx, filter, owner)
	if err != nil {
		return []*mgmtPB.ConnectorExecuteChartRecord{}, nil
	}

	filter, err = s.connectorUIDLookup(ctx, filter, owner)
	if err != nil {
		return []*mgmtPB.ConnectorExecuteChartRecord{}, nil
	}

	connectorExecuteChartRecords, err := s.influxDB.QueryConnectorExecuteChartRecords(ctx, *owner.Uid, aggregationWindow, filter)
	if err != nil {
		return nil, err
	}

	return connectorExecuteChartRecords, nil
}
