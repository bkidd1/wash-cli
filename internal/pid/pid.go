package pid

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// PIDManager handles PID file operations
type PIDManager struct {
	pidFile string
}

// NewPIDManager creates a new PID manager
func NewPIDManager(pidFile string) *PIDManager {
	return &PIDManager{
		pidFile: pidFile,
	}
}

// WritePID writes the current process ID to the PID file
func (p *PIDManager) WritePID() error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(p.pidFile), 0755); err != nil {
		return fmt.Errorf("failed to create PID directory: %w", err)
	}

	// Write PID to file
	pid := os.Getpid()
	if err := os.WriteFile(p.pidFile, []byte(fmt.Sprintf("%d\n", pid)), 0644); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	fmt.Printf("Debug: Wrote PID %d to file %s\n", pid, p.pidFile)
	return nil
}

// CheckRunning checks if a process is already running
func (p *PIDManager) CheckRunning() (int, error) {
	// Check if PID file exists
	if _, err := os.Stat(p.pidFile); os.IsNotExist(err) {
		fmt.Printf("Debug: PID file %s does not exist\n", p.pidFile)
		return 0, nil
	}

	// Read PID from file
	pidBytes, err := os.ReadFile(p.pidFile)
	if err != nil {
		fmt.Printf("Debug: Failed to read PID file: %v\n", err)
		// Can't read PID file, assume no running instance
		os.Remove(p.pidFile)
		return 0, nil
	}

	// Clean up the PID string and convert to integer
	pidStr := string(pidBytes)
	pidStr = strings.TrimSpace(pidStr)
	pidStr = strings.TrimSuffix(pidStr, "%")
	fmt.Printf("Debug: Read PID string: '%s'\n", pidStr)

	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		fmt.Printf("Debug: Invalid PID format: %v\n", err)
		// Invalid PID in file, clean up
		os.Remove(p.pidFile)
		return 0, nil
	}

	// Check if process exists and is running
	process, err := os.FindProcess(pid)
	if err != nil {
		fmt.Printf("Debug: Process not found: %v\n", err)
		// Process not found, clean up
		os.Remove(p.pidFile)
		return 0, nil
	}

	// On Unix systems, FindProcess always succeeds, so we need to check if the process is actually running
	if err := process.Signal(syscall.Signal(0)); err != nil {
		fmt.Printf("Debug: Process not running: %v\n", err)
		// Process not running, clean up
		os.Remove(p.pidFile)
		return 0, nil
	}

	fmt.Printf("Debug: Found running process with PID %d\n", pid)
	return pid, nil
}

// Cleanup removes the PID file if it belongs to the current process
func (p *PIDManager) Cleanup() error {
	if pid, err := p.CheckRunning(); err == nil && pid == os.Getpid() {
		fmt.Printf("Debug: Cleaning up PID file for process %d\n", pid)
		return os.Remove(p.pidFile)
	}
	return nil
}
