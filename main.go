package main

import (
	"log"

	"github.com/ItsJimi/casa/cmd"
	"github.com/ItsJimi/casa/logger"
	_ "github.com/lib/pq"
)

func main() {
	config := logger.Configuration{
		EnableConsole:     true,
		ConsoleLevel:      logger.Debug,
		ConsoleJSONFormat: false,
		EnableFile:        true,
		FileLevel:         logger.Info,
		FileJSONFormat:    true,
		FileLocation:      "log.log",
	}

	err := logger.NewLogger(config, logger.InstanceZapLogger)
	if err != nil {
		log.Fatalf("Could not instantiate log %s", err.Error())
	}

	contextLogger := logger.WithFields(logger.Fields{})
	contextLogger.Debugf("Start casa")

	cmd.Execute()
}
