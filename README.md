# RCMCli

A simple, fast, pure Go CLI tool for launching payloads on Nintendo Switch via RCM.

## Features

- ✅ **Auto-Download**: Download and extract payloads from GitHub
- ✅ **Supported Payloads**: Hekate, Atmosphere, Lockpick RCM, BRICCMII, Memloader
- ✅ **Integrity Checking**: SHA256 verification
- ✅ **Device Detection**: Detect and list RCM devices (Linux)
- ✅ **Cross-Platform**: Linux, Windows, macOS
- ✅ **Multi-Architecture**: amd64, arm64
- ✅ **Simple Logging**: `[*]`, `[+]`, `[-]` format
- ✅ **Pure Go**: No CGO, no external dependencies (except Cobra)

## Prerequisites

- Go 1.21+
- On Linux: `go build` works out of the box (pure Go)

## Build

### Quick Build
```bash
chmod +x build.sh
./build.sh
```

### Output Structure
```
dist/
├── Linux/
│   ├── RCMCli_1.0.0_linux_amd64
│   └── RCMCli_1.0.0_linux_arm64
├── Windows/
│   ├── RCMCli_1.0.0_windows_amd64.exe
│   └── RCMCli_1.0.0_windows_arm64.exe
└── checksums.txt
```

### Verify Checksums
```bash
cd dist && sha256sum -c checksums.txt && cd ..
```

## Usage

```bash
# Download a payload
./dist/Linux/RCMCli_1.0.0_linux_amd64 download hekate

# Detect Switch
./dist/Linux/RCMCli_1.0.0_linux_amd64 detect

# List devices
./dist/Linux/RCMCli_1.0.0_linux_amd64 list

# Launch payload
./dist/Linux/RCMCli_1.0.0_linux_amd64 launch hekate

# Show version
./dist/Linux/RCMCli_1.0.0_linux_amd64 version
```

## Supported Payloads

- **hekate** - Bootloader/recovery
- **atmosphere** - Custom firmware
- **lockpickrcm** - Encryption key dumper
- **briccmii** - NAND backup tool
- **memloader** - Memory loader

## How to Enter RCM Mode

1. Power off your Nintendo Switch completely
2. Hold the VOL+ button and press the POWER button once
3. Keep holding VOL+ until you see the RCM screen (black screen)
4. Connect your Switch to your computer using a USB-C cable
5. Run the launcher

## Troubleshooting

### Device Not Detected
- Make sure your Switch is in RCM mode
- Try a different USB cable (some cables are charge-only)
- Try a different USB port

### Payload Fails to Load
- Try a different USB port
- Make sure your Switch's battery is charged (at least 20% recommended)
- Ensure the payload file exists at the specified path

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Disclaimer

This software is provided "as is" without any warranties. Use at your own risk. The authors are not responsible for any damage to your Nintendo Switch or any other devices.
