package cmd

import (
	"github.com/spf13/cobra"
)

const (
	Version = "1.0.0"
	AppName = "RCMCli"
)

var (
	verbose bool
)

func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "rcmcli",
		Short: "Nintendo Switch RCM Launcher",
		Long:  `RCMCli - Launch payloads on Nintendo Switch via RCM mode`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			initGlobals(verbose)
			return nil
		},
	}

	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Enable verbose output")

	rootCmd.AddCommand(NewLaunchCommand())
	rootCmd.AddCommand(NewDetectCommand())
	rootCmd.AddCommand(NewListCommand())
	rootCmd.AddCommand(NewDownloadCommand())
	rootCmd.AddCommand(NewVersionCommand())

	return rootCmd
}

func Execute() error {
	rootCmd := NewRootCommand()
	return rootCmd.Execute()
}
