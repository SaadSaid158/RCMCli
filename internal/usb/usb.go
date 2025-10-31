package usb

import (
	"fmt"
	"runtime"
)

const (
	// NVIDIA Tegra RCM USB IDs
	TegraPID = 0x7321
	TegraBID = 0x0955
)

type Device struct {
	ID       string
	Bus      int
	Address  int
	Vendor   int
	Product  int
	DevPath  string
}

type Manager struct {
	devices []*Device
}

func NewManager() *Manager {
	return &Manager{
		devices: make([]*Device, 0),
	}
}

// Detect finds all Tegra RCM devices connected
func (m *Manager) Detect() ([]*Device, error) {
	m.devices = make([]*Device, 0)

	switch runtime.GOOS {
	case "linux":
		return m.detectLinux()
	case "windows":
		return m.detectWindows()
	case "darwin":
		return m.detectDarwin()
	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// detectLinux, detectWindows, and detectDarwin are implemented in platform-specific files

// HasDevice checks if any Tegra device is connected
func (m *Manager) HasDevice() bool {
	devices, err := m.Detect()
	if err != nil {
		return false
	}
	return len(devices) > 0
}

// GetDevices returns all detected devices
func (m *Manager) GetDevices() []*Device {
	return m.devices
}

// SendPayload is implemented in platform-specific files:
// - usb_linux.go for Linux
// - usb_windows.go for Windows
// - usb_darwin.go for macOS
