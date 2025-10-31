//go:build darwin && cgo

package usb

/*
#cgo LDFLAGS: -framework IOKit -framework CoreFoundation
#include <IOKit/IOKitLib.h>
#include <IOKit/usb/IOUSBLib.h>
#include <IOKit/usb/USB.h>
#include <CoreFoundation/CoreFoundation.h>

// Helper to create matching dictionary for USB devices
CFMutableDictionaryRef createMatchingDict() {
    return IOServiceMatching(kIOUSBDeviceClassName);
}

// Helper to get device properties
int getDeviceIDs(io_service_t device, uint16_t *vendorID, uint16_t *productID) {
    CFNumberRef vendorIDRef = (CFNumberRef)IORegistryEntryCreateCFProperty(
        device, CFSTR(kUSBVendorID), kCFAllocatorDefault, 0);
    CFNumberRef productIDRef = (CFNumberRef)IORegistryEntryCreateCFProperty(
        device, CFSTR(kUSBProductID), kCFAllocatorDefault, 0);
    
    if (!vendorIDRef || !productIDRef) {
        if (vendorIDRef) CFRelease(vendorIDRef);
        if (productIDRef) CFRelease(productIDRef);
        return -1;
    }
    
    CFNumberGetValue(vendorIDRef, kCFNumberSInt16Type, vendorID);
    CFNumberGetValue(productIDRef, kCFNumberSInt16Type, productID);
    
    CFRelease(vendorIDRef);
    CFRelease(productIDRef);
    return 0;
}
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// detectDarwin finds devices via IOKit
func (m *Manager) detectDarwin() ([]*Device, error) {
	// Create matching dictionary
	matchingDict := C.createMatchingDict()
	if matchingDict == nil {
		return m.devices, fmt.Errorf("failed to create matching dictionary")
	}

	// Get iterator for matching services
	var iterator C.io_iterator_t
	kr := C.IOServiceGetMatchingServices(C.kIOMasterPortDefault, matchingDict, &iterator)
	if kr != C.KERN_SUCCESS {
		return m.devices, fmt.Errorf("failed to get matching services: %d", kr)
	}
	defer C.IOObjectRelease(iterator)

	// Iterate through devices
	deviceIndex := 0
	for {
		service := C.IOIteratorNext(iterator)
		if service == 0 {
			break
		}

		var vendorID, productID C.uint16_t
		if C.getDeviceIDs(service, &vendorID, &productID) == 0 {
			if uint16(vendorID) == uint16(TegraBID) && uint16(productID) == uint16(TegraPID) {
				device := &Device{
					ID:      fmt.Sprintf("tegra_%d", deviceIndex),
					Vendor:  TegraBID,
					Product: TegraPID,
					Address: deviceIndex,
					DevPath: fmt.Sprintf("IOService:%d", service),
				}
				m.devices = append(m.devices, device)
				deviceIndex++
			}
		}

		C.IOObjectRelease(service)
	}

	return m.devices, nil
}

// SendPayload sends the payload using IOKit USB interface
func (m *Manager) SendPayload(device *Device, payload []byte) error {
	// Build RCM payload
	rcmPayload, err := BuildRCMPayload(payload)
	if err != nil {
		return fmt.Errorf("failed to build RCM payload: %w", err)
	}

	// Find the device
	matchingDict := C.createMatchingDict()
	if matchingDict == nil {
		return fmt.Errorf("failed to create matching dictionary")
	}

	var iterator C.io_iterator_t
	kr := C.IOServiceGetMatchingServices(C.kIOMasterPortDefault, matchingDict, &iterator)
	if kr != C.KERN_SUCCESS {
		return fmt.Errorf("failed to get matching services")
	}
	defer C.IOObjectRelease(iterator)

	// Find our Tegra device
	var targetService C.io_service_t
	for {
		service := C.IOIteratorNext(iterator)
		if service == 0 {
			break
		}

		var vendorID, productID C.uint16_t
		if C.getDeviceIDs(service, &vendorID, &productID) == 0 {
			if uint16(vendorID) == uint16(TegraBID) && uint16(productID) == uint16(TegraPID) {
				targetService = service
				break
			}
		}
		C.IOObjectRelease(service)
	}

	if targetService == 0 {
		return fmt.Errorf("device not found")
	}
	defer C.IOObjectRelease(targetService)

	// Create plugin interface
	var plugInInterface *C.IOCFPlugInInterface
	var score C.SInt32
	kr = C.IOCreatePlugInInterfaceForService(
		targetService,
		C.kIOUSBDeviceUserClientTypeID,
		C.kIOCFPlugInInterfaceID,
		&plugInInterface,
		&score,
	)
	if kr != C.KERN_SUCCESS || plugInInterface == nil {
		return fmt.Errorf("failed to create plugin interface: %d", kr)
	}
	defer C.IODestroyPlugInInterface(plugInInterface)

	// Query for device interface
	var deviceInterface *C.IOUSBDeviceInterface320
	var iid C.CFUUIDBytes = C.CFUUIDGetUUIDBytes(C.CFUUIDGetConstantUUIDWithBytes(
		nil,
		0x2d, 0x97, 0x86, 0xc6,
		0x9e, 0xf3,
		0x11, 0xd4,
		0xad, 0x51, 0x00, 0x0a, 0x27, 0x05, 0x28, 0x61,
	))

	result := (*plugInInterface).QueryInterface(
		unsafe.Pointer(plugInInterface),
		iid,
		(*unsafe.Pointer)(unsafe.Pointer(&deviceInterface)),
	)
	if result != 0 || deviceInterface == nil {
		return fmt.Errorf("failed to get device interface: %d", result)
	}
	defer (*deviceInterface).Release(unsafe.Pointer(deviceInterface))

	// Open device
	kr = (*deviceInterface).USBDeviceOpen(unsafe.Pointer(deviceInterface))
	if kr != C.KERN_SUCCESS {
		return fmt.Errorf("failed to open device: %d (try running with sudo)", kr)
	}
	defer (*deviceInterface).USBDeviceClose(unsafe.Pointer(deviceInterface))

	// Get configuration
	var configNum C.UInt8
	kr = (*deviceInterface).GetConfiguration(unsafe.Pointer(deviceInterface), &configNum)
	if kr != C.KERN_SUCCESS {
		return fmt.Errorf("failed to get configuration: %d", kr)
	}

	// Set configuration if needed
	if configNum == 0 {
		kr = (*deviceInterface).SetConfiguration(unsafe.Pointer(deviceInterface), 1)
		if kr != C.KERN_SUCCESS {
			return fmt.Errorf("failed to set configuration: %d", kr)
		}
	}

	// Send payload in chunks
	chunks := ChunkPayload(rcmPayload, RCM_BUFFER_SIZE)
	for i, chunk := range chunks {
		if err := m.sendChunkIOKit(deviceInterface, chunk); err != nil {
			return fmt.Errorf("failed to send chunk %d: %w", i, err)
		}
	}

	// Send high buffer
	highBuffer := make([]byte, 0x1000)
	if err := m.sendChunkIOKit(deviceInterface, highBuffer); err != nil {
		return fmt.Errorf("failed to send high buffer: %w", err)
	}

	return nil
}

// sendChunkIOKit sends a chunk via IOKit device request
func (m *Manager) sendChunkIOKit(deviceInterface *C.IOUSBDeviceInterface320, data []byte) error {
	var request C.IOUSBDevRequest
	request.bmRequestType = C.USBmakebmRequestType(C.kUSBOut, C.kUSBStandard, C.kUSBDevice)
	request.bRequest = 0
	request.wValue = 0
	request.wIndex = 0
	request.wLength = C.UInt16(len(data))
	request.pData = unsafe.Pointer(&data[0])
	request.wLenDone = 0

	kr := (*deviceInterface).DeviceRequest(unsafe.Pointer(deviceInterface), &request)
	if kr != C.KERN_SUCCESS {
		return fmt.Errorf("device request failed: %d", kr)
	}

	return nil
}
