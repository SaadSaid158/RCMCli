package usb

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	// RCM Protocol Constants
	RCM_PAYLOAD_ADDRESS = 0x40010000
	INTERMEZZO_LOCATION = 0x4001F000
	PAYLOAD_START_ADDR  = 0x40010E40

	// USB Transfer sizes
	RCM_BUFFER_SIZE = 0x1000
	MAX_PAYLOAD_SIZE = 0x30000

	// Stack spray constants
	STACK_SPRAY_START = 0x40010000
	STACK_SPRAY_END   = 0x40017000
)

// RCMCommand represents an RCM command packet
type RCMCommand struct {
	Length      uint32
	Reserved    [0x28]byte
	PayloadSize uint32
	Reserved2   [0x10]byte
}

// BuildRCMPayload constructs the full RCM exploit payload
func BuildRCMPayload(payload []byte) ([]byte, error) {
	if len(payload) > MAX_PAYLOAD_SIZE {
		return nil, fmt.Errorf("payload too large: %d bytes (max: %d)", len(payload), MAX_PAYLOAD_SIZE)
	}

	// Create the RCM command header
	cmd := RCMCommand{
		Length:      0x30298, // Magic length that triggers the overflow
		PayloadSize: uint32(len(payload)),
	}

	var buf bytes.Buffer

	// Write RCM command header
	if err := binary.Write(&buf, binary.LittleEndian, cmd); err != nil {
		return nil, fmt.Errorf("failed to write RCM header: %w", err)
	}

	// Pad to align payload
	padding := make([]byte, 0x28)
	buf.Write(padding)

	// Add the intermezzo (stack pivot code)
	intermezzo := getIntermezzo()
	buf.Write(intermezzo)

	// Add more padding
	paddingSize := PAYLOAD_START_ADDR - INTERMEZZO_LOCATION - len(intermezzo)
	if paddingSize > 0 {
		buf.Write(make([]byte, paddingSize))
	}

	// Add the actual payload
	buf.Write(payload)

	return buf.Bytes(), nil
}

// getIntermezzo returns the intermezzo bytecode for stack pivoting
// This is the small ARM code that pivots the stack to our payload
func getIntermezzo() []byte {
	return []byte{
		0x44, 0x00, 0x9F, 0xE5, // ldr r0, [pc, #0x44]
		0x01, 0x11, 0xA0, 0xE3, // mov r1, #0x40000000
		0x40, 0x20, 0x9F, 0xE5, // ldr r2, [pc, #0x40]
		0x00, 0x20, 0x42, 0xE0, // sub r2, r2, r0
		0x08, 0x00, 0x00, 0xEB, // bl #0x28
		0x01, 0x01, 0xA0, 0xE3, // mov r0, #0x40000000
		0x10, 0xFF, 0x2F, 0xE1, // bx r0
		0x00, 0x00, 0x00, 0x00, // Padding
		0x00, 0x00, 0x00, 0x00, // Padding
		0x00, 0x00, 0x00, 0x00, // Padding
		0x00, 0x00, 0x00, 0x00, // Padding
		0x00, 0x00, 0x00, 0x00, // Padding
		0x00, 0x00, 0x00, 0x00, // Padding
		0x00, 0x00, 0x00, 0x00, // Padding
		0x00, 0x00, 0x00, 0x00, // Padding
		0x04, 0x30, 0x90, 0xE4, // ldr r3, [r0], #4
		0x04, 0x30, 0x81, 0xE4, // str r3, [r1], #4
		0x04, 0x20, 0x52, 0xE2, // subs r2, r2, #4
		0xFB, 0xFF, 0xFF, 0x1A, // bne #-0x14
		0x1E, 0xFF, 0x2F, 0xE1, // bx lr
		0x40, 0x0E, 0x01, 0x40, // Payload address
		0x40, 0x0E, 0x01, 0x40, // Payload address
	}
}

// ChunkPayload splits a payload into chunks for USB transfer
func ChunkPayload(data []byte, chunkSize int) [][]byte {
	var chunks [][]byte
	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}
		chunks = append(chunks, data[i:end])
	}
	return chunks
}
