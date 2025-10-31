package usb

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
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

// detectLinux finds devices via /sys/bus/usb
func (m *Manager) detectLinux() ([]*Device, error) {
	devPath := "/sys/bus/usb/devices"
	entries, err := ioutil.ReadDir(devPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		idVendor := filepath.Join(devPath, entry.Name(), "idVendor")
		idProduct := filepath.Join(devPath, entry.Name(), "idProduct")
		busNum := filepath.Join(devPath, entry.Name(), "busnum")
		devNum := filepath.Join(devPath, entry.Name(), "devnum")

		vendorData, _ := ioutil.ReadFile(idVendor)
		productData, _ := ioutil.ReadFile(idProduct)
		busData, _ := ioutil.ReadFile(busNum)
		devData, _ := ioutil.ReadFile(devNum)

		if string(vendorData) == "0955\n" && string(productData) == "7321\n" {
			device := &Device{
				ID:      entry.Name(),
				Vendor:  TegraBID,
				Product: TegraPID,
				DevPath: filepath.Join(devPath, entry.Name()),
			}

			// Parse bus and device numbers
			fmt.Sscanf(string(busData), "%d", &device.Bus)
			fmt.Sscanf(string(devData), "%d", &device.Address)

			m.devices = append(m.devices, device)
		}
	}

	return m.devices, nil
}

// detectWindows finds devices via Windows USB API (stub)
func (m *Manager) detectWindows() ([]*Device, error) {
	// Windows requires WinUSB or libusb with proper drivers
	// For now, return empty - would need CGO for real implementation
	return m.devices, nil
}

// detectDarwin finds devices via macOS IOKit (stub)
func (m *Manager) detectDarwin() ([]*Device, error) {
	// macOS requires IOKit framework
	// For now, return empty - would need CGO for real implementation
	return m.devices, nil
}

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

// SendPayload sends payload to device (stub - requires libusb)
func (m *Manager) SendPayload(device *Device, payload []byte) error {
	if len(m.devices) == 0 {
		return fmt.Errorf("no device found")
	}

	// Payload transfer requires low-level USB communication
	// This would need libusb bindings or Windows USB API
	// For now, provide informative error
	return fmt.Errorf("payload transfer not yet implemented in pure Go. Use CGO build for full functionality")
}
