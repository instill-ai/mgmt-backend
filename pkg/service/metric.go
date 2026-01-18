package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"go.einride.tech/aip/filtering"

	"github.com/instill-ai/mgmt-backend/pkg/constant"
	"github.com/instill-ai/mgmt-backend/pkg/repository"

	mgmtpb "github.com/instill-ai/protogen-go/mgmt/v1beta"
	pipelinepb "github.com/instill-ai/protogen-go/pipeline/v1beta"
	errorsx "github.com/instill-ai/x/errors"
	gatewayx "github.com/instill-ai/x/server/grpc/gateway"
)

var ErrNoPermission = errors.New("no permission")

// checkPipelineOwnership validates ownership and returns the owner UID for filtering.
// The ownerUID parameter is the internal UUID of the owner (user).
func (s *service) checkPipelineOwnership(ctx context.Context, filter filtering.Filter, owner *mgmtpb.User, ownerUID uuid.UUID) (*string, string, string, string, filtering.Filter, error) {
	ownerID := owner.Id
	ownerUIDStr := ownerUID.String()
	resultOwnerUID := &ownerUIDStr
	ownerType := "users"
	ownerQueryString := ""

	if len(filter.CheckedExpr.GetExpr().GetCallExpr().GetArgs()) > 0 {

		ownerName, _ := repository.ExtractConstExpr(filter.CheckedExpr.GetExpr(), constant.OwnerName, false)

		if ownerName != "" {

			if strings.HasPrefix(ownerName, "users") {
				if ownerName != fmt.Sprintf("users/%s", owner.Id) {
					return nil, "", "", "", filter, ErrNoPermission
				}
				repository.HijackConstExpr(filter.CheckedExpr.GetExpr(), constant.OwnerName, constant.PipelineOwnerUID, ownerUIDStr, false)
			} else if strings.HasPrefix(ownerName, "organizations") {
				// Organizations are EE-only in CE
				return nil, "", "", "", filter, ErrNoPermission
			} else {
				return nil, "", "", "", filter, errorsx.ErrInvalidOwnerNamespace
			}
		} else {
			ownerQueryString = fmt.Sprintf("|> filter(fn: (r) => r[\"owner_uid\"] == \"%v\")", ownerUIDStr)
		}
	} else {
		ownerQueryString = fmt.Sprintf("|> filter(fn: (r) => r[\"owner_uid\"] == \"%v\")", ownerUIDStr)
	}
	return resultOwnerUID, ownerID, ownerType, ownerQueryString, filter, nil
}

// pipelineUIDLookup looks up pipeline UIDs for filtering.
// The ownerUID parameter is the internal UUID of the owner (user).
func (s *service) pipelineUIDLookup(ctx context.Context, ownerID string, ownerType string, filter filtering.Filter, owner *mgmtpb.User, ownerUID uuid.UUID) (filtering.Filter, error) {

	ctx = gatewayx.InjectOwnerToContext(ctx, ownerUID.String())

	// lookup pipeline uid
	if len(filter.CheckedExpr.GetExpr().GetCallExpr().GetArgs()) > 0 {
		pipelineID, _ := repository.ExtractConstExpr(filter.CheckedExpr.GetExpr(), constant.PipelineID, false)
		namespaceID := strings.Split(owner.Name, "/")[1]

		if pipelineID != "" {
			respPipeline, err := s.pipelinePublicServiceClient.GetNamespacePipeline(ctx, &pipelinepb.GetNamespacePipelineRequest{
				Name: fmt.Sprintf("namespaces/%s/pipelines/%s", namespaceID, pipelineID),
			})
			if err != nil {
				return filter, err
			}

			// Use pipeline Id since Uid is no longer in the public API
			repository.HijackConstExpr(filter.CheckedExpr.GetExpr(), constant.PipelineID, constant.PipelineUID, respPipeline.Pipeline.Id, false)

			// lookup pipeline release uid
			pipelineReleaseID, _ := repository.ExtractConstExpr(filter.CheckedExpr.GetExpr(), constant.PipelineReleaseID, false)

			respPipelineRelease, err := s.pipelinePublicServiceClient.GetNamespacePipelineRelease(ctx, &pipelinepb.GetNamespacePipelineReleaseRequest{
				Name: fmt.Sprintf("namespaces/%s/pipelines/%s/releases/%s", namespaceID, pipelineID, pipelineReleaseID),
			})
			if err == nil {
				// Use release Id since Uid is no longer in the public API
				repository.HijackConstExpr(filter.CheckedExpr.GetExpr(), constant.PipelineID, constant.PipelineUID, respPipelineRelease.Release.Id, false)
			}
		}
	}

	return filter, nil
}

