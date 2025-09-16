package music

import (
	"fmt"
	"os/exec"
	"strings"
)

// GetNowPlayingInfo retrieves the currently playing media information.
func GetNowPlayingInfo(tool string, args ...string) (string, error) {
	cmd := exec.Command(tool, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to execute command '%s %s': %w\nOutput: %s", tool, strings.Join(args, " "), err, string(output))
	}

	trackInfo := strings.TrimSpace(string(output))
	return trackInfo, nil
}

