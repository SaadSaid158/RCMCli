//go:build linux

package usb

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

const (
	// USB IOCTLs for Linux
	USBDEVFS_RESET        = 0x00005514
	USBDEVFS_CLAIMINTERFACE = 0x8004550F
	USBDEVFS_RELEASEINTERFACE = 0x80045510
	USBDEVFS_BULK         = 0xC0185502
	USBDEVFS_CONTROL      = 0xC0185500

	// USB endpoints
	USB_ENDPOINT_OUT = 0x01
	USB_ENDPOINT_IN  = 0x81

	// USB request types
	USB_TYPE_STANDARD = 0x00 << 5
	USB_RECIP_INTERFACE = 0x01
	USB_DIR_OUT = 0x00
	USB_DIR_IN = 0x80
)

// usbdevfs_urb represents a USB Request Block for Linux
type usbdevfs_urb struct {
	urb_type   uint8
	endpoint   uint8
	status     int32
	flags      uint32
	buffer     uintptr
	buffer_length int32
	actual_length int32
	start_frame   int32
	number_of_packets int32
	error_count   int32
	signr         uint32
	usercontext   uintptr
}

// usbdevfs_bulktransfer represents a bulk transfer request
type usbdevfs_bulktransfer struct {
	ep      uint32
	len     uint32
	timeout uint32
	data    uintptr
}

// usbdevfs_ctrltransfer represents a control transfer request
type usbdevfs_ctrltransfer struct {
	bRequestType uint8
	bRequest     uint8
	wValue       uint16
	wIndex       uint16
	wLength      uint16
	timeout      uint32
	data         uintptr
}

// SendPayload sends the payload to the device using Linux USB device files
func (m *Manager) SendPayload(device *Device, payload []byte) error {
	// Find the USB device file
	devPath, err := m.findDeviceFile(device)
	if err != nil {
		return fmt.Errorf("failed to find device file: %w", err)
	}

	// Open the USB device
	fd, err := syscall.Open(devPath, syscall.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("failed to open device %s: %w (try running with sudo)", devPath, err)
	}
	defer syscall.Close(fd)

	// Claim interface 0
	iface := uint32(0)
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), USBDEVFS_CLAIMINTERFACE, uintptr(unsafe.Pointer(&iface)))
	if errno != 0 {
		return fmt.Errorf("failed to claim interface: %v", errno)
	}
	defer func() {
		syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), USBDEVFS_RELEASEINTERFACE, uintptr(unsafe.Pointer(&iface)))
	}()

	// Build the RCM exploit payload
	rcmPayload, err := BuildRCMPayload(payload)
	if err != nil {
		return fmt.Errorf("failed to build RCM payload: %w", err)
	}

	// Send the payload in chunks
	chunks := ChunkPayload(rcmPayload, RCM_BUFFER_SIZE)
	for i, chunk := range chunks {
		if err := m.bulkWrite(fd, USB_ENDPOINT_OUT, chunk, 1000); err != nil {
			return fmt.Errorf("failed to send chunk %d: %w", i, err)
		}
	}

	// Send the high buffer to trigger the exploit
	highBuffer := make([]byte, 0x1000)
	if err := m.bulkWrite(fd, USB_ENDPOINT_OUT, highBuffer, 1000); err != nil {
		return fmt.Errorf("failed to send high buffer: %w", err)
	}

	return nil
}

// bulkWrite performs a bulk write transfer
func (m *Manager) bulkWrite(fd int, endpoint uint8, data []byte, timeout uint32) error {
	transfer := usbdevfs_bulktransfer{
		ep:      uint32(endpoint),
		len:     uint32(len(data)),
		timeout: timeout,
		data:    uintptr(unsafe.Pointer(&data[0])),
	}

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), USBDEVFS_BULK, uintptr(unsafe.Pointer(&transfer)))
	if errno != 0 {
		return fmt.Errorf("bulk write failed: %v", errno)
	}

	return nil
}

// findDeviceFile finds the /dev/bus/usb device file for a device
func (m *Manager) findDeviceFile(device *Device) (string, error) {
	// The device file is at /dev/bus/usb/<bus>/<device>
	devPath := fmt.Sprintf("/dev/bus/usb/%03d/%03d", device.Bus, device.Address)
	
	// Check if the file exists
	if _, err := os.Stat(devPath); err != nil {
		return "", fmt.Errorf("device file not found: %s", devPath)
	}

	return devPath, nil
}

// detectLinux finds devices via /sys/bus/usb
func (m *Manager) detectLinux() ([]*Device, error) {
	devPath := "/sys/bus/usb/devices"
	entries, err := os.ReadDir(devPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		basePath := fmt.Sprintf("%s/%s", devPath, entry.Name())
		
		// Read vendor and product IDs
		vendorData, err1 := os.ReadFile(fmt.Sprintf("%s/idVendor", basePath))
		productData, err2 := os.ReadFile(fmt.Sprintf("%s/idProduct", basePath))
		
		if err1 != nil || err2 != nil {
			continue
		}

		// Check if this is a Tegra RCM device
		if string(vendorData) == "0955\n" && string(productData) == "7321\n" {
			device := &Device{
				ID:      entry.Name(),
				Vendor:  TegraBID,
				Product: TegraPID,
				DevPath: basePath,
			}

			// Parse bus and device numbers
			if busData, err := os.ReadFile(fmt.Sprintf("%s/busnum", basePath)); err == nil {
				fmt.Sscanf(string(busData), "%d", &device.Bus)
			}
			if devData, err := os.ReadFile(fmt.Sprintf("%s/devnum", basePath)); err == nil {
				fmt.Sscanf(string(devData), "%d", &device.Address)
			}

			m.devices = append(m.devices, device)
		}
	}

	return m.devices, nil
}
