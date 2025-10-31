//go:build darwin && !cgo

package usb

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// detectDarwin finds devices via system_profiler command
func (m *Manager) detectDarwin() ([]*Device, error) {
	// Use system_profiler to list USB devices
	cmd := exec.Command("system_profiler", "SPUSBDataType", "-xml")
	output, err := cmd.Output()
	if err != nil {
		// Fallback: try ioreg
		return m.detectDarwinIOReg()
	}

	// Parse the output for Tegra devices
	outputStr := string(output)
	if strings.Contains(outputStr, "0x0955") && strings.Contains(outputStr, "0x7321") {
		device := &Device{
			ID:      "tegra_0",
			Vendor:  TegraBID,
			Product: TegraPID,
			Address: 0,
			DevPath: "/dev/usb",
		}
		m.devices = append(m.devices, device)
	}

	return m.devices, nil
}

// detectDarwinIOReg uses ioreg as fallback
func (m *Manager) detectDarwinIOReg() ([]*Device, error) {
	cmd := exec.Command("ioreg", "-p", "IOUSB", "-l", "-w", "0")
	output, err := cmd.Output()
	if err != nil {
		return m.devices, nil // Return empty list if command fails
	}

	// Look for Tegra device in ioreg output
	lines := bytes.Split(output, []byte("\n"))
	vendorRe := regexp.MustCompile(`"idVendor"\s*=\s*(\d+)`)
	productRe := regexp.MustCompile(`"idProduct"\s*=\s*(\d+)`)

	var currentVendor, currentProduct int
	deviceIndex := 0

	for _, line := range lines {
		lineStr := string(line)

		if vendorMatch := vendorRe.FindStringSubmatch(lineStr); vendorMatch != nil {
			currentVendor, _ = strconv.Atoi(vendorMatch[1])
		}

		if productMatch := productRe.FindStringSubmatch(lineStr); productMatch != nil {
			currentProduct, _ = strconv.Atoi(productMatch[1])

			// Check if this is a Tegra RCM device (VID: 2389 = 0x0955, PID: 29473 = 0x7321)
			if currentVendor == 2389 && currentProduct == 29473 {
				device := &Device{
					ID:      fmt.Sprintf("tegra_%d", deviceIndex),
					Vendor:  TegraBID,
					Product: TegraPID,
					Address: deviceIndex,
					DevPath: fmt.Sprintf("/dev/usb/tegra_%d", deviceIndex),
				}
				m.devices = append(m.devices, device)
				deviceIndex++
			}
		}
	}

	return m.devices, nil
}

// SendPayload sends the payload to the device using direct USB access
func (m *Manager) SendPayload(device *Device, payload []byte) error {
	// macOS requires IOKit framework for direct USB access
	// Since we're avoiding CGO, we'll use a helper approach
	
	// Build the RCM exploit payload
	rcmPayload, err := BuildRCMPayload(payload)
	if err != nil {
		return fmt.Errorf("failed to build RCM payload: %w", err)
	}

	// Try to use TegraRcmSmash if available (common macOS tool)
	if err := m.sendViaTegraRcmSmash(rcmPayload); err == nil {
		return nil
	}

	// Otherwise, provide instructions
	return fmt.Errorf("macOS payload sending requires either:\n" +
		"1. Install TegraRcmSmash: brew install tegrarcmsmash\n" +
		"2. Build with CGO enabled for native IOKit support\n" +
		"3. Use a pre-built binary with IOKit support\n" +
		"Run with sudo for USB device access")
}

// sendViaTegraRcmSmash attempts to use the TegraRcmSmash tool if installed
func (m *Manager) sendViaTegraRcmSmash(payload []byte) error {
	// Check if TegraRcmSmash is installed
	_, err := exec.LookPath("TegraRcmSmash")
	if err != nil {
		return fmt.Errorf("TegraRcmSmash not found")
	}

	// Write payload to temp file
	tmpFile := "/tmp/rcm_payload.bin"
	if err := os.WriteFile(tmpFile, payload, 0600); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	defer os.Remove(tmpFile)

	// Execute TegraRcmSmash
	cmd := exec.Command("TegraRcmSmash", tmpFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("TegraRcmSmash failed: %w\nOutput: %s", err, output)
	}

	return nil
}
