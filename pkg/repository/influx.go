package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/log"
	"go.einride.tech/aip/filtering"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	client "github.com/influxdata/influxdb-client-go/v2"

	"github.com/instill-ai/mgmt-backend/config"
	"github.com/instill-ai/mgmt-backend/pkg/constant"
	"github.com/instill-ai/x/paginate"

	mgmtpb "github.com/instill-ai/protogen-go/core/mgmt/v1beta"
	errorsx "github.com/instill-ai/x/errors"
	logx "github.com/instill-ai/x/log"
)

// Default aggregate window
var defaultAggregationWindow = time.Hour.Nanoseconds()

// InfluxDB interface
type InfluxDB interface {
	QueryPipelineTriggerRecords(ctx context.Context, owner string, ownerQueryString string, pageSize int64, pageToken string, filter filtering.Filter) (pipelines []*mgmtpb.PipelineTriggerRecord, totalSize int64, nextPageToken string, err error)
	QueryPipelineTriggerTableRecords(ctx context.Context, owner string, ownerQueryString string, pageSize int64, pageToken string, filter filtering.Filter) (records []*mgmtpb.PipelineTriggerTableRecord, totalSize int64, nextPageToken string, err error)
	QueryPipelineTriggerChartRecordsV0(ctx context.Context, owner string, ownerQueryString string, aggregationWindow int64, filter filtering.Filter) (records []*mgmtpb.PipelineTriggerChartRecordV0, err error)

	ListPipelineTriggerChartRecords(context.Context, ListTriggerChartRecordsParams) (*mgmtpb.ListPipelineTriggerChartRecordsResponse, error)
	GetPipelineTriggerCount(context.Context, GetTriggerCountParams) (*mgmtpb.GetPipelineTriggerCountResponse, error)

	ListModelTriggerChartRecords(context.Context, ListTriggerChartRecordsParams) (*mgmtpb.ListModelTriggerChartRecordsResponse, error)
	GetModelTriggerCount(context.Context, GetTriggerCountParams) (*mgmtpb.GetModelTriggerCountResponse, error)

	Bucket() string
	QueryAPI() api.QueryAPI
	Close()
}

type influxDB struct {
	client client.Client
	api    api.QueryAPI
	bucket string
}

// MustNewInfluxDB returns an initialized InfluxDB repository.
func MustNewInfluxDB(ctx context.Context, cfg config.AppConfig) InfluxDB {
	logger, _ := logx.GetZapLogger(ctx)

	opts := client.DefaultOptions()
	if cfg.Server.Debug {
		opts = opts.SetLogLevel(log.DebugLevel)
	}

	flush := uint(cfg.InfluxDB.FlushInterval.Milliseconds())
	opts = opts.SetFlushInterval(flush)

	var creds credentials.TransportCredentials
	var err error

	if cfg.InfluxDB.HTTPS.Cert != "" && cfg.InfluxDB.HTTPS.Key != "" {
		// TODO support TLS
		creds, err = credentials.NewServerTLSFromFile(cfg.InfluxDB.HTTPS.Cert, cfg.InfluxDB.HTTPS.Key)
		if err != nil {
			logger.With(zap.Error(err)).Fatal("Couldn't initialize InfluxDB client")
		}

		logger = logger.With(zap.String("influxServer", creds.Info().ServerName))
	}

	i := new(influxDB)
	i.client = client.NewClientWithOptions(
		cfg.InfluxDB.URL,
		cfg.InfluxDB.Token,
		opts,
	)
	i.bucket = cfg.InfluxDB.Bucket

	org := cfg.InfluxDB.Org
	i.api = i.client.QueryAPI(org)
	logger = logger.With(zap.String("bucket", i.bucket)).
		With(zap.String("org", org))

	logger.Info("InfluxDB client initialized")
	if _, err := i.client.Ping(ctx); err != nil {
		logger.With(zap.Error(err)).Warn("Failed to ping InfluxDB")
	}

	return i
}

// Close  cleans up the InfluxDB connections.
func (i *influxDB) Close() {
	i.client.Close()
}

// QueryAPI return the InfluxDB client's Query API.
// TODO this is a shortcut to avoid refactoring client packages (e.g. worker)
// but we should use a TimeSeriesRepository interface in them.
func (i *influxDB) QueryAPI() api.QueryAPI {
	return i.api
}

// Bucket returns the InfluxDB bucket the repository reads from.
func (i *influxDB) Bucket() string {
	return i.bucket
}

