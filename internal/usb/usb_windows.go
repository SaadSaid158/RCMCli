//go:build windows

package usb

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const (
	GUID_DEVINTERFACE_USB_DEVICE = "{A5DCBF10-6530-11D2-901F-00C04FB951ED}"
	
	// Windows USB constants
	GENERIC_WRITE = 0x40000000
	GENERIC_READ  = 0x80000000
	FILE_SHARE_WRITE = 0x00000002
	FILE_SHARE_READ  = 0x00000001
	OPEN_EXISTING = 3
	FILE_FLAG_OVERLAPPED = 0x40000000
)

var (
	winusb = windows.NewLazySystemDLL("winusb.dll")
	setupapi = windows.NewLazySystemDLL("setupapi.dll")
	
	procWinUsb_Initialize = winusb.NewProc("WinUsb_Initialize")
	procWinUsb_Free = winusb.NewProc("WinUsb_Free")
	procWinUsb_WritePipe = winusb.NewProc("WinUsb_WritePipe")
	procWinUsb_SetPipePolicy = winusb.NewProc("WinUsb_SetPipePolicy")
	
	procSetupDiGetClassDevs = setupapi.NewProc("SetupDiGetClassDevsW")
	procSetupDiEnumDeviceInterfaces = setupapi.NewProc("SetupDiEnumDeviceInterfaces")
	procSetupDiGetDeviceInterfaceDetail = setupapi.NewProc("SetupDiGetDeviceInterfaceDetailW")
	procSetupDiDestroyDeviceInfoList = setupapi.NewProc("SetupDiDestroyDeviceInfoList")
)

type SP_DEVICE_INTERFACE_DATA struct {
	cbSize             uint32
	InterfaceClassGuid windows.GUID
	Flags              uint32
	Reserved           uintptr
}

type SP_DEVICE_INTERFACE_DETAIL_DATA struct {
	cbSize     uint32
	DevicePath [1]uint16
}

// detectWindows finds devices via Windows Registry and SetupAPI
func (m *Manager) detectWindows() ([]*Device, error) {
	// Open registry key for USB devices
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Enum\USB`, registry.QUERY_VALUE)
	if err != nil {
		return m.devices, nil // Return empty list if registry access fails
	}
	defer key.Close()

	// Enumerate all USB device keys
	subkeys, err := key.ReadSubKeyNames(-1)
	if err != nil {
		return m.devices, nil
	}

	deviceIndex := 0
	for _, subkey := range subkeys {
		deviceKey, err := registry.OpenKey(registry.LOCAL_MACHINE, fmt.Sprintf(`SYSTEM\CurrentControlSet\Enum\USB\%s`, subkey), registry.QUERY_VALUE)
		if err != nil {
			continue
		}
		defer deviceKey.Close()

		// Read device instances
		instances, err := deviceKey.ReadSubKeyNames(-1)
		if err != nil {
			continue
		}

		for _, instance := range instances {
			instanceKey, err := registry.OpenKey(registry.LOCAL_MACHINE, fmt.Sprintf(`SYSTEM\CurrentControlSet\Enum\USB\%s\%s`, subkey, instance), registry.QUERY_VALUE)
			if err != nil {
				continue
			}
			defer instanceKey.Close()

			// Read device properties
			hardwareID, _, err := instanceKey.GetStringsValue("HardwareID")
			if err != nil {
				continue
			}

			// Check if this is a Tegra RCM device (VID: 0955, PID: 7321)
			if len(hardwareID) > 0 && (hardwareID[0] == "USB\\VID_0955&PID_7321" || hardwareID[0] == "USB\\VID_0955&PID_7321&REV_0100") {
				device := &Device{
					ID:      fmt.Sprintf("%s\\%s", subkey, instance),
					Vendor:  TegraBID,
					Product: TegraPID,
					DevPath: fmt.Sprintf("\\\\.\\USB#VID_0955&PID_7321#%s", instance),
					Bus:     0,
					Address: deviceIndex,
				}
				m.devices = append(m.devices, device)
				deviceIndex++
			}
		}
	}

	return m.devices, nil
}

// SendPayload sends the payload to the device using WinUSB
func (m *Manager) SendPayload(device *Device, payload []byte) error {
	// Get device path using SetupAPI
	devicePath, err := m.getDevicePath(device)
	if err != nil {
		return fmt.Errorf("failed to get device path: %w", err)
	}

	// Open the device handle
	pathPtr, err := syscall.UTF16PtrFromString(devicePath)
	if err != nil {
		return fmt.Errorf("failed to convert path: %w", err)
	}

	handle, err := windows.CreateFile(
		pathPtr,
		GENERIC_WRITE|GENERIC_READ,
		FILE_SHARE_WRITE|FILE_SHARE_READ,
		nil,
		OPEN_EXISTING,
		FILE_FLAG_OVERLAPPED,
		0,
	)
	if err != nil {
		return fmt.Errorf("failed to open device: %w (ensure libusbK or WinUSB driver is installed)", err)
	}
	defer windows.CloseHandle(handle)

	// Initialize WinUSB
	var winusbHandle uintptr
	ret, _, _ := procWinUsb_Initialize.Call(
		uintptr(handle),
		uintptr(unsafe.Pointer(&winusbHandle)),
	)
	if ret == 0 {
		return fmt.Errorf("WinUsb_Initialize failed (ensure WinUSB driver is installed for Tegra device)")
	}
	defer procWinUsb_Free.Call(winusbHandle)

	// Build the RCM exploit payload
	rcmPayload, err := BuildRCMPayload(payload)
	if err != nil {
		return fmt.Errorf("failed to build RCM payload: %w", err)
	}

	// Send payload in chunks
	chunks := ChunkPayload(rcmPayload, RCM_BUFFER_SIZE)
	for i, chunk := range chunks {
		var bytesWritten uint32
		ret, _, _ := procWinUsb_WritePipe.Call(
			winusbHandle,
			uintptr(0x01), // Endpoint OUT
			uintptr(unsafe.Pointer(&chunk[0])),
			uintptr(len(chunk)),
			uintptr(unsafe.Pointer(&bytesWritten)),
			0,
		)
		if ret == 0 {
			return fmt.Errorf("failed to write chunk %d", i)
		}
	}

	// Send high buffer to trigger exploit
	highBuffer := make([]byte, 0x1000)
	var bytesWritten uint32
	ret, _, _ = procWinUsb_WritePipe.Call(
		winusbHandle,
		uintptr(0x01),
		uintptr(unsafe.Pointer(&highBuffer[0])),
		uintptr(len(highBuffer)),
		uintptr(unsafe.Pointer(&bytesWritten)),
		0,
	)
	if ret == 0 {
		return fmt.Errorf("failed to send high buffer")
	}

	return nil
}

// getDevicePath retrieves the device path for WinUSB access
func (m *Manager) getDevicePath(device *Device) (string, error) {
	// For simplicity, construct a standard device path
	// In production, you'd use SetupAPI to enumerate properly
	// This is a simplified approach that works with WinUSB driver
	return fmt.Sprintf("\\\\.\\USB#VID_0955&PID_7321#%s#{a5dcbf10-6530-11d2-901f-00c04fb951ed}", device.ID), nil
}
