package eventsub

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/websocket"

	"twitch-go/internal/colors"
	"twitch-go/internal/config"
)

const (
	eventSubWSURL = "wss://eventsub.wss.twitch.tv/ws"
	apiURL        = "https://api.twitch.tv/helix/eventsub/subscriptions"
)

// Run connects to Twitch EventSub over WebSocket and handles events.
func Run(cfg config.Config) {
	u := eventSubWSURL
	if cfg.ShowLogs {
		log.Printf("Connecting to EventSub at %s", u)
	}

	conn, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		log.Fatal("WebSocket connection error:", err)
	}
	defer conn.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				if cfg.ShowLogs {
					log.Println("Read error:", err)
				}
				return
			}
			var msg map[string]interface{}
			if err := json.Unmarshal(message, &msg); err != nil {
				if cfg.ShowLogs {
					log.Println("JSON unmarshal error:", err)
				}
				continue
			}

			metadata, ok := msg["metadata"].(map[string]any)
			if !ok {
				if cfg.ShowLogs {
					log.Println("metadata not found in message")
				}
				continue
			}

			messageType, ok := metadata["message_type"].(string)
			if !ok {
				if cfg.ShowLogs {
					log.Println("message_type not found in metadata")
				}
				continue
			}

			switch messageType {
			case "session_welcome":
				sessionID := msg["payload"].(map[string]any)["session"].(map[string]any)["id"].(string)
				if cfg.ShowLogs {
					log.Println("Received session welcome. Session ID:", sessionID)
				}
				subscribeToEvents(cfg, sessionID)
			case "session_keepalive":
				if cfg.ShowLogs {
					log.Println("Received keepalive message.")
				}
			case "notification":
				handleEventSubNotification(cfg, msg)
			case "revocation":
				if cfg.ShowLogs {
					log.Println("Received revocation. Session revoked.")
				}
				return
			case "session_reconnect":
				if cfg.ShowLogs {
					log.Println("Received reconnect message. Disconnecting and reconnecting...")
				}
				return
			default:
				if cfg.ShowLogs {
					log.Printf("Received unhandled message type: %s", messageType)
				}
			}
		}
	}()

	<-done
	if cfg.ShowLogs {
		log.Println("EventSub connection closed.")
	}
}

func subscribeToEvents(cfg config.Config, sessionID string) {
	subscriptionTypes := []string{
		"channel.subscribe",
		"channel.cheer",
		"channel.channel_points_custom_reward_redemption.add",
	}

	for _, eventType := range subscriptionTypes {
		data := map[string]any{
			"type":      eventType,
			"version":   "1",
			"condition": map[string]string{"broadcaster_user_id": cfg.ChannelID},
			"transport": map[string]string{"method": "websocket", "session_id": sessionID},
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			if cfg.ShowLogs {
				log.Printf("Error marshalling JSON for %s: %v", eventType, err)
			}
			continue
		}

		client := &http.Client{}
		req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
		if err != nil {
			if cfg.ShowLogs {
				log.Printf("Error creating request for %s: %v", eventType, err)
			}
			continue
		}

		req.Header.Add("Client-ID", cfg.ClientID)
		req.Header.Add("Authorization", "Bearer "+cfg.OAuthToken)
		req.Header.Add("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			if cfg.ShowLogs {
				log.Printf("Error making request for %s: %v", eventType, err)
			}
			continue
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			if cfg.ShowLogs {
				log.Printf("Error reading response body for %s: %v", eventType, err)
			}
			continue
		}

		if resp.StatusCode == http.StatusAccepted || resp.StatusCode == http.StatusOK {
			if cfg.ShowLogs {
				log.Printf("Successfully subscribed to %s", eventType)
			}
		} else {
			if cfg.ShowLogs {
				log.Printf("Failed to subscribe to %s. Status: %s, Body: %s", eventType, resp.Status, string(bodyBytes))
			}
		}
	}

	if cfg.ShowLogs {
		fmt.Println("----------------- Activity Feed -----------------")
		fmt.Println("Application is now ready to receive events.")
		fmt.Println("-------------------------------------------------")
	}
}

func handleEventSubNotification(cfg config.Config, msg map[string]any) {
	payload, ok := msg["payload"].(map[string]any)
	if !ok {
		return
	}
	event, ok := payload["event"].(map[string]any)
	if !ok {
		return
	}
	eventType, ok := payload["subscription"].(map[string]any)["type"].(string)
	if !ok {
		return
	}

	switch eventType {
	case "channel.subscribe":
		username := event["user_name"].(string)
		fmt.Printf("%s [ACTIVITY] New Subscriber: %s!%s\n", colors.ColorWhite, username, colors.ColorReset)
	case "channel.cheer":
		username := event["user_name"].(string)
		bitsAmount := event["bits"].(float64)
		fmt.Printf("%s [ACTIVITY] %s cheered %d bits!%s\n", colors.ColorPurple, username, int(bitsAmount), colors.ColorReset)
	case "channel.channel_points_custom_reward_redemption.add":
		username := event["user_name"].(string)
		rewardTitle := event["reward"].(map[string]any)["title"].(string)
		rewardCost := event["reward"].(map[string]any)["cost"].(float64)
		fmt.Printf("%s [ACTIVITY] %s redeemed %d channel points for: %s%s\n", colors.ColorCyan, username, int(rewardCost), rewardTitle, colors.ColorReset)
	}
}
