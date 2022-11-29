package config

import (
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
	defApiMatchersUriExcludesComplete = "localhost:8080"
	envApiMatchersUriExcludesPartial  = "API_MATCHERS_URI_EXCLUDES_PARTIAL"
	defApiMatchersUriExcludesPartial  = "localhost:8080"
	envApiMatchersUriIncludesComplete = "API_MATCHERS_URI_INCLUDES_COMPLETE"
	defApiMatchersUriIncludesComplete = "localhost:8080"
	envApiMatchersUriIncludesPartial  = "API_MATCHERS_URI_INCLUDES_PARTIAL"
	defApiMatchersUriIncludesPartial  = "localhost:8080"
	envDbUri                          = "DB_URI"
	defDbUri                          = "mongodb+srv://localhost/?retryWrites=true&w=majority"
	envDbName                         = "DB_NAME"
	defDbName                         = "subscriptions"
	envDbTableName                    = "DB_TABLE_NAME"
	defDbTableName                    = "subscriptions"
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
	return
}

func getEnvOrDefault(envKey string, defVal string) (val string) {
	val = os.Getenv(envKey)
	if val == "" {
		val = defVal
	}
	return
}
