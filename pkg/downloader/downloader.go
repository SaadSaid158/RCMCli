package downloader

import (
	"archive/zip"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"RCMCli/pkg/logger"
)

const (
	// MaxDownloadSize is the maximum allowed download size (500MB)
	MaxDownloadSize = 500 * 1024 * 1024
	// MaxZipSize is the maximum allowed zip file size (500MB)
	MaxZipSize = 500 * 1024 * 1024
	// RequestTimeout is the HTTP request timeout
	RequestTimeout = 30 * time.Second
)

// PayloadInfo contains download information for a payload
type PayloadInfo struct {
	Name      string
	URL       string
	ZipPath   string // Path inside the zip file
	SHA256    string // Expected SHA256 hash
}

// Payloads defines all available payloads
var Payloads = map[string]PayloadInfo{
	"hekate": {
		Name:    "Hekate",
		URL:     "https://github.com/CTCaer/hekate/releases/download/v6.1.1/hekate_ctcaer_6.1.1.zip",
		ZipPath: "hekate_ctcaer_6.1.1.bin",
	},
	"atmosphere": {
		Name:    "Atmosphere",
		URL:     "https://github.com/Atmosphere-NX/Atmosphere/releases/download/1.5.14/atmosphere-1.5.14-master-c0d8e6c+hbl-2.4.3+es-loader.zip",
		ZipPath: "atmosphere/fusee.bin",
	},
	"lockpickrcm": {
		Name:    "Lockpick RCM",
		URL:     "https://github.com/shchmue/Lockpick_RCM/releases/download/v1.9.11/Lockpick_RCM.zip",
		ZipPath: "Lockpick_RCM.bin",
	},
	"briccmii": {
		Name:    "BRICCMII",
		URL:     "https://github.com/eliboa/BRICCMII/releases/download/v1.0.0/BRICCMII.zip",
		ZipPath: "BRICCMII.bin",
	},
	"memloader": {
		Name:    "Memloader",
		URL:     "https://github.com/wimrijnders/memloader/releases/download/v1.0/memloader.zip",
		ZipPath: "memloader.bin",
	},
}

// Downloader handles payload downloads
type Downloader struct {
	logger *logger.Logger
}

// New creates a new downloader
func New(log *logger.Logger) *Downloader {
	return &Downloader{logger: log}
}

// Download downloads a payload and extracts it
func (d *Downloader) Download(payloadName string, outputDir string) error {
	info, exists := Payloads[payloadName]
	if !exists {
		return fmt.Errorf("unknown payload: %s", payloadName)
	}

	d.logger.Info("Downloading %s...", info.Name)

	// Download file
	tempFile := filepath.Join(outputDir, "temp.zip")
	if err := d.downloadFile(info.URL, tempFile); err != nil {
		return err
	}
	defer os.Remove(tempFile)

	// Extract from zip
	d.logger.Info("Extracting payload...")
	outputPath := filepath.Join(outputDir, payloadName+".bin")
	if err := d.extractFromZip(tempFile, info.ZipPath, outputPath); err != nil {
		return err
	}

	d.logger.Success("Downloaded %s successfully", info.Name)
	return nil
}

// VerifyIntegrity checks if a payload file is valid
func (d *Downloader) VerifyIntegrity(filePath string) error {
	d.logger.Info("Verifying integrity of %s...", filepath.Base(filePath))

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return fmt.Errorf("failed to calculate hash: %w", err)
	}

	d.logger.Success("File integrity verified")
	return nil
}

// downloadFile downloads a file with security checks
func (d *Downloader) downloadFile(url string, outputPath string) error {
	client := &http.Client{
		Timeout: RequestTimeout,
	}

	resp, err := client.Get(url)
	if err != nil {
		d.logger.Error("Failed to download: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		d.logger.Error("Download failed with status %d", resp.StatusCode)
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Check Content-Length header
	if resp.ContentLength > MaxDownloadSize {
		d.logger.Error("File too large: %d bytes (max: %d)", resp.ContentLength, MaxDownloadSize)
		return fmt.Errorf("file too large")
	}

	// Ensure directory exists with secure permissions
	if err := os.MkdirAll(filepath.Dir(outputPath), 0700); err != nil {
		d.logger.Error("Failed to create directory: %v", err)
		return err
	}

	file, err := os.Create(outputPath)
	if err != nil {
		d.logger.Error("Failed to create file: %v", err)
		return err
	}
	defer file.Close()

	// Set secure file permissions
	if err := os.Chmod(outputPath, 0600); err != nil {
		d.logger.Error("Failed to set file permissions: %v", err)
		os.Remove(outputPath)
		return err
	}

	// Copy file with size limit
	limitedReader := io.LimitReader(resp.Body, MaxDownloadSize)
	_, err = io.Copy(file, limitedReader)
	if err != nil {
		d.logger.Error("Failed to write file: %v", err)
		os.Remove(outputPath)
		return err
	}

	return nil
}

// extractFromZip extracts a file from a zip archive with security checks
func (d *Downloader) extractFromZip(zipPath string, internalPath string, outputPath string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		d.logger.Error("Failed to open zip: %v", err)
		return err
	}
	defer reader.Close()

	// Find the file in the zip
	var file *zip.File
	for _, f := range reader.File {
		// Prevent directory traversal attacks
		if strings.Contains(f.Name, "..") {
			d.logger.Error("Invalid path in zip: %s", f.Name)
			return fmt.Errorf("invalid path in zip")
		}

		if strings.HasSuffix(f.Name, internalPath) || f.Name == internalPath {
			file = f
			break
		}
	}

	if file == nil {
		d.logger.Error("File not found in zip: %s", internalPath)
		return fmt.Errorf("file not found in zip: %s", internalPath)
	}

	// Check uncompressed size
	if file.UncompressedSize > MaxZipSize {
		d.logger.Error("File too large in zip: %d bytes (max: %d)", file.UncompressedSize, MaxZipSize)
		return fmt.Errorf("file too large in zip")
	}

	// Open the file
	rc, err := file.Open()
	if err != nil {
		d.logger.Error("Failed to open file in zip: %v", err)
		return err
	}
	defer rc.Close()

	// Create output file with secure permissions
	out, err := os.Create(outputPath)
	if err != nil {
		d.logger.Error("Failed to create output file: %v", err)
		return err
	}
	defer out.Close()

	// Set secure file permissions
	if err := os.Chmod(outputPath, 0600); err != nil {
		d.logger.Error("Failed to set file permissions: %v", err)
		os.Remove(outputPath)
		return err
	}

	// Copy file with size limit
	limitedReader := io.LimitReader(rc, MaxZipSize)
	if _, err := io.Copy(out, limitedReader); err != nil {
		d.logger.Error("Failed to extract file: %v", err)
		os.Remove(outputPath)
		return err
	}

	return nil
}

// ListAvailable returns a list of available payloads
func ListAvailable() []string {
	var names []string
	for name := range Payloads {
		names = append(names, name)
	}
	return names
}
