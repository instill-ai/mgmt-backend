package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/log"
	"go.einride.tech/aip/filtering"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	client "github.com/influxdata/influxdb-client-go/v2"

	"github.com/instill-ai/mgmt-backend/config"
	"github.com/instill-ai/mgmt-backend/pkg/constant"
	"github.com/instill-ai/mgmt-backend/pkg/logger"
	"github.com/instill-ai/x/paginate"

	mgmtpb "github.com/instill-ai/protogen-go/core/mgmt/v1beta"
)

// DefaultPageSize is the default pagination page size when page size is not assigned
const DefaultPageSize = 100

// MaxPageSize is the maximum pagination page size if the assigned value is over this number
const MaxPageSize = 1000

// InfluxDB interface
type InfluxDB interface {
	QueryPipelineTriggerTableRecords(ctx context.Context, owner string, ownerQueryString string, pageSize int64, pageToken string, filter filtering.Filter) (records []*mgmtpb.PipelineTriggerTableRecord, totalSize int64, nextPageToken string, err error)

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
	logger, _ := logger.GetZapLogger(ctx)

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

// TranspileFilter transpiles a parsed AIP filter expression to Flux query expression
func (i *influxDB) transpileFilter(filter filtering.Filter) (string, error) {
	return (&Transpiler{
		filter: filter,
	}).Transpile()
}