func (i *influxDB) QueryPipelineTriggerTableRecords(ctx context.Context, owner string, ownerQueryString string, pageSize int64, pageToken string, filter filtering.Filter) (records []*mgmtpb.PipelineTriggerTableRecord, totalSize int64, nextPageToken string, err error) {

	logger, _ := logx.GetZapLogger(ctx)

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
		return nil, 0, "", status.Error(codes.Internal, err.Error())
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
				|> sort(columns: ["trigger_time"], desc: true)
				|> first(column: "trigger_time")
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

	result, err := i.api.Query(ctx, query)
	if err != nil {
		return nil, 0, "", status.Errorf(codes.InvalidArgument, "Invalid query: %s", err.Error())
	} else {
		// Iterate over query response
		for result.Next() {
			// Notice when group key has changed
			if result.TableChanged() {
				logger.Debug(fmt.Sprintf("table: %s\n", result.TableMetadata().String()))
			}

			tableRecord := &mgmtpb.PipelineTriggerTableRecord{}

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
			if v, match := result.Record().ValueByKey(mgmtpb.Status_STATUS_COMPLETED.String()).(int64); match {
				tableRecord.TriggerCountCompleted = int32(v)
			}
			if v, match := result.Record().ValueByKey(mgmtpb.Status_STATUS_ERRORED.String()).(int64); match {
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
	totalQueryResult, err := i.api.Query(ctx, totalQuery)
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

func (i *influxDB) QueryPipelineTriggerChartRecordsV0(ctx context.Context, owner string, ownerQueryString string, aggregationWindow int64, filter filtering.Filter) (records []*mgmtpb.PipelineTriggerChartRecordV0, err error) {

	logger, _ := logx.GetZapLogger(ctx)

	start := time.Time{}.Format(time.RFC3339Nano)
	stop := time.Now().Format(time.RFC3339Nano)

	if aggregationWindow < time.Minute.Nanoseconds() {
		aggregationWindow = defaultAggregationWindow
	}

	// TODO: design better filter expression to flux transpiler
	expr, err := i.transpileFilter(filter)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
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

	result, err := i.api.Query(ctx, query)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid query: %s", err.Error())
	}

	var currentTablePosition = -1
	var chartRecord *mgmtpb.PipelineTriggerChartRecordV0

	// Iterate over query response
	for result.Next() {
		// Notice when group key has changed
		if result.TableChanged() {
			logger.Debug(fmt.Sprintf("table: %s\n", result.TableMetadata().String()))
		}

		if result.Record().Table() != currentTablePosition { // only insert a new object when iterated to a new pipeline
			chartRecord = &mgmtpb.PipelineTriggerChartRecordV0{}

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

	return records, nil
}

const (
	pipelineTriggerMeasurement = "pipeline.trigger.v1"
	modelTriggerMeasurement    = "model.trigger.v1"
)

const qTriggerCount = `
from(bucket: "%s")
	|> range(start: %s, stop: %s)
	|> filter(fn: (r) => r._measurement == "%s" and r.requester_uid == "%s")
	|> filter(fn: (r) => r._field == "trigger_time")
	|> group(columns: ["requester_uid", "status"])
	|> count(column: "_value")
`

// GetTriggerCountParams contains the required information to query the
// pipeline or model trigger count of a namespace.
// TODO jvallesm: this should be defined in the service package for better
// decoupling. At the moment this implies breaking an import cycle with many
// dependencies.
type GetTriggerCountParams struct {
	RequesterUID uuid.UUID
	Start        time.Time
	Stop         time.Time
}

func (i *influxDB) getTriggerCount(
	ctx context.Context,
	p GetTriggerCountParams,
	measurement string,
) ([]*mgmtpb.TriggerCount, error) {
	l, _ := logx.GetZapLogger(ctx)
	l = l.With(
		zap.Reflect("triggerCountParams", p),
		zap.String("measurement", measurement),
	)

	query := fmt.Sprintf(
		qTriggerCount,
		i.Bucket(),
		p.Start.Format(time.RFC3339Nano),
		p.Stop.Format(time.RFC3339Nano),
		measurement,
		p.RequesterUID.String(),
	)
	result, err := i.QueryAPI().Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%w: querying data from InfluxDB: %w", errorsx.ErrInvalidArgument, err)
	}

	defer result.Close()

	// We'll have one record per status.
	countRecords := make([]*mgmtpb.TriggerCount, 0, 2)
	for result.Next() {
		l := l.With(zap.Time("_time", result.Record().Time()))

		statusStr := result.Record().ValueByKey("status").(string)
		status := mgmtpb.Status(mgmtpb.Status_value[statusStr])
		if status == mgmtpb.Status_STATUS_UNSPECIFIED {
			l.Error("Missing status on trigger count record.")
		}

		count, match := result.Record().Value().(int64)
		if !match {
			l.Error("Missing count on trigger count record.")
		}

		countRecords = append(countRecords, &mgmtpb.TriggerCount{
			TriggerCount: int32(count),
			Status:       &status,
		})
	}

	if result.Err() != nil {
		return nil, fmt.Errorf("collecting information from trigger count records: %w", err)
	}

	if result.Record() == nil {
		return nil, nil
	}

	return countRecords, nil
}

func (i *influxDB) GetPipelineTriggerCount(ctx context.Context, p GetTriggerCountParams) (*mgmtpb.GetPipelineTriggerCountResponse, error) {
	countRecords, err := i.getTriggerCount(ctx, p, pipelineTriggerMeasurement)
	if err != nil {
		return nil, err
	}

	return &mgmtpb.GetPipelineTriggerCountResponse{
		PipelineTriggerCounts: countRecords,
	}, nil
}

func (i *influxDB) GetModelTriggerCount(ctx context.Context, p GetTriggerCountParams) (*mgmtpb.GetModelTriggerCountResponse, error) {
	countRecords, err := i.getTriggerCount(ctx, p, modelTriggerMeasurement)
	if err != nil {
		return nil, err
	}

	return &mgmtpb.GetModelTriggerCountResponse{
		ModelTriggerCounts: countRecords,
	}, nil
}

const qTriggerChartRecords = `
from(bucket: "%s")
	|> range(start: %s, stop: %s)
	|> filter(fn: (r) => r._measurement == "%s" and r.requester_uid == "%s")
	|> filter(fn: (r) => r._field == "trigger_time")
	|> group(columns:["requester_uid"])
	|> aggregateWindow(every: %s, column:"_value", fn: count, createEmpty: true, offset: %s)
`

func (i *influxDB) listTriggerChartRecords(
	ctx context.Context,
	p ListTriggerChartRecordsParams,
	measurement string,
) (*api.QueryTableResult, error) {
	query := fmt.Sprintf(
		qTriggerChartRecords,
		i.Bucket(),
		p.Start.Format(time.RFC3339Nano),
		p.Stop.Format(time.RFC3339Nano),
		measurement,
		p.RequesterUID.String(),
		p.AggregationWindow,
		AggregationWindowOffset(p.Start).String(),
	)
	result, err := i.QueryAPI().Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%w: querying data from InfluxDB: %w", errorsx.ErrInvalidArgument, err)
	}

	return result, nil
}

// ListTriggerChartRecordsParams contains the required information to query the
// triggers of a requester.
// TODO jvallesm: this should be defined in the service package for better
// decoupling. At the moment this implies breaking an import cycle with many
// dependencies.
type ListTriggerChartRecordsParams struct {
	RequesterID       string
	RequesterUID      uuid.UUID
	AggregationWindow time.Duration
	Start             time.Time
	Stop              time.Time
}

func (i *influxDB) ListPipelineTriggerChartRecords(
	ctx context.Context,
	p ListTriggerChartRecordsParams,
) (*mgmtpb.ListPipelineTriggerChartRecordsResponse, error) {
	l, _ := logx.GetZapLogger(ctx)
	l = l.With(
		zap.Reflect("triggerChartParams", p),
		zap.String("measurement", pipelineTriggerMeasurement),
	)

	result, err := i.listTriggerChartRecords(ctx, p, pipelineTriggerMeasurement)
	if err != nil {
		return nil, err
	}

	defer result.Close()

	record := &mgmtpb.PipelineTriggerChartRecord{
		RequesterId:   p.RequesterID,
		TimeBuckets:   []*timestamppb.Timestamp{},
		TriggerCounts: []int32{},
	}

	// Until filtering and grouping are implemented, we'll only have one record
	// (total triggers by requester).
	records := []*mgmtpb.PipelineTriggerChartRecord{record}

	for result.Next() {
		t := result.Record().Time()
		record.TimeBuckets = append(record.TimeBuckets, timestamppb.New(t))

		v, match := result.Record().Value().(int64)
		if !match {
			l.With(zap.Time("_time", result.Record().Time())).
				Error("Missing count on trigger chart record.")
		}

		record.TriggerCounts = append(record.TriggerCounts, int32(v))
	}

	if result.Err() != nil {
		return nil, fmt.Errorf("collecting information from trigger chart records: %w", err)
	}

	if result.Record() == nil {
		return nil, nil
	}

	return &mgmtpb.ListPipelineTriggerChartRecordsResponse{
		PipelineTriggerChartRecords: records,
	}, nil
}

func (i *influxDB) ListModelTriggerChartRecords(
	ctx context.Context,
	p ListTriggerChartRecordsParams,
) (*mgmtpb.ListModelTriggerChartRecordsResponse, error) {
	l, _ := logx.GetZapLogger(ctx)
	l = l.With(
		zap.Reflect("triggerChartParams", p),
		zap.String("measurement", modelTriggerMeasurement),
	)

	result, err := i.listTriggerChartRecords(ctx, p, modelTriggerMeasurement)
	if err != nil {
		return nil, err
	}

	defer result.Close()

	record := &mgmtpb.ModelTriggerChartRecord{
		RequesterId:   p.RequesterID,
		TimeBuckets:   []*timestamppb.Timestamp{},
		TriggerCounts: []int32{},
	}

	// Until filtering and grouping are implemented, we'll only have one record
	// (total triggers by requester).
	records := []*mgmtpb.ModelTriggerChartRecord{record}

	for result.Next() {
		t := result.Record().Time()
		record.TimeBuckets = append(record.TimeBuckets, timestamppb.New(t))

		v, match := result.Record().Value().(int64)
		if !match {
			l.With(zap.Time("_time", result.Record().Time())).
				Error("Missing count on model trigger chart record.")
		}

		record.TriggerCounts = append(record.TriggerCounts, int32(v))
	}

	if result.Err() != nil {
		return nil, fmt.Errorf("collecting information from model trigger chart records: %w", err)
	}

	if result.Record() == nil {
		return nil, nil
	}

	return &mgmtpb.ListModelTriggerChartRecordsResponse{
		ModelTriggerChartRecords: records,
	}, nil
}

// TranspileFilter transpiles a parsed AIP filter expression to Flux query expression
func (i *influxDB) transpileFilter(filter filtering.Filter) (string, error) {
	return (&Transpiler{
		filter: filter,
	}).Transpile()
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
		return "", 0, status.Errorf(codes.Internal, "%s", err.Error())
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
	totalQueryResult, err := i.api.Query(ctx, totalQuery)
	if err != nil {
		return "", 0, status.Errorf(codes.InvalidArgument, "Invalid query: %s", err.Error())
	}

	if totalQueryResult.Next() {
		total = totalQueryResult.Record().ValueByKey(sortKey).(int64)
	}

	return query, total, nil
}

func (i *influxDB) QueryPipelineTriggerRecords(ctx context.Context, owner string, ownerQueryString string, pageSize int64, pageToken string, filter filtering.Filter) (records []*mgmtpb.PipelineTriggerRecord, totalSize int64, nextPageToken string, err error) {
	logger, _ := logx.GetZapLogger(ctx)
	query, total, err := i.constructRecordQuery(ctx, ownerQueryString, pageSize, pageToken, filter, constant.PipelineTriggerMeasurement, constant.TriggerTime)
	if err != nil {
		return nil, 0, "", status.Errorf(codes.InvalidArgument, "Invalid query: %s", err.Error())
	}
	result, err := i.api.Query(ctx, query)
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
			record := &mgmtpb.PipelineTriggerRecord{}
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
				record.TriggerMode = mgmtpb.Mode(mgmtpb.Mode_value[v])
			}
			if v, match := result.Record().ValueByKey(constant.ComputeTimeDuration).(float64); match {
				record.ComputeTimeDuration = float32(v)
			}
			// TODO: temporary solution for legacy data format, currently there is no way to update the tags in influxdb
			if v, match := result.Record().ValueByKey(constant.Status).(string); match {
				switch v {
				case constant.Completed:
					record.Status = mgmtpb.Status_STATUS_COMPLETED
				case constant.Errored:
					record.Status = mgmtpb.Status_STATUS_ERRORED
				default:
					record.Status = mgmtpb.Status(mgmtpb.Status_value[v])
				}
			}
			if v, match := result.Record().ValueByKey(constant.PipelineMode).(string); match {
				record.TriggerMode = mgmtpb.Mode(mgmtpb.Mode_value[v])
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

// AggregationWindowOffset computes the offset to apply to InfluxDB's
// aggregateWindow function when aggregating by day. This function computes
// windows independently, starting from the Unix epoch, rather than from the
// provided time range start. This function computes the offset to shift the
// windows correctly.
func AggregationWindowOffset(t time.Time) time.Duration {
	startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	return t.Sub(startOfDay)
}
