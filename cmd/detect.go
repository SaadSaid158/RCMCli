package cmd

import (
	"fmt"

	"RCMCli/internal/rcm"

	"github.com/spf13/cobra"
)

func NewDetectCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "detect",
		Short: "Detect Switch in RCM mode",
		RunE: func(cmd *cobra.Command, args []string) error {
			globalLogger.Info("Scanning for devices...")

			rcmHandler := rcm.New()
			defer rcmHandler.Close()

			if rcmHandler.DetectDevice() {
				globalLogger.Success("Nintendo Switch detected in RCM mode")
				return nil
			}

			globalLogger.Error("No Switch in RCM mode detected")
			return fmt.Errorf("device not found")
		},
	}
}
