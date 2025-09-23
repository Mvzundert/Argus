package services

import (
	"argus/dependencies"
	"argus/music"
	"fmt"
	"log"
	"runtime"
	"strconv"
	"strings"
)

// NowPlayingData represents the data to be returned by the API.
type NowPlayingData struct {
	IsPlaying  bool   `json:"is_playing"`
	ProgressMs int64  `json:"progress_ms,omitempty"`
	Item       *Track `json:"item,omitempty"`
}

// Track represents the now playing song information.
type Track struct {
	Name       string   `json:"name"`
	Artists    []Artist `json:"artists"`
	DurationMs int64    `json:"duration_ms,omitempty"`
}

// Artist represents the artist's name.
type Artist struct {
	Name string `json:"name"`
}

// NowPlayingService is the INTERFACE that defines the contract.
type NowPlayingService interface {
	GetNowPlayingInfo() (NowPlayingData, error)
}

// nowPlayingServiceImpl is the CONCRETE STRUCT that implements the interface.
type nowPlayingServiceImpl struct{}

// NewNowPlayingService creates a new instance of the service.
func NewNowPlayingService() NowPlayingService {
	return &nowPlayingServiceImpl{}
}

// GetNowPlayingInfo retrieves the current track information.
func (s *nowPlayingServiceImpl) GetNowPlayingInfo() (NowPlayingData, error) {
	dep := dependencies.GetNowPlayingDependency()
	if dep == nil {
		return NowPlayingData{}, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	ok, errMessage := dep.Check()
	if !ok {
		return NowPlayingData{}, fmt.Errorf("dependency check failed: %s", errMessage)
	}

	var rawOutput string
	var err error
	var playerFound bool

	if runtime.GOOS == "linux" {
		playerPriority := []string{"spotify", "vlc"}
		for _, player := range playerPriority {
			args := []string{"--player=" + player, "metadata", "--format", "{{status}};{{artist}};{{title}};{{position}};{{mpris:length}}"}
			output, innerErr := music.GetNowPlayingInfo(dep.Command, args...)

			if innerErr == nil && strings.HasPrefix(output, "Playing") {
				rawOutput = output
				playerFound = true
				break
			}
		}
	} else if runtime.GOOS == "darwin" {
		args := []string{"current"}
		rawOutput, err = music.GetNowPlayingInfo(dep.Command, args...)
		if err != nil {
			return NowPlayingData{}, err
		}
		playerFound = true
	}

	if !playerFound {
		return NowPlayingData{IsPlaying: false}, nil
	}

	if rawOutput == "" {
		return NowPlayingData{IsPlaying: false}, nil
	}

	return parseRawOutput(rawOutput, runtime.GOOS)
}

func parseRawOutput(rawOutput, os string) (NowPlayingData, error) {
	if os == "linux" {
		parts := strings.Split(rawOutput, ";")

		if len(parts) < 5 {
			log.Println("Parsing failed, not enough parts.")
			return NowPlayingData{IsPlaying: false}, nil
		}

		position, err := parseTime(parts[3])
		if err != nil {
			log.Printf("Error parsing position: %v", err)
			position = 0
		}

		length, err := parseTime(parts[4])
		if err != nil {
			log.Printf("Error parsing length: %v", err)
			length = 0
		}

		if length == 9223372036854775807 {
			length = 0
		}

		return NowPlayingData{
			IsPlaying:  true,
			ProgressMs: position,
			Item: &Track{
				Name:       parts[2],
				Artists:    []Artist{{Name: parts[1]}},
				DurationMs: length,
			},
		}, nil
	}

	if os == "darwin" {
		parts := strings.SplitN(rawOutput, " by ", 2)

		if len(parts) < 2 {
			log.Println("Parsing failed, not enough parts.")
			return NowPlayingData{IsPlaying: false}, nil
		}

		return NowPlayingData{
			IsPlaying: true,
			Item: &Track{
				Name:    parts[0],
				Artists: []Artist{{Name: parts[1]}},
			},
		}, nil
	}

	return NowPlayingData{IsPlaying: false}, nil
}

func parseTime(s string) (int64, error) {
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return val / 1000, nil
}
