//go:build darwin

package usb

import "fmt"

// detectLinux stub for macOS
func (m *Manager) detectLinux() ([]*Device, error) {
	return m.devices, fmt.Errorf("Linux detection not available on macOS")
}

// detectWindows stub for macOS
func (m *Manager) detectWindows() ([]*Device, error) {
	return m.devices, fmt.Errorf("Windows detection not available on macOS")
}
