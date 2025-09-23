package music

import (
	"runtime"
	"strings"
	"testing"
)

func TestGetNowPlayingInfo(t *testing.T) {
	// Test Case 1: Command succeeds with expected output
	t.Run("successful command execution", func(t *testing.T) {
		var tool, expectedOutput string
		var args []string

		// Use a simple, reliable command that works on most systems
		if runtime.GOOS == "windows" {
			tool = "cmd"
			args = []string{"/C", "echo", "Hello, World!"}
			expectedOutput = "Hello, World!"
		} else {
			tool = "echo"
			args = []string{"Hello, World!"}
			expectedOutput = "Hello, World!"
		}

		output, err := GetNowPlayingInfo(tool, args...)

		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}

		if strings.TrimSpace(output) != expectedOutput {
			t.Errorf("Expected output %q, but got %q", expectedOutput, strings.TrimSpace(output))
		}
	})

	// Test Case 2: Command fails
	t.Run("failed command execution", func(t *testing.T) {
		// Use a command that is guaranteed to fail
		tool := "nonexistentcommand"
		args := []string{}

		output, err := GetNowPlayingInfo(tool, args...)

		if err == nil {
			t.Error("Expected an error, but got nil")
		}

		if output != "" {
			t.Errorf("Expected empty output on failure, but got %q", output)
		}
	})
}
