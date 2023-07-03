package repository

import (
	"context"
	"fmt"
	"strings"

	// "strconv"
	"time"

	"github.com/influxdata/influxdb-client-go/v2/api"
	"go.einride.tech/aip/filtering"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/instill-ai/mgmt-backend/pkg/logger"
	"github.com/instill-ai/x/paginate"

	mgmtPB "github.com/instill-ai/protogen-go/base/mgmt/v1alpha"
	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1alpha"
)

// DefaultPageSize is the default pagination page size when page size is not assigned
const DefaultPageSize = 100

// MaxPageSize is the maximum pagination page size if the assigned value is over this number
const MaxPageSize = 1000

// InfluxDB interface
type InfluxDB interface {
	QueryPipelineTriggerRecords(ctx context.Context, owner string, pageSize int64, pageToken string, filter filtering.Filter) (pipelines []*mgmtPB.PipelineTriggerRecord, totalSize int64, nextPageToken string, err error)
}

type influxDB struct {
	queryAPI api.QueryAPI
	bucket   string
}

// NewInfluxDB initiates a repository instance
func NewInfluxDB(queryAPI api.QueryAPI, bucket string) InfluxDB {
	return &influxDB{
		queryAPI: queryAPI,
		bucket:   bucket,
	}
}

func (i *influxDB) QueryPipelineTriggerRecords(ctx context.Context, owner string, pageSize int64, pageToken string, filter filtering.Filter) (records []*mgmtPB.PipelineTriggerRecord, totalSize int64, nextPageToken string, err error) {

	logger, _ := logger.GetZapLogger(ctx)

	if pageSize == 0 {
		pageSize = DefaultPageSize
	} else if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	start := time.Time{}.Format(time.RFC3339Nano)
	stop := time.Now().Format(time.RFC3339Nano)

	// TODO: design better filter expression to flux transpiler
	expr, err := i.transpileFilter(filter)
	if err != nil {
		return nil, 0, "", status.Errorf(codes.Internal, err.Error())
	}

	if expr != "" {
		exprs := strings.Split(expr, "&&")

		iTime := 0
		for _, expr := range exprs {
			if strings.HasPrefix(expr, "start") {
				start = strings.Split(expr, "@")[1]
				iTime += 1
			}
			if strings.HasPrefix(expr, "stop") {
				stop = strings.Split(expr, "@")[1]
				iTime += 1
			}
		}

		expr = strings.Join(exprs[iTime:], "")
	}

	if pageToken != "" {
		startTime, _, err := paginate.DecodeToken(pageToken)
		if err != nil {
			return nil, 0, "", status.Errorf(codes.InvalidArgument, "Invalid page token: %s", err.Error())
		}
		startTime = startTime.Add(time.Duration(1))
		start = startTime.Format(time.RFC3339Nano)
	}

	query := fmt.Sprintf(
		`from(bucket: "%v")
			|> range(start: %v, stop: %v)
			|> filter(fn: (r) => r["_measurement"] == "pipeline.trigger")
			|> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")
			%v
			|> group()
			|> limit(n: %v)`,
		i.bucket,
		start,
		stop,
		expr,
		pageSize,
	)

	totalQuery := fmt.Sprintf(
		`from(bucket: "%v")
			|> range(start: %v, stop: %v)
			|> filter(fn: (r) => r["_measurement"] == "pipeline.trigger")
			|> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")
			%v
			|> group()
			|> count(column: "pipeline_trigger_id")`,
		i.bucket,
		start,
		stop,
		expr,
	)

	totalQueryResult, err := i.queryAPI.Query(ctx, totalQuery)
	total := int64(0)
	if err != nil {
		return nil, 0, "", status.Errorf(codes.InvalidArgument, "Invalid query: %s", err.Error())
	} else {
		if totalQueryResult.Next() {
			total = totalQueryResult.Record().ValueByKey("pipeline_trigger_id").(int64)
		}
	}

	result, err := i.queryAPI.Query(ctx, query)
	var lastTimestamp time.Time
	var ownerUUID string
	if err != nil {
		return nil, 0, "", status.Errorf(codes.InvalidArgument, "Invalid query: %s", err.Error())
	} else {
		// Iterate over query response
		for result.Next() {
			// Notice when group key has changed
			if result.TableChanged() {
				logger.Debug(fmt.Sprintf("table: %s\n", result.TableMetadata().String()))
			}

			triggerTime, err := time.Parse(time.RFC3339Nano, result.Record().ValueByKey("trigger_time").(string))
			if err != nil {
				return nil, 0, "", status.Errorf(codes.InvalidArgument, "Invalid parse key: %s", err.Error())
			}
			records = append(
				records,
				&mgmtPB.PipelineTriggerRecord{
					TriggerTime:         timestamppb.New(triggerTime),
					PipelineTriggerId:   result.Record().ValueByKey("pipeline_trigger_id").(string),
					PipelineName:        result.Record().ValueByKey("pipeline_name").(string),
					PipelinePermalink:   result.Record().ValueByKey("pipeline_permalink").(string),
					PipelineMode:        pipelinePB.Pipeline_Mode(pipelinePB.Pipeline_Mode_value[result.Record().ValueByKey("pipeline_mode").(string)]),
					ComputeTimeDuration: float32(result.Record().ValueByKey("compute_time_duration").(float64)),
					Status:              result.Record().ValueByKey("status").(string),
				},
			)
		}
		// Check for an error
		if result.Err() != nil {
			return nil, 0, "", status.Errorf(codes.InvalidArgument, "Invalid query: %s", err.Error())
		}
		if result.Record() == nil {
			return nil, 0, "", status.Errorf(codes.NotFound, "Empty query result")
		}

		lastTimestamp = result.Record().Time()
		ownerUUID = result.Record().ValueByKey("owner_uid").(string)
	}

	if int64(len(records)) < total {
		pageToken = paginate.EncodeToken(lastTimestamp, ownerUUID)
	} else {
		pageToken = ""
	}

	return records, int64(len(records)), pageToken, nil
}

// TranspileFilter transpiles a parsed AIP filter expression to GORM DB clauses
func (i *influxDB) transpileFilter(filter filtering.Filter) (string, error) {
	return (&Transpiler{
		filter: filter,
	}).Transpile()
}
