package cmd

import (
	"fmt"

	"RCMCli/internal/rcm"

	"github.com/spf13/cobra"
)

func NewListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List RCM devices",
		RunE: func(cmd *cobra.Command, args []string) error {
			globalLogger.Info("Listing RCM devices...")

			rcmHandler := rcm.New()
			defer rcmHandler.Close()

			devices, err := rcmHandler.ListDevices()
			if err != nil {
				globalLogger.Error("Failed to list devices")
				return err
			}

			if len(devices) == 0 {
				globalLogger.Error("No RCM devices found")
				return fmt.Errorf("no devices found")
			}

			globalLogger.Success("Found %d device(s)", len(devices))
			for i, device := range devices {
				fmt.Printf("  %d. %s\n", i+1, device)
			}

			return nil
		},
	}
}
