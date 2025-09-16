package dependencies

import (
	"fmt"
	"os/exec"
	"runtime"
)

// Dependency represents a required command-line tool for a specific OS.
type Dependency struct {
	Command string
	OS      string
}

// Check verifies if the command-line tool exists in the system's PATH.
func (d *Dependency) Check() (bool, string) {
	if runtime.GOOS != d.OS {
		return false, fmt.Sprintf("Unsupported OS: %s for tool %s", runtime.GOOS, d.Command)
	}
	_, err := exec.LookPath(d.Command)
	if err != nil {
		return false, fmt.Sprintf("Required tool '%s' not found. Please install it.", d.Command)
	}
	return true, ""
}

// GetNowPlayingDependency returns the appropriate dependency based on the current OS.
func GetNowPlayingDependency() *Dependency {
	switch runtime.GOOS {
	case "linux":
		return &Dependency{Command: "playerctl", OS: "linux"}
	case "darwin": // macOS
		return &Dependency{Command: "nowplaying-cli", OS: "darwin"}
	default:
		return nil
	}
}

