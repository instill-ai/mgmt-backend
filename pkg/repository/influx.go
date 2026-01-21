package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/log"
	"go.uber.org/zap"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/timestamppb"

	client "github.com/influxdata/influxdb-client-go/v2"

	"github.com/instill-ai/mgmt-backend/config"

	mgmtpb "github.com/instill-ai/protogen-go/mgmt/v1beta"
	errorsx "github.com/instill-ai/x/errors"
	logx "github.com/instill-ai/x/log"
)

// InfluxDB interface
type InfluxDB interface {
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

// AggregationWindowOffset computes the offset to apply to InfluxDB's
// aggregateWindow function when aggregating by day. This function computes
// windows independently, starting from the Unix epoch, rather than from the
// provided time range start. This function computes the offset to shift the
// windows correctly.
func AggregationWindowOffset(t time.Time) time.Duration {
	startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	return t.Sub(startOfDay)
}
