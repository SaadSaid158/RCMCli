package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"RCMCli/internal/rcm"

	"github.com/spf13/cobra"
)

func NewLaunchCommand() *cobra.Command {
	var retryCount int

	cmd := &cobra.Command{
		Use:   "launch [payload]",
		Short: "Launch a payload on the Switch",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			payload := args[0]

			// Resolve payload path
			payloadPath := filepath.Join(globalConfig.PayloadDir, payload+".bin")

			// Verify payload exists
			if _, err := os.Stat(payloadPath); os.IsNotExist(err) {
				globalLogger.Error("Payload file not found: %s", payloadPath)
				return fmt.Errorf("payload file not found")
			}

			globalLogger.Info("Launching %s...", payload)

			// Initialize RCM handler
			rcmHandler := rcm.New()
			defer rcmHandler.Close()

			// Attempt launch with retries
			for attempt := 1; attempt <= retryCount; attempt++ {
				if err := rcmHandler.LaunchPayload(payloadPath); err != nil {
					if attempt < retryCount {
						globalLogger.Info("Retrying... (%d/%d)", attempt, retryCount)
						continue
					}
					globalLogger.Error("Failed to launch payload")
					return err
				}

				globalLogger.Success("Payload launched successfully!")
				return nil
			}

			return fmt.Errorf("failed to launch after %d attempts", retryCount)
		},
	}

	cmd.Flags().IntVar(&retryCount, "retry", 3, "Number of retry attempts")

	return cmd
}
