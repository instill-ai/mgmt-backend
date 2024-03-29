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

	"github.com/instill-ai/mgmt-backend/pkg/constant"
	"github.com/instill-ai/mgmt-backend/pkg/logger"
	"github.com/instill-ai/x/paginate"

	mgmtPB "github.com/instill-ai/protogen-go/core/mgmt/v1beta"
)

// DefaultPageSize is the default pagination page size when page size is not assigned
const DefaultPageSize = 100

// MaxPageSize is the maximum pagination page size if the assigned value is over this number
const MaxPageSize = 1000

// Default aggregate window
var defaultAggregationWindow = time.Hour.Nanoseconds()

// InfluxDB interface
type InfluxDB interface {
	QueryPipelineTriggerRecords(ctx context.Context, owner string, ownerQueryString string, pageSize int64, pageToken string, filter filtering.Filter) (pipelines []*mgmtPB.PipelineTriggerRecord, totalSize int64, nextPageToken string, err error)
	QueryPipelineTriggerTableRecords(ctx context.Context, owner string, ownerQueryString string, pageSize int64, pageToken string, filter filtering.Filter) (records []*mgmtPB.PipelineTriggerTableRecord, totalSize int64, nextPageToken string, err error)
	QueryPipelineTriggerChartRecords(ctx context.Context, owner string, ownerQueryString string, aggregationWindow int64, filter filtering.Filter) (records []*mgmtPB.PipelineTriggerChartRecord, err error)
	QueryConnectorExecuteRecords(ctx context.Context, owner string, ownerQueryString string, pageSize int64, pageToken string, filter filtering.Filter) (records []*mgmtPB.ConnectorExecuteRecord, totalSize int64, nextPageToken string, err error)
	QueryConnectorExecuteTableRecords(ctx context.Context, owner string, ownerQueryString string, pageSize int64, pageToken string, filter filtering.Filter) (records []*mgmtPB.ConnectorExecuteTableRecord, totalSize int64, nextPageToken string, err error)
	QueryConnectorExecuteChartRecords(ctx context.Context, owner string, ownerQueryString string, aggregationWindow int64, filter filtering.Filter) (records []*mgmtPB.ConnectorExecuteChartRecord, err error)
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

func (i *influxDB) constructRecordQuery(
	ctx context.Context,
	ownerQueryString string,
	pageSize int64,
	pageToken string,
	filter filtering.Filter,
	measurement string,
	sortKey string,
) (query string, total int64, err error) {

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
		return "", 0, status.Errorf(codes.Internal, err.Error())
	}

	if expr != "" {
		exprs := strings.Split(expr, "&&")
		for i, expr := range exprs {
			if strings.HasPrefix(expr, constant.Start) {
				start = strings.Split(expr, "@")[1]
				exprs[i] = ""
			}
			if strings.HasPrefix(expr, constant.Stop) {
				stop = strings.Split(expr, "@")[1]
				exprs[i] = ""
			}
		}
		expr = strings.Join(exprs, "")
	}

	if pageToken != "" {
		startTime, _, err := paginate.DecodeToken(pageToken)
		if err != nil {
			return "", 0, status.Errorf(codes.InvalidArgument, "Invalid page token: %s", err.Error())
		}
		startTime = startTime.Add(time.Duration(1))
		start = startTime.Format(time.RFC3339Nano)
	}

	baseQuery := fmt.Sprintf(
		`from(bucket: "%v")
			|> range(start: %v, stop: %v)
			|> filter(fn: (r) => r["_measurement"] == "%v")
			|> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")
			%v
			%v
			|> group()
			|> sort(columns: ["%v"])`,
		i.bucket,
		start,
		stop,
		measurement,
		ownerQueryString,
		expr,
		sortKey,
	)

	query = fmt.Sprintf(
		`%v
		|> limit(n: %v)`,
		baseQuery,
		pageSize,
	)

	totalQuery := fmt.Sprintf(
		`%v
		|> count(column: "%v")`,
		baseQuery,
		sortKey,
	)

	totalQueryResult, err := i.queryAPI.Query(ctx, totalQuery)

	if err != nil {
		return "", 0, status.Errorf(codes.InvalidArgument, "Invalid query: %s", err.Error())
	} else {
		if totalQueryResult.Next() {
			total = totalQueryResult.Record().ValueByKey(sortKey).(int64)
		}
	}

	return query, total, nil
}

