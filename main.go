package main

import (
	"fmt"
	"github.com/meandros-messaging/subscriptions/config"
	"github.com/sirupsen/logrus"
	"os"
)

func main() {
	log := logrus.WithFields(logrus.Fields{})
	log.Info("starting...")
	if len(os.Args) != 2 {
		log.Fatalf("invalid command line arguments, try: matchers <config-file-path>")
	}
	cfgFilePath := os.Args[1]
	log.Infof("loading the configuration from the file: %s ...", cfgFilePath)
	cfgFile, err := os.Open(cfgFilePath)
	if err != nil {
		log.Fatalf("failed to open the config file %s: %s", cfgFilePath, err)
	}
	cfg, err := config.NewConfigFromYaml(cfgFile)
	if err != nil {
		log.Fatalf("failed to load the yaml config from the file %s: %s", cfgFilePath, err)
	}
	fmt.Println(cfg)
}
