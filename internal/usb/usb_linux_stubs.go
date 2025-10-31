//go:build linux

package usb

import "fmt"

// detectWindows stub for Linux
func (m *Manager) detectWindows() ([]*Device, error) {
	return m.devices, fmt.Errorf("Windows detection not available on Linux")
}

// detectDarwin stub for Linux
func (m *Manager) detectDarwin() ([]*Device, error) {
	return m.devices, fmt.Errorf("macOS detection not available on Linux")
}
