package config

import (
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/redis/go-redis/v9"
)

// Config - Global variable to export
var Config AppConfig

// AppConfig defines
type AppConfig struct {
	Server          ServerConfig          `koanf:"server"`
	Database        DatabaseConfig        `koanf:"database"`
	Cache           CacheConfig           `koanf:"cache"`
	Log             LogConfig             `koanf:"log"`
	InfluxDB        InfluxDBConfig        `koanf:"influxdb"`
	PipelineBackend PipelineBackendConfig `koanf:"pipelinebackend"`
	OpenFGA         OpenFGAConfig         `koanf:"openfga"`
	Temporal        TemporalConfig        `koanf:"temporal"`
}

// ServerConfig defines HTTP server configurations
type ServerConfig struct {
	PrivatePort int `koanf:"privateport"`
	PublicPort  int `koanf:"publicport"`
	HTTPS       struct {
		Cert string `koanf:"cert"`
		Key  string `koanf:"key"`
	}
	Edition string `koanf:"edition"`
	Usage   struct {
		Enabled    bool   `koanf:"enabled"`
		TLSEnabled bool   `koanf:"tlsenabled"`
		Host       string `koanf:"host"`
		Port       int    `koanf:"port"`
	}
	Debug          bool   `koanf:"debug"`
	DefaultUserUID string `koanf:"defaultuseruid"`
}

// TemporalConfig related to Temporal
type TemporalConfig struct {
	HostPort   string `koanf:"hostport"`
	Namespace  string `koanf:"namespace"`
	Retention  string `koanf:"retention"`
	Ca         string `koanf:"ca"`
	Cert       string `koanf:"cert"`
	Key        string `koanf:"key"`
	ServerName string `koanf:"servername"`
}

// OpenFGA config
type OpenFGAConfig struct {
	Host string `koanf:"host"`
	Port int    `koanf:"port"`
}

// CacheConfig related to Redis
type CacheConfig struct {
	Redis struct {
		RedisOptions redis.Options `koanf:"redisoptions"`
	}
}

// PipelineBackendConfig related to pipeline-backend
type PipelineBackendConfig struct {
	Host       string `koanf:"host"`
	PublicPort int    `koanf:"publicport"`
	HTTPS      struct {
		Cert string `koanf:"cert"`
		Key  string `koanf:"key"`
	}
}

// DatabaseConfig related to database
type DatabaseConfig struct {
	Username string `koanf:"username"`
	Password string `koanf:"password"`
	Host     string `koanf:"host"`
	Port     int    `koanf:"port"`
	Replica  struct {
		Username             string `koanf:"username"`
		Password             string `koanf:"password"`
		Host                 string `koanf:"host"`
		Port                 int    `koanf:"port"`
		ReplicationTimeFrame int    `koanf:"replicationtimeframe"` // in seconds
	} `koanf:"replica"`
	Name     string `koanf:"name"`
	Version  uint   `koanf:"version"`
	TimeZone string `koanf:"timezone"`
	Pool     struct {
		IdleConnections int           `koanf:"idleconnections"`
		MaxConnections  int           `koanf:"maxconnections"`
		ConnLifeTime    time.Duration `koanf:"connlifetime"`
	}
}

// InfluxDBConfig related to influxDB database
type InfluxDBConfig struct {
	URL           string `koanf:"url"`
	Token         string `koanf:"token"`
	Org           string `koanf:"org"`
	Bucket        string `koanf:"bucket"`
	FlushInterval int    `koanf:"flushinterval"`
	HTTPS         struct {
		Cert string `koanf:"cert"`
		Key  string `koanf:"key"`
	}
}

// LogConfig related to logging
type LogConfig struct {
	External      bool `koanf:"external"`
	OtelCollector struct {
		Host string `koanf:"host"`
		Port string `koanf:"port"`
	}
}

// Init - Assign global config to decoded config struct
func Init() error {
	k := koanf.New(".")
	parser := yaml.Parser()

	if err := k.Load(confmap.Provider(map[string]interface{}{
		"database.replica.replicationtimeframe": 180,
	}, "."), nil); err != nil {
		log.Fatal(err.Error())
	}

	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fileRelativePath := fs.String("file", "config/config.yaml", "configuration file")
	flag.Parse()

	if err := k.Load(file.Provider(*fileRelativePath), parser); err != nil {
		log.Fatal(err.Error())
	}

	if err := k.Load(env.ProviderWithValue("CFG_", ".", func(s, v string) (string, interface{}) {
		key := strings.Replace(strings.ToLower(strings.TrimPrefix(s, "CFG_")), "_", ".", -1)
		if strings.Contains(v, ",") {
			return key, strings.Split(strings.TrimSpace(v), ",")
		}
		return key, v
	}), nil); err != nil {
		return err
	}

	if err := k.Unmarshal("", &Config); err != nil {
		return err
	}

	return ValidateConfig(&Config)
}

// ValidateConfig is for custom validation rules for the configuration
func ValidateConfig(cfg *AppConfig) error {
	return nil
}
