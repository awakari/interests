package config

import (
	"github.com/kelseyhightower/envconfig"
	"golang.org/x/exp/slog"
)

type (
	Config struct {
		Api struct {
			Port     uint16 `envconfig:"API_PORT" default:"50051" required:"true"`
			KiwiTree KiwiTree
		}
		Db  Db
		Log struct {
			Level slog.Level `envconfig:"LOG_LEVEL" default:"-4" required:"true"`
		}
	}

	KiwiTree struct {
		CompleteUri string `envconfig:"API_KIWI_TREE_COMPLETE_URI" default:"kiwi-tree-complete:50051" required:"true"`
		PartialUri  string `envconfig:"API_KIWI_TREE_PARTIAL_URI" default:"kiwi-tree-partial:50051" required:"true"`
	}

	Db struct {
		Uri      string `envconfig:"DB_URI" default:"mongodb://localhost:27017/?retryWrites=true&w=majority" required:"true"`
		Name     string `envconfig:"DB_NAME" default:"subscriptions" required:"true"`
		UserName string `envconfig:"DB_USERNAME" default:""`
		Password string `envconfig:"DB_PASSWORD" default:""`
		Table    struct {
			Name string `envconfig:"DB_NAME" default:"subscriptions" required:"true"`
		}
		Tls struct {
			Enabled  bool `envconfig:"DB_TLS_ENABLED" default:"false" required:"true"`
			Insecure bool `envconfig:"DB_TLS_INSECURE" default:"false" required:"true"`
		}
	}
)

func NewConfigFromEnv() (cfg Config, err error) {
	err = envconfig.Process("", &cfg)
	return
}
