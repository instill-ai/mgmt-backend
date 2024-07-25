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
	"github.com/instill-ai/mgmt-backend/pkg/logger"

	errdomain "github.com/instill-ai/mgmt-backend/pkg/errors"
	mgmtpb "github.com/instill-ai/protogen-go/core/mgmt/v1beta"
)

// InfluxDB interface
type InfluxDB interface {
	ListPipelineTriggerChartRecords(ctx context.Context, p ListPipelineTriggerChartRecordsParams) (*mgmtpb.ListPipelineTriggerChartRecordsResponse, error)
	GetPipelineTriggerCount(ctx context.Context, p GetPipelineTriggerCountParams) (*mgmtpb.GetPipelineTriggerCountResponse, error)

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

const qPipelineTriggerChartRecords = `
from(bucket: "%s")
	|> range(start: %s, stop: %s)
	|> filter(fn: (r) => r._measurement == "pipeline.trigger" and r.requester_uid == "%s")
	|> filter(fn: (r) => r._field == "trigger_time")
	|> group(columns:["requester_uid"])
	|> aggregateWindow(every: %s, column:"_value", fn: count, createEmpty: true, offset: %s)
`

// ListPipelineTriggerChartRecordsParams contains the required information to
// query the pipeline triggers of a namespace.
// TODO jvallesm: this should be defined in the service package for better
// decoupling. At the moment this implies breaking an import cycle with many
// dependencies.
type ListPipelineTriggerChartRecordsParams struct {
	NamespaceID       string
	NamespaceUID      uuid.UUID
	AggregationWindow time.Duration
	Start             time.Time
	Stop              time.Time
}

func (i *influxDB) ListPipelineTriggerChartRecords(
	ctx context.Context,
	p ListPipelineTriggerChartRecordsParams,
) (*mgmtpb.ListPipelineTriggerChartRecordsResponse, error) {
	l, _ := logger.GetZapLogger(ctx)
	l = l.With(zap.Reflect("triggerChartParams", p))

	query := fmt.Sprintf(
		qPipelineTriggerChartRecords,
		i.Bucket(),
		p.Start.Format(time.RFC3339Nano),
		p.Stop.Format(time.RFC3339Nano),
		p.NamespaceUID.String(),
		p.AggregationWindow,
		AggregationWindowOffset(p.Start).String(),
	)
	result, err := i.QueryAPI().Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%w: querying data from InfluxDB: %w", errdomain.ErrInvalidArgument, err)
	}

	defer result.Close()

	record := &mgmtpb.PipelineTriggerChartRecord{
		NamespaceId:   p.NamespaceID,
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
				Error("Missing count on pipeline trigger chart record.")
		}

		record.TriggerCounts = append(record.TriggerCounts, int32(v))
	}

	if result.Err() != nil {
		return nil, fmt.Errorf("collecting information from pipeline trigger chart records: %w", err)
	}

	if result.Record() == nil {
		return nil, nil
	}

	return &mgmtpb.ListPipelineTriggerChartRecordsResponse{
		PipelineTriggerChartRecords: records,
	}, nil
}

const qPipelineTriggerCount = `
from(bucket: "%s")
	|> range(start: %s, stop: %s)
	|> filter(fn: (r) => r._measurement == "pipeline.trigger" and r.requester_uid == "%s")
	|> filter(fn: (r) => r._field == "trigger_time")
	|> group(columns: ["requester_uid", "status"])
	|> count(column: "_value")
`

// GetPipelineTriggerCountParams contains the required information to
// query the pipeline trigger count of a namespace.
// TODO jvallesm: this should be defined in the service package for better
// decoupling. At the moment this implies breaking an import cycle with many
// dependencies.
type GetPipelineTriggerCountParams struct {
	NamespaceUID uuid.UUID
	Start        time.Time
	Stop         time.Time
}

func (i *influxDB) GetPipelineTriggerCount(
	ctx context.Context,
	p GetPipelineTriggerCountParams,
) (*mgmtpb.GetPipelineTriggerCountResponse, error) {
	l, _ := logger.GetZapLogger(ctx)
	l = l.With(zap.Reflect("triggerCountParams", p))

	query := fmt.Sprintf(
		qPipelineTriggerCount,
		i.Bucket(),
		p.Start.Format(time.RFC3339Nano),
		p.Stop.Format(time.RFC3339Nano),
		p.NamespaceUID.String(),
	)
	result, err := i.QueryAPI().Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%w: querying data from InfluxDB: %w", errdomain.ErrInvalidArgument, err)
	}

	defer result.Close()

	// We'll have one record per status.
	countRecords := make([]*mgmtpb.PipelineTriggerCount, 0, 2)
	for result.Next() {
		l := l.With(zap.Time("_time", result.Record().Time()))

		statusStr := result.Record().ValueByKey("status").(string)
		status := mgmtpb.Status(mgmtpb.Status_value[statusStr])
		if status == mgmtpb.Status_STATUS_UNSPECIFIED {
			l.Error("Missing status on trigger count record.")
		}

		count, match := result.Record().Value().(int64)
		if !match {
			l.Error("Missing count on pipeline trigger count record.")
		}

		countRecords = append(countRecords, &mgmtpb.PipelineTriggerCount{
			TriggerCount: int32(count),
			Status:       &status,
		})
	}

	if result.Err() != nil {
		return nil, fmt.Errorf("collecting information from pipeline trigger count records: %w", err)
	}

	if result.Record() == nil {
		return nil, nil
	}

	return &mgmtpb.GetPipelineTriggerCountResponse{
		PipelineTriggerCounts: countRecords,
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
