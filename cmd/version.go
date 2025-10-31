package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	versionInfo = struct {
		Version   string
		BuildTime string
		GitCommit string
	}{
		Version:   "1.0.0",
		BuildTime: "unknown",
		GitCommit: "unknown",
	}
)

// SetVersionInfo sets version information from main package
func SetVersionInfo(version, buildTime, gitCommit string) {
	versionInfo.Version = version
	versionInfo.BuildTime = buildTime
	versionInfo.GitCommit = gitCommit
}

func NewVersionCommand() *cobra.Command {
	var verbose bool
	
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s v%s\n", AppName, versionInfo.Version)
			
			if verbose {
				fmt.Println()
				fmt.Printf("Version:    %s\n", versionInfo.Version)
				fmt.Printf("Build Time: %s\n", versionInfo.BuildTime)
				fmt.Printf("Git Commit: %s\n", versionInfo.GitCommit)
				fmt.Printf("Go Version: %s\n", runtime.Version())
				fmt.Printf("OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
			}
		},
	}
	
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed version information")
	
	return cmd
}
