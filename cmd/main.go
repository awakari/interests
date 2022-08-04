package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
)

const (
	Name    = "synapse"
	Version = "1.0.0"
)

var (
	homePath           = fmt.Sprintf("~/.%s/%s", Name, Version)
	extPath            = fmt.Sprintf("%s/ext", homePath)
	handlerFactoryPath = fmt.Sprintf("%s/handler/factory", extPath)
)

func main() {
	log := logrus.WithFields(logrus.Fields{})
	log.Info(fmt.Sprintf("%s %s", Name, Version))
}
