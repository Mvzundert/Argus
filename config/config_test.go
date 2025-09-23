package config

import (
	"log"
	"os"
	"strings"
	"testing"

	"github.com/joho/godotenv"
)

// TestLoadValidConfig tests the happy path by loading the .env-test file
// and verifying that the configuration is loaded correctly.
func TestLoadValidConfig(t *testing.T) {
	// Temporarily load the .env-test file to set environment variables for this test.
	err := godotenv.Load("../.env-test")
	if err != nil {
		t.Fatalf("Could not load .env-test for testing: %v", err)
	}

	// Ensure that the test cleans up the environment variables after it runs.
	defer func() {
		// Get a list of the keys we've set.
		keys := []string{"TWITCH_NICK", "TWITCH_TOKEN", "TWITCH_CHANNEL", "TWITCH_CHANNEL_ID", "TWITCH_CLIENT_ID", "TWITCH_APP_ACCESS_TOKEN", "SHOW_LOGS", "PORT"}
		for _, key := range keys {
			os.Unsetenv(key)
		}
	}()

	// Capture log output to ensure no fatal error is logged.
	log.SetOutput(new(strings.Builder))

	// Call the function under test.
	cfg := Load()

	// Assert that the configuration was loaded correctly.
	if cfg.Nick != "test_nick" {
		t.Errorf("Expected Nick 'test_nick', got '%s'", cfg.Nick)
	}
	if cfg.ShowLogs != true {
		t.Errorf("Expected ShowLogs true, got %v", cfg.ShowLogs)
	}
}

