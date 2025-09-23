package web

import (
	"argus/services"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// mockNowPlayingService is a mock implementation of the services.NowPlayingService interface.
type mockNowPlayingService struct {
	data services.NowPlayingData
	err  error
}

// GetNowPlayingInfo is the mock method. It returns the predefined data and error.
func (m *mockNowPlayingService) GetNowPlayingInfo() (services.NowPlayingData, error) {
	return m.data, m.err
}

func TestIndexHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	IndexHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	body, err := io.ReadAll(rr.Body)
	if err != nil {
		t.Fatalf("could not read response body: %v", err)
	}
	bodyString := string(body)

	// Verify the essential HTML elements exist.
	if !strings.Contains(bodyString, "<title>Now Playing</title>") {
		t.Errorf("HTML body does not contain the expected title")
	}

	if !strings.Contains(bodyString, `<div id="spotify-widget"`) {
		t.Errorf("HTML body does not contain the spotify-widget element")
	}
}

func TestNowPlayingHandler(t *testing.T) {
	// Test Case 1: Successful Response with a playing song
	t.Run("successful response", func(t *testing.T) {
		mockData := services.NowPlayingData{
			IsPlaying: true,
			Item: &services.Track{
				Name: "Go Testing Anthem",
				Artists: []services.Artist{
					{Name: "Gopher"},
				},
				DurationMs: 1000,
			},
			ProgressMs: 500,
		}

		mockSvc := &mockNowPlayingService{data: mockData, err: nil}

		req, _ := http.NewRequest("GET", "/now-playing", nil)
		rr := httptest.NewRecorder()

		NowPlayingHandler(mockSvc, rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		var respData services.NowPlayingData
		if err := json.NewDecoder(rr.Body).Decode(&respData); err != nil {
			t.Fatalf("could not decode response body: %v", err)
		}

		if !respData.IsPlaying {
			t.Errorf("expected IsPlaying to be true, got false")
		}

		if respData.Item.Name != "Go Testing Anthem" {
			t.Errorf("expected song name %q, got %q", "Go Testing Anthem", respData.Item.Name)
		}
	})

	// Test Case 2: Service returns an error
	t.Run("service returns error", func(t *testing.T) {
		mockSvc := &mockNowPlayingService{data: services.NowPlayingData{}, err: fmt.Errorf("API error")}

		req, _ := http.NewRequest("GET", "/now-playing", nil)
		rr := httptest.NewRecorder()

		NowPlayingHandler(mockSvc, rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		var respData services.NowPlayingData
		if err := json.NewDecoder(rr.Body).Decode(&respData); err != nil {
			t.Fatalf("could not decode response body: %v", err)
		}

		if respData.IsPlaying {
			t.Error("expected IsPlaying to be false on error, got true")
		}
	})

	// Test Case 3: No song is playing
	t.Run("no song playing", func(t *testing.T) {
		mockData := services.NowPlayingData{IsPlaying: false}
		mockSvc := &mockNowPlayingService{data: mockData, err: nil}

		req, _ := http.NewRequest("GET", "/now-playing", nil)
		rr := httptest.NewRecorder()

		NowPlayingHandler(mockSvc, rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		var respData services.NowPlayingData
		if err := json.NewDecoder(rr.Body).Decode(&respData); err != nil {
			t.Fatalf("could not decode response body: %v", err)
		}

		if respData.IsPlaying {
			t.Error("expected IsPlaying to be false, got true")
		}
	})
}
