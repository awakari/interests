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
			Matchers Matchers
		}
		//
		Db Db
		//
		Log struct {
			//
			Level slog.Level
		}
	}

	Matchers struct {
		//
		UriExcludesComplete string
		//
		UriExcludesPartial string
		//
		UriIncludesComplete string
		//
		UriIncludesPartial string
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
	envApiPort                        = "API_PORT"
	defApiPort                        = "8080"
	envApiMatchersUriExcludesComplete = "API_MATCHERS_URI_EXCLUDES_COMPLETE"
	defApiMatchersUriExcludesComplete = "matchers-excludes-complete:8080"
	envApiMatchersUriExcludesPartial  = "API_MATCHERS_URI_EXCLUDES_PARTIAL"
	defApiMatchersUriExcludesPartial  = "matchers-excludes-partial:8080"
	envApiMatchersUriIncludesComplete = "API_MATCHERS_URI_INCLUDES_COMPLETE"
	defApiMatchersUriIncludesComplete = "matchers-includes-complete:8080"
	envApiMatchersUriIncludesPartial  = "API_MATCHERS_URI_INCLUDES_PARTIAL"
	defApiMatchersUriIncludesPartial  = "matchers-includes-partial:8080"
	envDbUri                          = "DB_URI"
	defDbUri                          = "mongodb+srv://localhost/?retryWrites=true&w=majority"
	envDbName                         = "DB_NAME"
	defDbName                         = "subscriptions"
	envDbTableName                    = "DB_TABLE_NAME"
	defDbTableName                    = "subscriptions"
	envLogLevel                       = "LOG_LEVEL"
	defLogLevel                       = "-4"
)

func NewConfigFromEnv() (cfg Config, err error) {
	apiPortStr := getEnvOrDefault(envApiPort, defApiPort)
	var apiPort uint64
	apiPort, err = strconv.ParseUint(apiPortStr, 10, 16)
	if err != nil {
		return
	}
	cfg.Api.Port = uint16(apiPort)
	cfg.Api.Matchers.UriExcludesComplete = getEnvOrDefault(envApiMatchersUriExcludesComplete, defApiMatchersUriExcludesComplete)
	cfg.Api.Matchers.UriExcludesPartial = getEnvOrDefault(envApiMatchersUriExcludesPartial, defApiMatchersUriExcludesPartial)
	cfg.Api.Matchers.UriIncludesComplete = getEnvOrDefault(envApiMatchersUriIncludesComplete, defApiMatchersUriIncludesComplete)
	cfg.Api.Matchers.UriIncludesPartial = getEnvOrDefault(envApiMatchersUriIncludesPartial, defApiMatchersUriIncludesPartial)
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
