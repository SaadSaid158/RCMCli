//go:build windows

package usb

import "fmt"

// detectLinux stub for Windows
func (m *Manager) detectLinux() ([]*Device, error) {
	return m.devices, fmt.Errorf("Linux detection not available on Windows")
}

// detectDarwin stub for Windows
func (m *Manager) detectDarwin() ([]*Device, error) {
	return m.devices, fmt.Errorf("macOS detection not available on Windows")
}
