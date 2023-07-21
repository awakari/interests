package config

import (
	"github.com/kelseyhightower/envconfig"
	"golang.org/x/exp/slog"
)

type Config struct {
	Api struct {
		Port uint16 `envconfig:"API_PORT" default:"50051" required:"true"`
	}
	Db  DbConfig
	Log struct {
		Level slog.Level `envconfig:"LOG_LEVEL" default:"-4" required:"true"`
	}
}

type DbConfig struct {
	Uri      string `envconfig:"DB_URI" default:"mongodb://localhost:27017/?retryWrites=true&w=majority" required:"true"`
	Name     string `envconfig:"DB_NAME" default:"subscriptions" required:"true"`
	UserName string `envconfig:"DB_USERNAME" default:""`
	Password string `envconfig:"DB_PASSWORD" default:""`
	Table    struct {
		Name  string `envconfig:"DB_NAME" default:"subscriptions" required:"true"`
		Shard bool   `envconfig:"DB_TABLE_SHARD" default:"true"`
	}
	Tls struct {
		Enabled  bool `envconfig:"DB_TLS_ENABLED" default:"false" required:"true"`
		Insecure bool `envconfig:"DB_TLS_INSECURE" default:"false" required:"true"`
	}
}

func NewConfigFromEnv() (cfg Config, err error) {
	err = envconfig.Process("", &cfg)
	return
}
