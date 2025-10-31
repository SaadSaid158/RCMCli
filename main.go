package main

import (
	"fmt"
	"os"

	"RCMCli/cmd"
)

// Version information (set via ldflags during build)
var (
	Version   = "1.0.0"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// Set version info for commands
	cmd.SetVersionInfo(Version, BuildTime, GitCommit)
	
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "[-] Error: %v\n", err)
		os.Exit(1)
	}
}
