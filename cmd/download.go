package cmd

import (
	"RCMCli/pkg/downloader"

	"github.com/spf13/cobra"
)

func NewDownloadCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "download [payload]",
		Short: "Download a payload",
		Long:  "Download and extract a payload (hekate, atmosphere, lockpickrcm, briccmii, memloader)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			payload := args[0]

			// Ensure payload directory exists
			if err := globalConfig.EnsurePayloadDir(); err != nil {
				globalLogger.Error("Failed to create payload directory")
				return err
			}

			// Create downloader
			dl := downloader.New(globalLogger)
			if err := dl.Download(payload, globalConfig.PayloadDir); err != nil {
				globalLogger.Error("Download failed: %v", err)
				return err
			}

			return nil
		},
	}

	return cmd
}
