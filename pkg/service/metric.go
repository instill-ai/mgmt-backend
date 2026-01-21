package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid"

	"github.com/instill-ai/mgmt-backend/pkg/repository"

	mgmtpb "github.com/instill-ai/protogen-go/mgmt/v1beta"
	errorsx "github.com/instill-ai/x/errors"
)

var ErrNoPermission = errors.New("no permission")

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
