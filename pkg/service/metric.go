package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/gofrs/uuid"
	"github.com/instill-ai/mgmt-backend/pkg/acl"
	"github.com/instill-ai/mgmt-backend/pkg/repository"

	errdomain "github.com/instill-ai/mgmt-backend/pkg/errors"
	mgmtpb "github.com/instill-ai/protogen-go/core/mgmt/v1beta"
)

func (s *service) GetPipelineTriggerCount(
	ctx context.Context,
	req *mgmtpb.GetPipelineTriggerCountRequest,
	ctxUserUID uuid.UUID,
) (*mgmtpb.GetPipelineTriggerCountResponse, error) {
	nsUID, err := s.GrantedNamespaceUID(ctx, req.GetNamespaceId(), ctxUserUID)
	if err != nil {
		return nil, fmt.Errorf("checking user permissions: %w", err)
	}

	now := time.Now().UTC()
	p := repository.GetPipelineTriggerCountParams{
		NamespaceUID: nsUID,

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

func (s *service) ListPipelineTriggerChartRecords(
	ctx context.Context,
	req *mgmtpb.ListPipelineTriggerChartRecordsRequest,
	ctxUserUID uuid.UUID,
) (*mgmtpb.ListPipelineTriggerChartRecordsResponse, error) {
	nsUID, err := s.GrantedNamespaceUID(ctx, req.GetNamespaceId(), ctxUserUID)
	if err != nil {
		return nil, fmt.Errorf("checking user permissions: %w", err)
	}

	now := time.Now().UTC()
	p := repository.ListPipelineTriggerChartRecordsParams{
		NamespaceID:  req.GetNamespaceId(),
		NamespaceUID: nsUID,

		// Default values
		AggregationWindow: 1 * time.Hour,
		Start:             time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()),
		Stop:              now,
	}

	if req.GetAggregationWindow() != "" {
		window, err := time.ParseDuration(req.GetAggregationWindow())
		if err != nil {
			return nil, fmt.Errorf("%w: extracting duration from aggregation window: %w", errdomain.ErrInvalidArgument, err)
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

// GrantedNamespaceUID returns the UID of a namespace, provided the
// authenticated user has access to it.
func (s *service) GrantedNamespaceUID(ctx context.Context, namespaceID string, authenticatedUserUID uuid.UUID) (uuid.UUID, error) {
	owner, err := s.repository.GetOwner(ctx, namespaceID, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errdomain.ErrUnauthorized
		}

		return uuid.Nil, fmt.Errorf("fetching namespace UID: %w", err)
	}

	nsUID := owner.UID
	if nsUID == authenticatedUserUID {
		// The authenticated user always has access to their own namespace.
		return nsUID, nil
	}

	// The user is requesting information about other namespace: only
	// organizations that the user is a member of are allowed.
	role, err := s.GetACLClient().GetOrganizationUserMembership(ctx, nsUID, authenticatedUserUID)
	if err != nil {
		if errors.Is(err, acl.ErrMembershipNotFound) {
			err = errdomain.ErrUnauthorized
		}

		return uuid.Nil, fmt.Errorf("fetching organization membership: %w", err)
	}

	if strings.HasPrefix(role, "pending") {
		return uuid.Nil, fmt.Errorf("invalid permission role: %w", errdomain.ErrUnauthorized)
	}

	return nsUID, nil
}
