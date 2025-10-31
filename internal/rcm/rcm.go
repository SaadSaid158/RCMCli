package rcm

import (
	"fmt"
	"io/ioutil"

	"RCMCli/internal/usb"
)

type RCM struct {
	manager *usb.Manager
}

func New() *RCM {
	return &RCM{
		manager: usb.NewManager(),
	}
}

func (r *RCM) Close() error {
	return nil
}

// DetectDevice checks if a Switch in RCM mode is connected
func (r *RCM) DetectDevice() bool {
	return r.manager.HasDevice()
}

// ListDevices returns a list of connected RCM devices
func (r *RCM) ListDevices() ([]string, error) {
	devices, err := r.manager.Detect()
	if err != nil {
		return nil, err
	}

	var deviceList []string
	for i, dev := range devices {
		deviceList = append(deviceList, fmt.Sprintf("Tegra RCM Device #%d (Bus: %d, Addr: %d)", i+1, dev.Bus, dev.Address))
	}

	return deviceList, nil
}

// LaunchPayload sends a payload to the Switch in RCM mode
func (r *RCM) LaunchPayload(payloadPath string) error {
	payload, err := ioutil.ReadFile(payloadPath)
	if err != nil {
		return fmt.Errorf("failed to read payload file: %w", err)
	}

	devices, err := r.manager.Detect()
	if err != nil {
		return fmt.Errorf("failed to detect devices: %w", err)
	}

	if len(devices) == 0 {
		return fmt.Errorf("no Tegra RCM device found")
	}

	device := devices[0]
	return r.manager.SendPayload(device, payload)
}
