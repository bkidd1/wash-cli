package platform

import (
	"runtime"
)

// OS represents the operating system type
type OS string

const (
	Darwin  OS = "darwin"
	Linux   OS = "linux"
	Windows OS = "windows"
)

// CurrentOS returns the current operating system
func CurrentOS() OS {
	return OS(runtime.GOOS)
}

// IsSupported returns whether the current OS is supported
func IsSupported() bool {
	switch CurrentOS() {
	case Darwin, Linux, Windows:
		return true
	default:
		return false
	}
}

// SupportsWindowCapture returns whether the current OS supports window-specific screenshot capture
func SupportsWindowCapture() bool {
	// Currently, only macOS has reliable window capture support
	// Linux and Windows implementations might be less reliable or require additional setup
	return CurrentOS() == Darwin
}

// GetOSName returns a human-readable name for the current OS
func GetOSName() string {
	switch CurrentOS() {
	case Darwin:
		return "macOS"
	case Linux:
		return "Linux"
	case Windows:
		return "Windows"
	default:
		return string(CurrentOS())
	}
}
