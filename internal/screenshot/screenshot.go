package screenshot

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"time"

	"github.com/bkidd1/wash-cli/pkg/platform"
	"github.com/kbinani/screenshot"
)

// Screenshot represents a captured screenshot
type Screenshot struct {
	Path string
}

// Capture takes a screenshot of the specified display
func Capture(displayIndex int) (*Screenshot, error) {
	if !platform.IsSupported() {
		return nil, fmt.Errorf("screenshot capture is not supported on %s", platform.GetOSName())
	}

	// Get bounds of the display
	bounds := screenshot.GetDisplayBounds(displayIndex)

	// Capture the screenshot
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return nil, fmt.Errorf("failed to capture screenshot: %w", err)
	}

	// Create screenshots directory if it doesn't exist
	dir := filepath.Join(os.Getenv("HOME"), ".wash-screenshots")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create screenshots directory: %w", err)
	}

	// Generate filename with timestamp
	filename := fmt.Sprintf("screenshot-%s.png", time.Now().Format("2006-01-02-15-04-05"))
	path := filepath.Join(dir, filename)

	// Save the screenshot
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create screenshot file: %w", err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		return nil, fmt.Errorf("failed to encode screenshot: %w", err)
	}

	return &Screenshot{
		Path: path,
	}, nil
}

// GetDisplayCount returns the number of displays
func GetDisplayCount() int {
	if !platform.IsSupported() {
		return 0
	}
	return screenshot.NumActiveDisplays()
}

// CaptureWindow takes a screenshot of a specific window by title
func CaptureWindow(windowTitle string, outputPath string) error {
	if !platform.IsSupported() {
		return fmt.Errorf("screenshot capture is not supported on %s", platform.GetOSName())
	}

	// If window capture is not supported, fall back to full screen capture
	if !platform.SupportsWindowCapture() {
		fmt.Printf("Window-specific capture is not supported on %s. Capturing entire screen instead.\n", platform.GetOSName())
		return CaptureFullScreen(outputPath)
	}

	// For now, we'll just capture the entire primary display
	// In the future, we can add window-specific capture using platform-specific APIs
	bounds := screenshot.GetDisplayBounds(0)

	// Capture the screenshot
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return fmt.Errorf("failed to capture screenshot: %w", err)
	}

	// Create parent directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Save the screenshot
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create screenshot file: %w", err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		return fmt.Errorf("failed to encode screenshot: %w", err)
	}

	return nil
}

// CaptureFullScreen captures the entire primary display
func CaptureFullScreen(outputPath string) error {
	bounds := screenshot.GetDisplayBounds(0)
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return fmt.Errorf("failed to capture full screen: %w", err)
	}

	// Create parent directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Save the screenshot
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create screenshot file: %w", err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		return fmt.Errorf("failed to encode screenshot: %w", err)
	}

	return nil
}
