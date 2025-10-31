//go:build !linux && !windows && !darwin

package usb

import "fmt"

// detectWindows stub for unsupported platforms
func (m *Manager) detectWindows() ([]*Device, error) {
	return m.devices, fmt.Errorf("Windows detection not supported on this platform")
}

// detectDarwin stub for unsupported platforms
func (m *Manager) detectDarwin() ([]*Device, error) {
	return m.devices, fmt.Errorf("macOS detection not supported on this platform")
}

// SendPayload stub for unsupported platforms
func (m *Manager) SendPayload(device *Device, payload []byte) error {
	return fmt.Errorf("payload sending not supported on this platform")
}
