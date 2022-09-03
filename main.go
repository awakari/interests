package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
)

const (
	name           = "subscriptions"
	version        = "1.0.0"
	envApiPort     = "API_PORT"
	envDbUri       = "DB_URI"
	envDbName      = "DB_NAME"
	envDbTblName   = "DB_TBL_NAME"
	envPatternsUri = "PATTERNS_URI"
)

var (
	homePath = fmt.Sprintf("%s/.%s/%s", os.Getenv("HOME"), name, version)
)

type (
	configApi struct {
		port uint16
	}
	configDb struct {
		uri  string
		name string
		tbl  string
	}
	config struct {
		api         configApi
		db          configDb
		patternsUri string
	}
)

func main() {
	logger := logrus.WithFields(logrus.Fields{})
	logger.Info(fmt.Sprintf("%s %s", name, version))
	cfg := loadEnvConfig(logger)
	println(cfg)
}

func loadEnvConfig(logger *logrus.Entry) (cfg config) {
	apiPortRaw := os.Getenv(envApiPort)
	if apiPortRaw == "" {
		logger.Fatalf("%s value not set", envApiPort)
	}
	apiPort, err := strconv.ParseUint(apiPortRaw, 10, 16)
	if err != nil {
		logger.Fatalf("Invalid %s value: %s", envApiPort, apiPortRaw)
	}
	cfg = config{
		api: configApi{
			port: uint16(apiPort),
		},
		db: configDb{
			uri:  os.Getenv(envDbUri),
			name: os.Getenv(envDbName),
			tbl:  os.Getenv(envDbTblName),
		},
		patternsUri: os.Getenv(envPatternsUri),
	}
	if cfg.db.uri == "" {
		logger.Fatalf("%s value not set", envDbUri)
	}
	if cfg.db.name == "" {
		logger.Fatalf("%s value not set", envDbName)
	}
	if cfg.db.tbl == "" {
		logger.Fatalf("%s value not set", envDbTblName)
	}
	return
}
