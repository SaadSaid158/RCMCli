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
	var listPayloads bool

	cmd := &cobra.Command{
		Use:   "launch [payload]",
		Short: "Launch a payload on the Switch",
		Long:  "Launch a payload file. Specify payload name (without .bin) or full path to .bin file",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// List available payloads if --list flag is set
			if listPayloads {
				return listAvailablePayloads()
			}

			if len(args) == 0 {
				return fmt.Errorf("payload name or path required (use --list to see available payloads)")
			}

			payload := args[0]

			// Resolve payload path - support multiple formats
			payloadPath := resolvePayloadPath(payload)

			// Verify payload exists
			if _, err := os.Stat(payloadPath); os.IsNotExist(err) {
				globalLogger.Error("Payload file not found: %s", payloadPath)
				globalLogger.Info("Use 'launch --list' to see available payloads")
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
	cmd.Flags().BoolVarP(&listPayloads, "list", "l", false, "List available payloads in payloads directory")

	return cmd
}

// resolvePayloadPath resolves the payload path from various input formats
func resolvePayloadPath(payload string) string {
	// If it's already a full path to a .bin file, use it
	if filepath.Ext(payload) == ".bin" {
		if filepath.IsAbs(payload) {
			return payload
		}
		// Relative path with .bin extension
		return payload
	}

	// Check if file exists in payloads directory with .bin extension
	payloadPath := filepath.Join(globalConfig.PayloadDir, payload+".bin")
	if _, err := os.Stat(payloadPath); err == nil {
		return payloadPath
	}

	// Check if the payload name itself exists (without adding .bin)
	payloadPath = filepath.Join(globalConfig.PayloadDir, payload)
	if _, err := os.Stat(payloadPath); err == nil {
		return payloadPath
	}

	// Default: assume it's a name and add .bin
	return filepath.Join(globalConfig.PayloadDir, payload+".bin")
}

// listAvailablePayloads lists all .bin files in the payloads directory
func listAvailablePayloads() error {
	// Ensure payload directory exists
	if err := globalConfig.EnsurePayloadDir(); err != nil {
		globalLogger.Error("Failed to access payload directory")
		return err
	}

	globalLogger.Info("Scanning payloads directory: %s", globalConfig.PayloadDir)

	// Read directory contents
	entries, err := os.ReadDir(globalConfig.PayloadDir)
	if err != nil {
		globalLogger.Error("Failed to read payloads directory: %v", err)
		return err
	}

	// Filter and display .bin files
	var payloads []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) == ".bin" {
			payloads = append(payloads, entry.Name())
		}
	}

	if len(payloads) == 0 {
		globalLogger.Info("No payload files found in %s", globalConfig.PayloadDir)
		globalLogger.Info("Use 'download' command to download payloads")
		return nil
	}

	globalLogger.Success("Found %d payload(s):", len(payloads))
	fmt.Println()
	for i, payload := range payloads {
		// Get file info for size
		filePath := filepath.Join(globalConfig.PayloadDir, payload)
		info, err := os.Stat(filePath)
		if err == nil {
			sizeKB := info.Size() / 1024
			// Remove .bin extension for display name
			displayName := payload[:len(payload)-4]
			fmt.Printf("  %d. %-20s (%s, %d KB)\n", i+1, displayName, payload, sizeKB)
		} else {
			fmt.Printf("  %d. %s\n", i+1, payload)
		}
	}
	fmt.Println()
	globalLogger.Info("Launch with: launch <name>")

	return nil
}
