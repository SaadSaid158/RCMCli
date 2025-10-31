package cmd

import (
	"RCMCli/pkg/config"
	"RCMCli/pkg/logger"
)

var (
	globalConfig *config.Config
	globalLogger *logger.Logger
)

func initGlobals(verbose bool) {
	globalConfig = config.DefaultConfig()
	globalConfig.Verbose = verbose
	globalLogger = logger.New(verbose)
}
