package config

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds application configuration and secrets.
type Config struct {
	Nick           string
	OAuthToken     string
	ClientID       string
	AppAccessToken string
	Channel        string
	ChannelID      string
	ShowLogs       bool
	Port           string
}

// Load reads configuration from environment (with optional .env) and validates required fields.
func Load() Config {
	// On Unix, including macOS, it returns the $HOME environment variable.
	// On Windows, it returns %USERPROFILE%.
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Error getting home directory: %v", err)
	}

	// look for $HOME/.config/argus/argus.conf
	configPath := filepath.Join(homeDir, ".config", "argus", "argus.conf")

	// Load the config
	err = godotenv.Load(configPath)

	if err != nil {
		// Tell the user we cannot load the env
		log.Printf("Error Loading config file from %s. Falling back to .env instead", configPath)

		// Load environment variables from the .env file if present.
		if err := godotenv.Load(); err != nil {
			log.Println("Note: No .env file found. Falling back to system environment variables.")
		}
	}

	cfg := Config{
		Nick:           os.Getenv("TWITCH_NICK"),
		OAuthToken:     os.Getenv("TWITCH_TOKEN"),
		Channel:        os.Getenv("TWITCH_CHANNEL"),
		ChannelID:      os.Getenv("TWITCH_CHANNEL_ID"),
		ClientID:       os.Getenv("TWITCH_CLIENT_ID"),
		AppAccessToken: os.Getenv("TWITCH_APP_ACCESS_TOKEN"),
		ShowLogs:       os.Getenv("SHOW_LOGS") == "true",
		Port:           os.Getenv("PORT"),
	}

	var missingVars []string
	if cfg.Nick == "" {
		missingVars = append(missingVars, "TWITCH_NICK")
	}
	if cfg.OAuthToken == "" {
		missingVars = append(missingVars, "TWITCH_TOKEN")
	}
	if cfg.Channel == "" {
		missingVars = append(missingVars, "TWITCH_CHANNEL")
	}
	if cfg.ChannelID == "" {
		missingVars = append(missingVars, "TWITCH_CHANNEL_ID")
	}
	if cfg.ClientID == "" {
		missingVars = append(missingVars, "TWITCH_CLIENT_ID")
	}
	if cfg.AppAccessToken == "" {
		missingVars = append(missingVars, "TWITCH_APP_ACCESS_TOKEN")
	}
	if cfg.Port == "" {
		missingVars = append(missingVars, "PORT")
	}

	if len(missingVars) > 0 {
		log.Fatalf("Please set the following environment variables in your .env file: %s", strings.Join(missingVars, ", "))
	}

	return cfg
}