func (s *service) ListPipelineTriggerRecords(ctx context.Context, owner *mgmtpb.User, pageSize int64, pageToken string, filter filtering.Filter) ([]*mgmtpb.PipelineTriggerRecord, int64, string, error) {
	// Look up the owner's internal UID from their public ID
	ownerInternalUID, err := s.GetUserUIDByID(ctx, owner.Id)
	if err != nil {
		return []*mgmtpb.PipelineTriggerRecord{}, 0, "", err
	}
	ownerUID, ownerID, ownerType, ownerQueryString, filter, err := s.checkPipelineOwnership(ctx, filter, owner, ownerInternalUID)
	if err != nil {
		return []*mgmtpb.PipelineTriggerRecord{}, 0, "", err
	}
	filter, err = s.pipelineUIDLookup(ctx, ownerID, ownerType, filter, owner, ownerInternalUID)
	if err != nil {
		return []*mgmtpb.PipelineTriggerRecord{}, 0, "", nil
	}
	pipelineTriggerRecords, ps, pt, err := s.influxDB.QueryPipelineTriggerRecords(ctx, *ownerUID, ownerQueryString, pageSize, pageToken, filter)
	if err != nil {
		return nil, 0, "", err
	}
	return pipelineTriggerRecords, ps, pt, nil
}

func (s *service) ListPipelineTriggerTableRecords(ctx context.Context, owner *mgmtpb.User, pageSize int64, pageToken string, filter filtering.Filter) ([]*mgmtpb.PipelineTriggerTableRecord, int64, string, error) {
	// Look up the owner's internal UID from their public ID
	ownerInternalUID, err := s.GetUserUIDByID(ctx, owner.Id)
	if err != nil {
		return []*mgmtpb.PipelineTriggerTableRecord{}, 0, "", err
	}

	ownerUID, ownerID, ownerType, ownerQueryString, filter, err := s.checkPipelineOwnership(ctx, filter, owner, ownerInternalUID)
	if err != nil {
		return []*mgmtpb.PipelineTriggerTableRecord{}, 0, "", err
	}

	filter, err = s.pipelineUIDLookup(ctx, ownerID, ownerType, filter, owner, ownerInternalUID)
	if err != nil {
		return []*mgmtpb.PipelineTriggerTableRecord{}, 0, "", nil
	}

	pipelineTriggerTableRecords, ps, pt, err := s.influxDB.QueryPipelineTriggerTableRecords(ctx, *ownerUID, ownerQueryString, pageSize, pageToken, filter)
	if err != nil {
		return nil, 0, "", err
	}

	return pipelineTriggerTableRecords, ps, pt, nil
}

func (s *service) ListPipelineTriggerChartRecords(
	ctx context.Context,
	req *mgmtpb.ListPipelineTriggerChartRecordsRequest,
	ctxUserUID uuid.UUID,
) (*mgmtpb.ListPipelineTriggerChartRecordsResponse, error) {
	nsUID, err := s.GrantedNamespaceUID(ctx, req.GetRequesterId(), ctxUserUID)
	if err != nil {
		return nil, fmt.Errorf("checking user permissions: %w", err)
	}

	now := time.Now().UTC()
	p := repository.ListTriggerChartRecordsParams{
		RequesterID:  req.GetRequesterId(),
		RequesterUID: nsUID,

		// Default values
		AggregationWindow: 1 * time.Hour,
		Start:             time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()),
		Stop:              now,
	}

	if req.GetAggregationWindow() != "" {
		window, err := time.ParseDuration(req.GetAggregationWindow())
		if err != nil {
			return nil, fmt.Errorf("%w: extracting duration from aggregation window: %w", errorsx.ErrInvalidArgument, err)
		}

		p.AggregationWindow = window
	}

	if req.GetStart() != nil {
		p.Start = req.GetStart().AsTime()
	}

	if req.GetStop() != nil {
		p.Stop = req.GetStop().AsTime()
	}

	return s.influxDB.ListPipelineTriggerChartRecords(ctx, p)
}

func (s *service) ListPipelineTriggerChartRecordsV0(ctx context.Context, owner *mgmtpb.User, aggregationWindow int64, filter filtering.Filter) ([]*mgmtpb.PipelineTriggerChartRecordV0, error) {
	// Look up the owner's internal UID from their public ID
	ownerInternalUID, err := s.GetUserUIDByID(ctx, owner.Id)
	if err != nil {
		return []*mgmtpb.PipelineTriggerChartRecordV0{}, err
	}

	ownerUID, ownerID, ownerType, ownerQueryString, filter, err := s.checkPipelineOwnership(ctx, filter, owner, ownerInternalUID)
	if err != nil {
		return []*mgmtpb.PipelineTriggerChartRecordV0{}, err
	}

	filter, err = s.pipelineUIDLookup(ctx, ownerID, ownerType, filter, owner, ownerInternalUID)
	if err != nil {
		return []*mgmtpb.PipelineTriggerChartRecordV0{}, nil
	}

	pipelineTriggerChartRecords, err := s.influxDB.QueryPipelineTriggerChartRecordsV0(ctx, *ownerUID, ownerQueryString, aggregationWindow, filter)
	if err != nil {
		return nil, err
	}

	return pipelineTriggerChartRecords, nil
}

