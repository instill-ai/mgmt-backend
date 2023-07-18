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
)

// DefaultPageSize is the default pagination page size when page size is not assigned
const DefaultPageSize = 100

// MaxPageSize is the maximum pagination page size if the assigned value is over this number
const MaxPageSize = 1000

// Default aggregate window
var defaultAggregationWindow = time.Hour.Nanoseconds()

// InfluxDB interface
type InfluxDB interface {
	QueryPipelineTriggerRecords(ctx context.Context, owner string, pageSize int64, pageToken string, filter filtering.Filter) (pipelines []*mgmtPB.PipelineTriggerRecord, totalSize int64, nextPageToken string, err error)
	QueryPipelineTriggerChartRecords(ctx context.Context, owner string, aggregationWindow int64, filter filtering.Filter) (records []*mgmtPB.PipelineTriggerChartRecord, err error)
	QueryConnectorExecuteRecords(ctx context.Context, owner string, pageSize int64, pageToken string, filter filtering.Filter) (records []*mgmtPB.ConnectorExecuteRecord, totalSize int64, nextPageToken string, err error)
	QueryConnectorExecuteChartRecords(ctx context.Context, owner string, aggregationWindow int64, filter filtering.Filter) (records []*mgmtPB.ConnectorExecuteChartRecord, err error)
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

//TODO: reuse duplicate codes

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
			|> filter(fn: (r) => r["owner_uid"] == "%v")
			%v
			|> group()
			|> limit(n: %v)`,
		i.bucket,
		start,
		stop,
		owner,
		expr,
		pageSize,
	)

	totalQuery := fmt.Sprintf(
		`from(bucket: "%v")
			|> range(start: %v, stop: %v)
			|> filter(fn: (r) => r["_measurement"] == "pipeline.trigger")
			|> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")
			|> filter(fn: (r) => r["owner_uid"] == "%v")
			%v
			|> group()
			|> count(column: "pipeline_trigger_id")`,
		i.bucket,
		start,
		stop,
		owner,
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
					PipelineId:          result.Record().ValueByKey("pipeline_id").(string),
					PipelineUid:         result.Record().ValueByKey("pipeline_uid").(string),
					TriggerMode:         mgmtPB.Mode(mgmtPB.Mode_value[result.Record().ValueByKey("trigger_mode").(string)]),
					ComputeTimeDuration: float32(result.Record().ValueByKey("compute_time_duration").(float64)),
					Status:              mgmtPB.Status(mgmtPB.Status_value[result.Record().ValueByKey("status").(string)]),
				},
			)
		}
		// Check for an error
		if result.Err() != nil {
			return nil, 0, "", status.Errorf(codes.InvalidArgument, "Invalid query: %s", err.Error())
		}
		if result.Record() == nil {
			return nil, 0, "", nil
		}

		lastTimestamp = result.Record().Time()
	}

	if int64(len(records)) < total {
		pageToken = paginate.EncodeToken(lastTimestamp, owner)
	} else {
		pageToken = ""
	}

	return records, int64(len(records)), pageToken, nil
}

func (i *influxDB) QueryPipelineTriggerChartRecords(ctx context.Context, owner string, aggregationWindow int64, filter filtering.Filter) (records []*mgmtPB.PipelineTriggerChartRecord, err error) {

	logger, _ := logger.GetZapLogger(ctx)

	start := time.Time{}.Format(time.RFC3339Nano)
	stop := time.Now().Format(time.RFC3339Nano)

	if aggregationWindow < time.Minute.Nanoseconds() {
		aggregationWindow = defaultAggregationWindow
	}

	// TODO: design better filter expression to flux transpiler
	expr, err := i.transpileFilter(filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
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

	query := fmt.Sprintf(
		`t1 = from(bucket: "%v")
			|> range(start: %v, stop: %v)
			|> filter(fn: (r) => r["_measurement"] == "pipeline.trigger")
			|> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")
			|> filter(fn: (r) => r["owner_uid"] == "%v")
			%v
			|> group(columns: ["pipeline_id", "pipeline_uid", "trigger_mode", "status"])
			|> aggregateWindow(every: duration(v: %v), column: "trigger_time", fn: count, createEmpty: false)
		t2 = from(bucket: "%v")
			|> range(start: %v, stop: %v)
			|> filter(fn: (r) => r["_measurement"] == "pipeline.trigger")
			|> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")
			|> filter(fn: (r) => r["owner_uid"] == "%v")
			%v
			|> group(columns: ["pipeline_id", "pipeline_uid", "trigger_mode", "status"])
			|> aggregateWindow(every: duration(v: %v), fn: sum, column: "compute_time_duration", createEmpty: false)
		join(tables: {t1: t1, t2:t2}, on: ["_start", "_stop", "_time", "pipeline_id", "pipeline_uid", "trigger_mode", "status"])`,
		i.bucket,
		start,
		stop,
		owner,
		expr,
		aggregationWindow,
		i.bucket,
		start,
		stop,
		owner,
		expr,
		aggregationWindow,
	)

	result, err := i.queryAPI.Query(ctx, query)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid query: %s", err.Error())
	} else {
		var chartRecord *mgmtPB.PipelineTriggerChartRecord
		var currentTablePosition = -1
		// Iterate over query response
		for result.Next() {
			// Notice when group key has changed
			if result.TableChanged() {
				logger.Debug(fmt.Sprintf("table: %s\n", result.TableMetadata().String()))
			}

			if result.Record().Table() != currentTablePosition {
				chartRecord = &mgmtPB.PipelineTriggerChartRecord{
					PipelineId:          result.Record().ValueByKey("pipeline_id").(string),
					PipelineUid:         result.Record().ValueByKey("pipeline_uid").(string),
					TriggerMode:         mgmtPB.Mode(mgmtPB.Mode_value[result.Record().ValueByKey("trigger_mode").(string)]),
					Status:              mgmtPB.Status(mgmtPB.Status_value[result.Record().ValueByKey("status").(string)]),
					TimeBuckets:         []*timestamppb.Timestamp{},
					TriggerCounts:       []int64{},
					ComputeTimeDuration: []float32{},
				}
				records = append(records, chartRecord)
				currentTablePosition = result.Record().Table()
			}

			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "Invalid parse key: %s", err.Error())
			}
			chartRecord.TimeBuckets = append(chartRecord.TimeBuckets, timestamppb.New(result.Record().ValueByKey("_time").(time.Time)))
			chartRecord.TriggerCounts = append(chartRecord.TriggerCounts, result.Record().ValueByKey("trigger_time").(int64))
			chartRecord.ComputeTimeDuration = append(chartRecord.ComputeTimeDuration, float32(result.Record().ValueByKey("compute_time_duration").(float64)))
		}
		// Check for an error
		if result.Err() != nil {
			return nil, status.Errorf(codes.InvalidArgument, "Invalid query: %s", err.Error())
		}
		if result.Record() == nil {
			return nil, nil
		}
	}

	return records, nil
}

func (i *influxDB) QueryConnectorExecuteRecords(ctx context.Context, owner string, pageSize int64, pageToken string, filter filtering.Filter) (records []*mgmtPB.ConnectorExecuteRecord, totalSize int64, nextPageToken string, err error) {

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
			|> filter(fn: (r) => r["_measurement"] == "connector.execute")
			|> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")
			|> filter(fn: (r) => r["connector_owner_uid"] == "%v")
			%v
			|> group()
			|> limit(n: %v)`,
		i.bucket,
		start,
		stop,
		owner,
		expr,
		pageSize,
	)

	totalQuery := fmt.Sprintf(
		`from(bucket: "%v")
			|> range(start: %v, stop: %v)
			|> filter(fn: (r) => r["_measurement"] == "connector.execute")
			|> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")
			|> filter(fn: (r) => r["connector_owner_uid"] == "%v")
			%v
			|> group()
			|> count(column: "connector_execute_id")`,
		i.bucket,
		start,
		stop,
		owner,
		expr,
	)

	totalQueryResult, err := i.queryAPI.Query(ctx, totalQuery)
	total := int64(0)
	if err != nil {
		return nil, 0, "", status.Errorf(codes.InvalidArgument, "Invalid query: %s", err.Error())
	} else {
		if totalQueryResult.Next() {
			total = totalQueryResult.Record().ValueByKey("connector_execute_id").(int64)
		}
	}

	result, err := i.queryAPI.Query(ctx, query)
	var lastTimestamp time.Time
	if err != nil {
		return nil, 0, "", status.Errorf(codes.InvalidArgument, "Invalid query: %s", err.Error())
	} else {
		// Iterate over query response
		for result.Next() {
			// Notice when group key has changed
			if result.TableChanged() {
				logger.Debug(fmt.Sprintf("table: %s\n", result.TableMetadata().String()))
			}

			executeTime, err := time.Parse(time.RFC3339Nano, result.Record().ValueByKey("execute_time").(string))
			if err != nil {
				return nil, 0, "", status.Errorf(codes.InvalidArgument, "Invalid parse key: %s", err.Error())
			}
			records = append(
				records,
				&mgmtPB.ConnectorExecuteRecord{
					ExecuteTime:            timestamppb.New(executeTime),
					ConnectorExecuteId:     result.Record().ValueByKey("connector_execute_id").(string),
					ConnectorId:            result.Record().ValueByKey("connector_id").(string),
					ConnectorUid:           result.Record().ValueByKey("connector_uid").(string),
					ConnectorDefinitionUid: result.Record().ValueByKey("connector_definition_uid").(string),
					PipelineTriggerId:      result.Record().ValueByKey("pipeline_trigger_id").(string),
					PipelineId:             result.Record().ValueByKey("pipeline_id").(string),
					PipelineUid:            result.Record().ValueByKey("pipeline_uid").(string),
					ComputeTimeDuration:    float32(result.Record().ValueByKey("compute_time_duration").(float64)),
					Status:                 mgmtPB.Status(mgmtPB.Status_value[result.Record().ValueByKey("status").(string)]),
				},
			)
		}
		// Check for an error
		if result.Err() != nil {
			return nil, 0, "", status.Errorf(codes.InvalidArgument, "Invalid query: %s", err.Error())
		}
		if result.Record() == nil {
			return nil, 0, "", nil
		}

		lastTimestamp = result.Record().Time()
	}

	if int64(len(records)) < total {
		pageToken = paginate.EncodeToken(lastTimestamp, owner)
	} else {
		pageToken = ""
	}

	return records, int64(len(records)), pageToken, nil
}

