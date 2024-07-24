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

func InjectOwnerToContext(ctx context.Context, uid string) context.Context {
	ctx = metadata.AppendToOutgoingContext(ctx, "instill-auth-type", "user")
	ctx = metadata.AppendToOutgoingContext(ctx, "instill-user-uid", uid)
	return ctx
}

func (s *service) checkPipelineOwnership(ctx context.Context, filter filtering.Filter, owner *mgmtPB.User) (*string, string, string, string, filtering.Filter, error) {
	ownerID := owner.Id
	ownerUID := owner.Uid
	ownerType := "users"
	ownerQueryString := ""

	if len(filter.CheckedExpr.GetExpr().GetCallExpr().GetArgs()) > 0 {

		ownerName, _ := repository.ExtractConstExpr(filter.CheckedExpr.GetExpr(), constant.OwnerName, false)

		if ownerName != "" {

			if strings.HasPrefix(ownerName, "users") {
				if ownerName != fmt.Sprintf("users/%s", owner.Id) {
					return nil, "", "", "", filter, ErrNoPermission
				}
				repository.HijackConstExpr(filter.CheckedExpr.GetExpr(), constant.OwnerName, constant.PipelineOwnerUID, *owner.Uid, false)
			} else if strings.HasPrefix(ownerName, "organizations") {
				ownerType = "organizations"
				id, err := resource.GetRscNameID(ownerName)
				if err != nil {
					return nil, "", "", "", filter, err
				}
				ownerID = id
				org, err := s.GetOrganizationAdmin(ctx, id)
				if err != nil {
					return nil, "", "", "", filter, err
				}
				granted, err := s.GetACLClient().CheckPermission(ctx, "organization", uuid.FromStringOrNil(org.Uid), "user", uuid.FromStringOrNil(owner.GetUid()), "", "member")
				if err != nil {
					return nil, "", "", "", filter, err
				}
				if !granted {
					return nil, "", "", "", filter, ErrNoPermission
				}
				repository.HijackConstExpr(filter.CheckedExpr.GetExpr(), constant.OwnerName, constant.PipelineOwnerUID, org.Uid, false)
				ownerUID = &org.Uid
			} else {
				return nil, "", "", "", filter, ErrInvalidOwnerNamespace
			}
		} else {
			ownerQueryString = fmt.Sprintf("|> filter(fn: (r) => r[\"owner_uid\"] == \"%v\")", *owner.Uid)
		}
	} else {
		ownerQueryString = fmt.Sprintf("|> filter(fn: (r) => r[\"owner_uid\"] == \"%v\")", *owner.Uid)
	}
	return ownerUID, ownerID, ownerType, ownerQueryString, filter, nil
}

func (s *service) pipelineUIDLookup(ctx context.Context, ownerID string, ownerType string, filter filtering.Filter, owner *mgmtPB.User) (filtering.Filter, error) {

	ctx = InjectOwnerToContext(ctx, *owner.Uid)

	// lookup pipeline uid
	if len(filter.CheckedExpr.GetExpr().GetCallExpr().GetArgs()) > 0 {
		pipelineID, _ := repository.ExtractConstExpr(filter.CheckedExpr.GetExpr(), constant.PipelineID, false)

		if pipelineID != "" {
			if ownerType == "users" {
				//nolint:staticcheck
				if respPipeline, err := s.pipelinePublicServiceClient.GetUserPipeline(ctx, &pipelinePB.GetUserPipelineRequest{
					Name: fmt.Sprintf("%s/pipelines/%s", owner.Name, pipelineID),
				}); err != nil {
					return filter, err
				} else {
					repository.HijackConstExpr(filter.CheckedExpr.GetExpr(), constant.PipelineID, constant.PipelineUID, respPipeline.Pipeline.Uid, false)

					// lookup pipeline release uid
					pipelineReleaseID, _ := repository.ExtractConstExpr(filter.CheckedExpr.GetExpr(), constant.PipelineReleaseID, false)

					//nolint:staticcheck
					respPipelineRelease, err := s.pipelinePublicServiceClient.GetUserPipelineRelease(ctx, &pipelinePB.GetUserPipelineReleaseRequest{
						Name: fmt.Sprintf("%s/pipelines/%s/releases/%s", owner.Name, pipelineID, pipelineReleaseID),
					})
					if err == nil {
						repository.HijackConstExpr(filter.CheckedExpr.GetExpr(), constant.PipelineID, constant.PipelineUID, respPipelineRelease.Release.Uid, false)
					}
				}
			} else if ownerType == "organizations" {
				//nolint:staticcheck
				if respPipeline, err := s.pipelinePublicServiceClient.GetOrganizationPipeline(ctx, &pipelinePB.GetOrganizationPipelineRequest{
					Name: fmt.Sprintf("organizations/%s/pipelines/%s", ownerID, pipelineID),
				}); err != nil {
					return filter, err
				} else {
					repository.HijackConstExpr(filter.CheckedExpr.GetExpr(), constant.PipelineID, constant.PipelineUID, respPipeline.Pipeline.Uid, false)

					// lookup pipeline release uid
					pipelineReleaseID, _ := repository.ExtractConstExpr(filter.CheckedExpr.GetExpr(), constant.PipelineReleaseID, false)

					//nolint:staticcheck
					respPipelineRelease, err := s.pipelinePublicServiceClient.GetUserPipelineRelease(ctx, &pipelinePB.GetUserPipelineReleaseRequest{
						Name: fmt.Sprintf("organizations/%s/pipelines/%s/releases/%s", ownerID, pipelineID, pipelineReleaseID),
					})
					if err == nil {
						repository.HijackConstExpr(filter.CheckedExpr.GetExpr(), constant.PipelineID, constant.PipelineUID, respPipelineRelease.Release.Uid, false)
					}
				}
			}
		}
	}

	return filter, nil
}

func (s *service) ListPipelineTriggerTableRecords(ctx context.Context, owner *mgmtPB.User, pageSize int64, pageToken string, filter filtering.Filter) ([]*mgmtPB.PipelineTriggerTableRecord, int64, string, error) {

	ownerUID, ownerID, ownerType, ownerQueryString, filter, err := s.checkPipelineOwnership(ctx, filter, owner)
	if err != nil {
		return []*mgmtPB.PipelineTriggerTableRecord{}, 0, "", err
	}

	filter, err = s.pipelineUIDLookup(ctx, ownerID, ownerType, filter, owner)
	if err != nil {
		return []*mgmtPB.PipelineTriggerTableRecord{}, 0, "", nil
	}

	pipelineTriggerTableRecords, ps, pt, err := s.influxDB.QueryPipelineTriggerTableRecords(ctx, *ownerUID, ownerQueryString, pageSize, pageToken, filter)
	if err != nil {
		return nil, 0, "", err
	}

	return pipelineTriggerTableRecords, ps, pt, nil
}