func (s *service) GetPipelineTriggerCount(
	ctx context.Context,
	req *mgmtpb.GetPipelineTriggerCountRequest,
	ctxUserUID uuid.UUID,
) (*mgmtpb.GetPipelineTriggerCountResponse, error) {
	requesterUID, err := s.GrantedNamespaceUID(ctx, req.GetRequesterId(), ctxUserUID)
	if err != nil {
		return nil, fmt.Errorf("checking user permissions: %w", err)
	}

	now := time.Now().UTC()
	p := repository.GetTriggerCountParams{
		RequesterUID: requesterUID,

		// Default values
		Start: time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()),
		Stop:  now,
	}

	if req.GetStart() != nil {
		p.Start = req.GetStart().AsTime()
	}

	if req.GetStop() != nil {
		p.Stop = req.GetStop().AsTime()
	}

	return s.influxDB.GetPipelineTriggerCount(ctx, p)
}

func (s *service) GetModelTriggerCount(ctx context.Context, req *mgmtpb.GetModelTriggerCountRequest, ctxUserUID uuid.UUID) (*mgmtpb.GetModelTriggerCountResponse, error) {
	requesterUID, err := s.GrantedNamespaceUID(ctx, req.GetRequesterId(), ctxUserUID)
	if err != nil {
		return nil, fmt.Errorf("checking user permissions: %w", err)
	}

	now := time.Now().UTC()
	p := repository.GetTriggerCountParams{
		RequesterUID: requesterUID,

		// Default values
		Start: time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()),
		Stop:  now,
	}

	if req.GetStart() != nil {
		p.Start = req.GetStart().AsTime()
	}

	if req.GetStop() != nil {
		p.Stop = req.GetStop().AsTime()
	}

	return s.influxDB.GetModelTriggerCount(ctx, p)
}

func (s *service) ListModelTriggerChartRecords(
	ctx context.Context,
	req *mgmtpb.ListModelTriggerChartRecordsRequest,
	ctxUserUID uuid.UUID,
) (*mgmtpb.ListModelTriggerChartRecordsResponse, error) {
	nsUID, err := s.GrantedNamespaceUID(ctx, req.GetRequesterId(), ctxUserUID)
	if err != nil {
		return nil, fmt.Errorf("checking user permissions: %w", err)
	}

	now := time.Now().UTC()
	p := repository.ListTriggerChartRecordsParams{
		RequesterID:  req.GetRequesterId(),
		RequesterUID: nsUID,

		// Default values
		AggregationWindow: 1 * time.Hour,
		Start:             time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()),
		Stop:              now,
	}

	if req.GetAggregationWindow() != "" {
		window, err := time.ParseDuration(req.GetAggregationWindow())
		if err != nil {
			return nil, fmt.Errorf("%w: extracting duration from aggregation window: %w", errorsx.ErrInvalidArgument, err)
		}

		p.AggregationWindow = window
	}

	if req.GetStart() != nil {
		p.Start = req.GetStart().AsTime()
	}

	if req.GetStop() != nil {
		p.Stop = req.GetStop().AsTime()
	}

	return s.influxDB.ListModelTriggerChartRecords(ctx, p)
}

// GrantedNamespaceUID returns the UID of a namespace, provided the
// authenticated user has access to it.
// In CE, only the user's own namespace is accessible. Organizations are EE-only.
func (s *service) GrantedNamespaceUID(ctx context.Context, namespaceID string, authenticatedUserUID uuid.UUID) (uuid.UUID, error) {
	owner, err := s.repository.GetOwner(ctx, namespaceID, false)
	if err != nil {
		if errors.Is(err, errorsx.ErrNotFound) {
			err = errorsx.ErrUnauthorized
		}

		return uuid.Nil, fmt.Errorf("fetching namespace UID: %w", err)
	}

	nsUID := owner.UID
	if nsUID == authenticatedUserUID {
		// The authenticated user always has access to their own namespace.
		return nsUID, nil
	}

	// In CE, organizations are not supported. Users can only access their own namespace.
	return uuid.Nil, errorsx.ErrUnauthorized
}