func (i *influxDB) QueryConnectorExecuteChartRecords(ctx context.Context, owner string, aggregationWindow int64, filter filtering.Filter) (records []*mgmtPB.ConnectorExecuteChartRecord, err error) {

	logger, _ := logger.GetZapLogger(ctx)

	start := time.Time{}.Format(time.RFC3339Nano)
	stop := time.Now().Format(time.RFC3339Nano)

	if aggregationWindow < time.Minute.Nanoseconds() {
		aggregationWindow = defaultAggregationWindow
	}

	// TODO: design better filter expression to flux transpiler
	expr, err := i.transpileFilter(filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
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

	query := fmt.Sprintf(
		`t1 = from(bucket: "%v")
			|> range(start: %v, stop: %v)
			|> filter(fn: (r) => r["_measurement"] == "connector.execute")
			|> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")
			|> filter(fn: (r) => r["connector_owner_uid"] == "%v")
			%v
			|> group(columns: ["connector_id", "connector_uid", "status"])
			|> aggregateWindow(every: duration(v: %v), column: "execute_time", fn: count, createEmpty: false)
		t2 = from(bucket: "%v")
			|> range(start: %v, stop: %v)
			|> filter(fn: (r) => r["_measurement"] == "connector.execute")
			|> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")
			|> filter(fn: (r) => r["connector_owner_uid"] == "%v")
			%v
			|> group(columns: ["connector_id", "connector_uid", "status"])
			|> aggregateWindow(every: duration(v: %v), fn: sum, column: "compute_time_duration", createEmpty: false)
		join(tables: {t1: t1, t2:t2}, on: ["_start", "_stop", "_time", "connector_id", "connector_uid", "status"])`,
		i.bucket,
		start,
		stop,
		owner,
		expr,
		aggregationWindow,
		i.bucket,
		start,
		stop,
		owner,
		expr,
		aggregationWindow,
	)

	result, err := i.queryAPI.Query(ctx, query)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid query: %s", err.Error())
	} else {
		var chartRecord *mgmtPB.ConnectorExecuteChartRecord
		var currentTablePosition = -1
		// Iterate over query response
		for result.Next() {
			// Notice when group key has changed
			if result.TableChanged() {
				logger.Debug(fmt.Sprintf("table: %s\n", result.TableMetadata().String()))
			}

			if result.Record().Table() != currentTablePosition {
				chartRecord = &mgmtPB.ConnectorExecuteChartRecord{
					ConnectorId:         result.Record().ValueByKey("connector_id").(string),
					ConnectorUid:        result.Record().ValueByKey("connector_uid").(string),
					Status:              mgmtPB.Status(mgmtPB.Status_value[result.Record().ValueByKey("status").(string)]),
					TimeBuckets:         []*timestamppb.Timestamp{},
					ExecuteCounts:       []int64{},
					ComputeTimeDuration: []float32{},
				}
				records = append(records, chartRecord)
				currentTablePosition = result.Record().Table()
			}

			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "Invalid parse key: %s", err.Error())
			}
			chartRecord.TimeBuckets = append(chartRecord.TimeBuckets, timestamppb.New(result.Record().ValueByKey("_time").(time.Time)))
			chartRecord.ExecuteCounts = append(chartRecord.ExecuteCounts, result.Record().ValueByKey("execute_time").(int64))
			chartRecord.ComputeTimeDuration = append(chartRecord.ComputeTimeDuration, float32(result.Record().ValueByKey("compute_time_duration").(float64)))
		}
		// Check for an error
		if result.Err() != nil {
			return nil, status.Errorf(codes.InvalidArgument, "Invalid query: %s", err.Error())
		}
		if result.Record() == nil {
			return nil, nil
		}
	}

	return records, nil
}

// TranspileFilter transpiles a parsed AIP filter expression to Flux query expression
func (i *influxDB) transpileFilter(filter filtering.Filter) (string, error) {
	return (&Transpiler{
		filter: filter,
	}).Transpile()
}
