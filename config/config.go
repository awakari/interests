package config

import (
	"golang.org/x/exp/slog"
	"os"
	"strconv"
)

type (
	Config struct {
		//
		Api struct {
			//
			Port uint16
			//
			KiwiTree KiwiTree
		}
		//
		Db Db
		//
		Log struct {
			//
			Level slog.Level
		}
	}

	KiwiTree struct {
		//
		CompleteUri string
		//
		PartialUri string
	}

	Db struct {
		//
		Uri string
		//
		Name string
		//
		Table struct {
			//
			Name string
		}
	}
)

const (
	envApiPort                = "API_PORT"
	defApiPort                = "8080"
	envApiKiwiTreeCompleteUri = "API_KIWI_TREE_COMPLETE_URI"
	defApiKiwiTreeCompleteUri = "kiwi-tree-complete:8080"
	envApiKiwiTreePartialUri  = "API_KIWI_TREE_PARTIAL_URI"
	defApiKiwiTreePartialUri  = "kiwi-tree-partial:8080"
	envDbUri                  = "DB_URI"
	defDbUri                  = "mongodb+srv://localhost/?retryWrites=true&w=majority"
	envDbName                 = "DB_NAME"
	defDbName                 = "subscriptions"
	envDbTableName            = "DB_TABLE_NAME"
	defDbTableName            = "subscriptions"
	envLogLevel               = "LOG_LEVEL"
	defLogLevel               = "-4"
)

func NewConfigFromEnv() (cfg Config, err error) {
	apiPortStr := getEnvOrDefault(envApiPort, defApiPort)
	var apiPort uint64
	apiPort, err = strconv.ParseUint(apiPortStr, 10, 16)
	if err != nil {
		return
	}
	cfg.Api.Port = uint16(apiPort)
	cfg.Api.KiwiTree.CompleteUri = getEnvOrDefault(envApiKiwiTreeCompleteUri, defApiKiwiTreeCompleteUri)
	cfg.Api.KiwiTree.PartialUri = getEnvOrDefault(envApiKiwiTreePartialUri, defApiKiwiTreePartialUri)
	cfg.Db.Uri = getEnvOrDefault(envDbUri, defDbUri)
	cfg.Db.Name = getEnvOrDefault(envDbName, defDbName)
	cfg.Db.Table.Name = getEnvOrDefault(envDbTableName, defDbTableName)
	logLevelStr := getEnvOrDefault(envLogLevel, defLogLevel)
	var logLevel int64
	logLevel, err = strconv.ParseInt(logLevelStr, 10, 16)
	if err != nil {
		return
	}
	cfg.Log.Level = slog.Level(logLevel)
	return
}

func getEnvOrDefault(envKey string, defVal string) (val string) {
	val = os.Getenv(envKey)
	if val == "" {
		val = defVal
	}
	return
}
