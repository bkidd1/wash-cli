package screenshot

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Screenshot handles screenshot capture operations
type Screenshot struct {
	outputDir string
}

// NewScreenshot creates a new screenshot instance
func NewScreenshot() (*Screenshot, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	outputDir := filepath.Join(homeDir, ".wash-notes", "screenshots")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create screenshots directory: %w", err)
	}

	return &Screenshot{
		outputDir: outputDir,
	}, nil
}

// Capture takes a screenshot and saves it to the output directory
func (s *Screenshot) Capture(ctx context.Context) (string, error) {
	// TODO: Implement platform-specific screenshot capture
	// For now, create a placeholder file
	filename := fmt.Sprintf("screenshot-%s.txt", time.Now().Format("20060102-150405"))
	path := filepath.Join(s.outputDir, filename)

	if err := os.WriteFile(path, []byte("Screenshot placeholder"), 0644); err != nil {
		return "", fmt.Errorf("failed to create screenshot file: %w", err)
	}

	return path, nil
}
