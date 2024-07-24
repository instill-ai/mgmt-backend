package repository

import (
	"context"

	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/log"
	"go.uber.org/zap"
	"google.golang.org/grpc/credentials"

	client "github.com/influxdata/influxdb-client-go/v2"

	"github.com/instill-ai/mgmt-backend/config"
	"github.com/instill-ai/mgmt-backend/pkg/logger"
)

// DefaultPageSize is the default pagination page size when page size is not assigned
const DefaultPageSize = 100

// MaxPageSize is the maximum pagination page size if the assigned value is over this number
const MaxPageSize = 1000

// InfluxDB interface
type InfluxDB interface {
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