func (i *influxDB) QueryPipelineTriggerRecords(ctx context.Context, owner string, ownerQueryString string, pageSize int64, pageToken string, filter filtering.Filter) (records []*mgmtPB.PipelineTriggerRecord, totalSize int64, nextPageToken string, err error) {

	logger, _ := logger.GetZapLogger(ctx)

	query, total, err := i.constructRecordQuery(ctx, ownerQueryString, pageSize, pageToken, filter, constant.PipelineTriggerMeasurement, constant.TriggerTime)
	if err != nil {
		return nil, 0, "", status.Errorf(codes.InvalidArgument, "Invalid query: %s", err.Error())
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

			record := &mgmtPB.PipelineTriggerRecord{}

			if v, match := result.Record().ValueByKey(constant.TriggerTime).(string); match {
				triggerTime, err := time.Parse(time.RFC3339Nano, v)
				if err != nil {
					return nil, 0, "", status.Errorf(codes.InvalidArgument, "Invalid parse key: %s", err.Error())
				}
				record.TriggerTime = timestamppb.New(triggerTime)
			}
			if v, match := result.Record().ValueByKey(constant.PipelineTriggerID).(string); match {
				record.PipelineTriggerId = v
			}
			if v, match := result.Record().ValueByKey(constant.PipelineID).(string); match {
				record.PipelineId = v
			}
			if v, match := result.Record().ValueByKey(constant.PipelineUID).(string); match {
				record.PipelineUid = v
			}
			if v, match := result.Record().ValueByKey(constant.PipelineReleaseID).(string); match {
				record.PipelineReleaseId = v
			}
			if v, match := result.Record().ValueByKey(constant.PipelineReleaseUID).(string); match {
				record.PipelineReleaseUid = v
			}
			if v, match := result.Record().ValueByKey(constant.TriggerMode).(string); match {
				record.TriggerMode = mgmtPB.Mode(mgmtPB.Mode_value[v])
			}
			if v, match := result.Record().ValueByKey(constant.ComputeTimeDuration).(float64); match {
				record.ComputeTimeDuration = float32(v)
			}
			// TODO: temporary solution for legacy data format, currently there is no way to update the tags in influxdb
			if v, match := result.Record().ValueByKey(constant.Status).(string); match {
				if v == constant.Completed {
					record.Status = mgmtPB.Status_STATUS_COMPLETED
				} else if v == constant.Errored {
					record.Status = mgmtPB.Status_STATUS_ERRORED
				} else {
					record.Status = mgmtPB.Status(mgmtPB.Status_value[v])
				}
			}
			if v, match := result.Record().ValueByKey(constant.PipelineMode).(string); match {
				record.TriggerMode = mgmtPB.Mode(mgmtPB.Mode_value[v])
			}

			records = append(records, record)
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

func (i *influxDB) QueryPipelineTriggerTableRecords(ctx context.Context, owner string, ownerQueryString string, pageSize int64, pageToken string, filter filtering.Filter) (records []*mgmtPB.PipelineTriggerTableRecord, totalSize int64, nextPageToken string, err error) {

	logger, _ := logger.GetZapLogger(ctx)

	if pageSize == 0 {
		pageSize = DefaultPageSize
	} else if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	start := time.Time{}.Format(time.RFC3339Nano)
	stop := time.Now().Format(time.RFC3339Nano)
	mostRecetTimeFilter := time.Now().Format(time.RFC3339Nano)

	// TODO: validate owner uid from token
	if pageToken != "" {
		mostRecetTime, _, err := paginate.DecodeToken(pageToken)
		if err != nil {
			return nil, 0, "", status.Errorf(codes.InvalidArgument, "Invalid page token: %s", err.Error())
		}
		mostRecetTime = mostRecetTime.Add(time.Duration(-1))
		mostRecetTimeFilter = mostRecetTime.Format(time.RFC3339Nano)
	}

	// TODO: design better filter expression to flux transpiler
	expr, err := i.transpileFilter(filter)
	if err != nil {
		return nil, 0, "", status.Errorf(codes.Internal, err.Error())
	}

	if expr != "" {
		exprs := strings.Split(expr, "&&")
		for i, expr := range exprs {
			if strings.HasPrefix(expr, constant.Start) {
				start = strings.Split(expr, "@")[1]
				exprs[i] = ""
			}
			if strings.HasPrefix(expr, constant.Stop) {
				stop = strings.Split(expr, "@")[1]
				exprs[i] = ""
			}
		}
		expr = strings.Join(exprs, "")
	}

	baseQuery := fmt.Sprintf(
		`base =
			from(bucket: "%v")
				|> range(start: %v, stop: %v)
				|> filter(fn: (r) => r["_measurement"] == "pipeline.trigger")
				|> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")
				%v
				%v
		triggerRank =
			base
				|> drop(
					columns: [
						"owner_uid",
						"trigger_mode",
						"compute_time_duration",
						"pipeline_trigger_id",
						"status",
					],
				)
				|> group(columns: ["pipeline_uid"])
				|> map(fn: (r) => ({r with trigger_time: time(v: r.trigger_time)}))
				|> max(column: "trigger_time")
				|> rename(columns: {trigger_time: "most_recent_trigger_time"})
		triggerCount =
			base
				|> drop(
					columns: ["owner_uid", "trigger_mode", "compute_time_duration", "pipeline_trigger_id"],
				)
				|> group(columns: ["pipeline_uid", "status"])
				|> count(column: "trigger_time")
				|> rename(columns: {trigger_time: "trigger_count"})
				|> group(columns: ["pipeline_uid"])
		triggerTable =
			join(tables: {t1: triggerRank, t2: triggerCount}, on: ["pipeline_uid"])
				|> group()
				|> pivot(
					rowKey: ["pipeline_uid", "most_recent_trigger_time"],
					columnKey: ["status"],
					valueColumn: "trigger_count",
				)
				|> filter(
					fn: (r) => r["most_recent_trigger_time"] < time(v: %v)
				)
		nameMap =
			base
				|> keep(columns: ["trigger_time", "pipeline_id", "pipeline_uid"])
				|> group(columns: ["pipeline_uid"])
				|> top(columns: ["trigger_time"], n: 1)
				|> drop(columns: ["trigger_time"])
				|> group()
		join(tables: {t1: triggerTable, t2: nameMap}, on: ["pipeline_uid"])`,
		i.bucket,
		start,
		stop,
		ownerQueryString,
		expr,
		mostRecetTimeFilter,
	)

	query := fmt.Sprintf(
		`%v
		|> group()
		|> sort(columns: ["most_recent_trigger_time"], desc: true)
		|> limit(n: %v)`,
		baseQuery,
		pageSize,
	)

	totalQuery := fmt.Sprintf(
		`%v
		|> group()
		|> count(column: "pipeline_uid")`,
		baseQuery,
	)

	var lastTimestamp time.Time

	result, err := i.queryAPI.Query(ctx, query)
	if err != nil {
		return nil, 0, "", status.Errorf(codes.InvalidArgument, "Invalid query: %s", err.Error())
	} else {
		// Iterate over query response
		for result.Next() {
			// Notice when group key has changed
			if result.TableChanged() {
				logger.Debug(fmt.Sprintf("table: %s\n", result.TableMetadata().String()))
			}

			tableRecord := &mgmtPB.PipelineTriggerTableRecord{}

			if v, match := result.Record().ValueByKey(constant.PipelineID).(string); match {
				tableRecord.PipelineId = v
			}
			if v, match := result.Record().ValueByKey(constant.PipelineUID).(string); match {
				tableRecord.PipelineUid = v
			}
			if v, match := result.Record().ValueByKey(constant.PipelineReleaseID).(string); match {
				tableRecord.PipelineReleaseId = v
			}
			if v, match := result.Record().ValueByKey(constant.PipelineReleaseUID).(string); match {
				tableRecord.PipelineReleaseUid = v
			}
			if v, match := result.Record().ValueByKey(mgmtPB.Status_STATUS_COMPLETED.String()).(int64); match {
				tableRecord.TriggerCountCompleted = int32(v)
			}
			if v, match := result.Record().ValueByKey(mgmtPB.Status_STATUS_ERRORED.String()).(int64); match {
				tableRecord.TriggerCountErrored = int32(v)
			}

			records = append(records, tableRecord)
		}

		// Check for an error
		if result.Err() != nil {
			return nil, 0, "", status.Errorf(codes.InvalidArgument, "Invalid query: %s", err.Error())
		}
		if result.Record() == nil {
			return nil, 0, "", nil
		}

		if v, match := result.Record().ValueByKey("most_recent_trigger_time").(time.Time); match {
			lastTimestamp = v
		}
	}

	var total int64
	totalQueryResult, err := i.queryAPI.Query(ctx, totalQuery)
	if err != nil {
		return nil, 0, "", status.Errorf(codes.InvalidArgument, "Invalid total query: %s", err.Error())
	} else {
		if totalQueryResult.Next() {
			total = totalQueryResult.Record().ValueByKey(constant.PipelineUID).(int64)
		}
	}

	if int64(len(records)) < total {
		pageToken = paginate.EncodeToken(lastTimestamp, owner)
	} else {
		pageToken = ""
	}

	return records, int64(len(records)), pageToken, nil
}

func (i *influxDB) QueryPipelineTriggerChartRecords(ctx context.Context, owner string, ownerQueryString string, aggregationWindow int64, filter filtering.Filter) (records []*mgmtPB.PipelineTriggerChartRecord, err error) {

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
		for i, expr := range exprs {
			if strings.HasPrefix(expr, constant.Start) {
				start = strings.Split(expr, "@")[1]
				exprs[i] = ""
			}
			if strings.HasPrefix(expr, constant.Stop) {
				stop = strings.Split(expr, "@")[1]
				exprs[i] = ""
			}
		}
		expr = strings.Join(exprs, "")
	}

	query := fmt.Sprintf(
		`base =
			from(bucket: "%v")
				|> range(start: %v, stop: %v)
				|> filter(fn: (r) => r["_measurement"] == "pipeline.trigger")
				|> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")
				%v
				%v
		bucketBase =
			base
				|> group(columns: ["pipeline_uid"])
				|> sort(columns: ["trigger_time"])
		bucketTrigger =
			bucketBase
				|> aggregateWindow(
					every: duration(v: %v),
					column: "trigger_time",
					fn: count,
					createEmpty: false,
				)
		bucketDuration =
			bucketBase
				|> aggregateWindow(
					every: duration(v: %v),
					fn: sum,
					column: "compute_time_duration",
					createEmpty: false,
				)
		bucket =
			join(
				tables: {t1: bucketTrigger, t2: bucketDuration},
				on: ["_start", "_stop", "_time", "pipeline_uid"],
			)
		nameMap =
			base
				|> keep(columns: ["trigger_time", "pipeline_id", "pipeline_uid"])
				|> group(columns: ["pipeline_uid"])
				|> top(columns: ["trigger_time"], n: 1)
				|> drop(columns: ["trigger_time"])
		join(tables: {t1: bucket, t2: nameMap}, on: ["pipeline_uid"])`,
		i.bucket,
		start,
		stop,
		ownerQueryString,
		expr,
		aggregationWindow,
		aggregationWindow,
	)

	result, err := i.queryAPI.Query(ctx, query)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid query: %s", err.Error())
	} else {

		var currentTablePosition = -1
		var chartRecord *mgmtPB.PipelineTriggerChartRecord

		// Iterate over query response
		for result.Next() {
			// Notice when group key has changed
			if result.TableChanged() {
				logger.Debug(fmt.Sprintf("table: %s\n", result.TableMetadata().String()))
			}

			if result.Record().Table() != currentTablePosition {

				chartRecord = &mgmtPB.PipelineTriggerChartRecord{}

				if v, match := result.Record().ValueByKey(constant.PipelineID).(string); match {
					chartRecord.PipelineId = v
				}
				if v, match := result.Record().ValueByKey(constant.PipelineUID).(string); match {
					chartRecord.PipelineUid = v
				}
				if v, match := result.Record().ValueByKey(constant.PipelineReleaseID).(string); match {
					chartRecord.PipelineReleaseId = v
				}
				if v, match := result.Record().ValueByKey(constant.PipelineReleaseUID).(string); match {
					chartRecord.PipelineReleaseUid = v
				}
				chartRecord.TimeBuckets = []*timestamppb.Timestamp{}
				chartRecord.TriggerCounts = []int32{}
				chartRecord.ComputeTimeDuration = []float32{}
				records = append(records, chartRecord)
				currentTablePosition = result.Record().Table()
			}

			if v, match := result.Record().ValueByKey("_time").(time.Time); match {
				chartRecord.TimeBuckets = append(chartRecord.TimeBuckets, timestamppb.New(v))
			}
			if v, match := result.Record().ValueByKey(constant.TriggerTime).(int64); match {
				chartRecord.TriggerCounts = append(chartRecord.TriggerCounts, int32(v))
			}
			if v, match := result.Record().ValueByKey(constant.ComputeTimeDuration).(float64); match {
				chartRecord.ComputeTimeDuration = append(chartRecord.ComputeTimeDuration, float32(v))
			}
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

func (i *influxDB) QueryConnectorExecuteRecords(ctx context.Context, owner string, ownerQueryString string, pageSize int64, pageToken string, filter filtering.Filter) (records []*mgmtPB.ConnectorExecuteRecord, totalSize int64, nextPageToken string, err error) {

	logger, _ := logger.GetZapLogger(ctx)

	query, total, err := i.constructRecordQuery(ctx, ownerQueryString, pageSize, pageToken, filter, constant.ConnectorExecuteMeasurement, "execute_time")
	if err != nil {
		return nil, 0, "", status.Errorf(codes.InvalidArgument, "Invalid query: %s", err.Error())
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

			record := &mgmtPB.ConnectorExecuteRecord{}

			if v, match := result.Record().ValueByKey(constant.ExecuteTime).(string); match {
				executeTime, err := time.Parse(time.RFC3339Nano, v)
				if err != nil {
					return nil, 0, "", status.Errorf(codes.InvalidArgument, "Invalid parse key: %s", err.Error())
				}
				record.ExecuteTime = timestamppb.New(executeTime)
			}
			if v, match := result.Record().ValueByKey(constant.PipelineTriggerID).(string); match {
				record.PipelineTriggerId = v
			}
			if v, match := result.Record().ValueByKey(constant.PipelineID).(string); match {
				record.PipelineId = v
			}
			if v, match := result.Record().ValueByKey(constant.PipelineUID).(string); match {
				record.PipelineUid = v
			}
			if v, match := result.Record().ValueByKey(constant.ConnectorExecuteID).(string); match {
				record.ConnectorExecuteId = v
			}
			if v, match := result.Record().ValueByKey(constant.ConnectorID).(string); match {
				record.ConnectorId = v
			}
			if v, match := result.Record().ValueByKey(constant.ConnectorUID).(string); match {
				record.ConnectorUid = v
			}
			if v, match := result.Record().ValueByKey(constant.ConnectorDefinitionUID).(string); match {
				record.ConnectorDefinitionUid = v
			}
			if v, match := result.Record().ValueByKey(constant.ComputeTimeDuration).(float64); match {
				record.ComputeTimeDuration = float32(v)
			}
			if v, match := result.Record().ValueByKey(constant.Status).(string); match {
				record.Status = mgmtPB.Status(mgmtPB.Status_value[v])
			}

			records = append(records, record)
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

func (i *influxDB) QueryConnectorExecuteTableRecords(ctx context.Context, owner string, ownerQueryString string, pageSize int64, pageToken string, filter filtering.Filter) (records []*mgmtPB.ConnectorExecuteTableRecord, totalSize int64, nextPageToken string, err error) {

	logger, _ := logger.GetZapLogger(ctx)

	if pageSize == 0 {
		pageSize = DefaultPageSize
	} else if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	start := time.Time{}.Format(time.RFC3339Nano)
	stop := time.Now().Format(time.RFC3339Nano)
	mostRecetTimeFilter := time.Now().Format(time.RFC3339Nano)

	// TODO: validate owner uid from token
	if pageToken != "" {
		mostRecetTime, _, err := paginate.DecodeToken(pageToken)
		if err != nil {
			return nil, 0, "", status.Errorf(codes.InvalidArgument, "Invalid page token: %s", err.Error())
		}
		mostRecetTime = mostRecetTime.Add(time.Duration(-1))
		mostRecetTimeFilter = mostRecetTime.Format(time.RFC3339Nano)
	}

	// TODO: design better filter expression to flux transpiler
	expr, err := i.transpileFilter(filter)
	if err != nil {
		return nil, 0, "", status.Errorf(codes.Internal, err.Error())
	}

	if expr != "" {
		exprs := strings.Split(expr, "&&")
		for i, expr := range exprs {
			if strings.HasPrefix(expr, constant.Start) {
				start = strings.Split(expr, "@")[1]
				exprs[i] = ""
			}
			if strings.HasPrefix(expr, constant.Stop) {
				stop = strings.Split(expr, "@")[1]
				exprs[i] = ""
			}
		}
		expr = strings.Join(exprs, "")
	}

	baseQuery := fmt.Sprintf(
		`base =
			from(bucket: "%v")
				|> range(start: %v, stop: %v)
				|> filter(fn: (r) => r["_measurement"] == "connector.execute")
				|> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")
				%v
				%v
		executeRank =
			base
				|> drop(
					columns: [
						"connector_owner_uid",
						"compute_time_duration",
						"pipeline_id",
						"pipeline_uid",
						"status",
					],
				)
				|> group(columns: ["connector_uid"])
				|> map(fn: (r) => ({r with execute_time: time(v: r.execute_time)}))
				|> max(column: "execute_time")
				|> rename(columns: {execute_time: "most_recent_execute_time"})
		executeCount =
			base
				|> drop(
					columns: ["connector_owner_uid", "compute_time_duration", "pipeline_id", "pipeline_uid"],
				)
				|> group(columns: ["connector_uid", "status"])
				|> count(column: "execute_time")
				|> rename(columns: {execute_time: "execute_count"})
				|> group(columns: ["connector_uid"])
		executeTable =
			join(tables: {t1: executeRank, t2: executeCount}, on: ["connector_uid"])
				|> group()
				|> pivot(
					rowKey: ["connector_uid", "most_recent_execute_time"],
					columnKey: ["status"],
					valueColumn: "execute_count",
				)
				|> filter(
					fn: (r) => r["most_recent_execute_time"] < time(v: %v)
				)
		nameMap =
			base
				|> keep(columns: ["execute_time", "connector_id", "connector_uid"])
				|> group(columns: ["connector_uid"])
				|> top(columns: ["execute_time"], n: 1)
				|> drop(columns: ["execute_time"])
				|> group()
		join(tables: {t1: executeTable, t2: nameMap}, on: ["connector_uid"])`,
		i.bucket,
		start,
		stop,
		ownerQueryString,
		expr,
		mostRecetTimeFilter,
	)

	query := fmt.Sprintf(
		`%v
		|> group()
		|> sort(columns: ["most_recent_execute_time"], desc: true)
		|> limit(n: %v)`,
		baseQuery,
		pageSize,
	)

	totalQuery := fmt.Sprintf(
		`%v
		|> group()
		|> count(column: "connector_uid")`,
		baseQuery,
	)

	var lastTimestamp time.Time

	result, err := i.queryAPI.Query(ctx, query)
	if err != nil {
		return nil, 0, "", status.Errorf(codes.InvalidArgument, "Invalid query: %s", err.Error())
	} else {
		// Iterate over query response
		for result.Next() {
			// Notice when group key has changed
			if result.TableChanged() {
				logger.Debug(fmt.Sprintf("table: %s\n", result.TableMetadata().String()))
			}

			tableRecord := &mgmtPB.ConnectorExecuteTableRecord{}

			if v, match := result.Record().ValueByKey(constant.ConnectorID).(string); match {
				tableRecord.ConnectorId = v
			}
			if v, match := result.Record().ValueByKey(constant.ConnectorUID).(string); match {
				tableRecord.ConnectorUid = v
			}
			if v, match := result.Record().ValueByKey(mgmtPB.Status_STATUS_COMPLETED.String()).(int64); match {
				tableRecord.ExecuteCountCompleted = int32(v)
			}
			if v, match := result.Record().ValueByKey(mgmtPB.Status_STATUS_ERRORED.String()).(int64); match {
				tableRecord.ExecuteCountErrored = int32(v)
			}

			records = append(records, tableRecord)
		}

		// Check for an error
		if result.Err() != nil {
			return nil, 0, "", status.Errorf(codes.InvalidArgument, "Invalid query: %s", err.Error())
		}
		if result.Record() == nil {
			return nil, 0, "", nil
		}

		if v, match := result.Record().ValueByKey("most_recent_execute_time").(time.Time); match {
			lastTimestamp = v
		}
	}

	var total int64
	totalQueryResult, err := i.queryAPI.Query(ctx, totalQuery)
	if err != nil {
		return nil, 0, "", status.Errorf(codes.InvalidArgument, "Invalid total query: %s", err.Error())
	} else {
		if totalQueryResult.Next() {
			total = totalQueryResult.Record().ValueByKey(constant.ConnectorUID).(int64)
		}
	}

	if int64(len(records)) < total {
		pageToken = paginate.EncodeToken(lastTimestamp, owner)
	} else {
		pageToken = ""
	}

	return records, int64(len(records)), pageToken, nil
}

func (i *influxDB) QueryConnectorExecuteChartRecords(ctx context.Context, owner string, ownerQueryString string, aggregationWindow int64, filter filtering.Filter) (records []*mgmtPB.ConnectorExecuteChartRecord, err error) {

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
		for i, expr := range exprs {
			if strings.HasPrefix(expr, constant.Start) {
				start = strings.Split(expr, "@")[1]
				exprs[i] = ""
			}
			if strings.HasPrefix(expr, constant.Stop) {
				stop = strings.Split(expr, "@")[1]
				exprs[i] = ""
			}
		}
		expr = strings.Join(exprs, "")
	}

	query := fmt.Sprintf(
		`base =
			from(bucket: "%v")
				|> range(start: %v, stop: %v)
				|> filter(fn: (r) => r["_measurement"] == "connector.execute")
				|> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")
				%v
				%v
		bucketBase =
			base
				|> group(columns: ["connector_uid"])
				|> sort(columns: ["execute_time"])
		bucketTrigger =
			bucketBase
				|> aggregateWindow(
					every: duration(v: %v),
					column: "execute_time",
					fn: count,
					createEmpty: false,
				)
		bucketDuration =
			bucketBase
				|> aggregateWindow(
					every: duration(v: %v),
					fn: sum,
					column: "compute_time_duration",
					createEmpty: false,
				)
		bucket =
			join(
				tables: {t1: bucketTrigger, t2: bucketDuration},
				on: ["_start", "_stop", "_time", "connector_uid"],
			)
		nameMap =
			base
				|> keep(columns: ["execute_time", "connector_id", "connector_uid"])
				|> group(columns: ["connector_uid"])
				|> top(columns: ["execute_time"], n: 1)
				|> drop(columns: ["execute_time"])
		join(tables: {t1: bucket, t2: nameMap}, on: ["connector_uid"])`,
		i.bucket,
		start,
		stop,
		ownerQueryString,
		expr,
		aggregationWindow,
		aggregationWindow,
	)

	result, err := i.queryAPI.Query(ctx, query)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid query: %s", err.Error())
	} else {

		var currentTablePosition = -1
		var chartRecord *mgmtPB.ConnectorExecuteChartRecord

		// Iterate over query response
		for result.Next() {
			// Notice when group key has changed
			if result.TableChanged() {
				logger.Debug(fmt.Sprintf("table: %s\n", result.TableMetadata().String()))
			}

			if result.Record().Table() != currentTablePosition {

				chartRecord = &mgmtPB.ConnectorExecuteChartRecord{}

				if v, match := result.Record().ValueByKey(constant.ConnectorID).(string); match {
					chartRecord.ConnectorId = v
				}
				if v, match := result.Record().ValueByKey(constant.ConnectorUID).(string); match {
					chartRecord.ConnectorUid = v
				}
				chartRecord.TimeBuckets = []*timestamppb.Timestamp{}
				chartRecord.ExecuteCounts = []int32{}
				chartRecord.ComputeTimeDuration = []float32{}
				records = append(records, chartRecord)
				currentTablePosition = result.Record().Table()
			}

			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "Invalid parse key: %s", err.Error())
			}
			if v, match := result.Record().ValueByKey("_time").(time.Time); match {
				chartRecord.TimeBuckets = append(chartRecord.TimeBuckets, timestamppb.New(v))
			}
			if v, match := result.Record().ValueByKey(constant.ExecuteTime).(int64); match {
				chartRecord.ExecuteCounts = append(chartRecord.ExecuteCounts, int32(v))
			}
			if v, match := result.Record().ValueByKey(constant.ComputeTimeDuration).(float64); match {
				chartRecord.ComputeTimeDuration = append(chartRecord.ComputeTimeDuration, float32(v))
			}
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
