package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

// --- Configuration ---
// Your Twitch username or a guest account. "justinfan123" is a valid guest.
var NICK string

// Your Twitch OAuth token. Get one from the Twitch Dev Console with required scopes.
var OAUTH_TOKEN string

// Your Twitch Client ID.
var CLIENT_ID string

// Your Twitch App Access Token. Get one for your application to make API calls.
var APP_ACCESS_TOKEN string

// The channel to connect to, including the # prefix.
var CHANNEL string

// Twitch EventSub configuration.
const (
	EVENTSUB_URL = "wss://eventsub.wss.twitch.tv/ws"
	API_URL      = "https://api.twitch.tv/helix/eventsub/subscriptions"
)

// --- IRC Chat Configuration ---
const (
	IRC_SERVER = "irc.chat.twitch.tv"
	IRC_PORT   = "6667"
)

// ANSI escape codes for coloring terminal output.
const (
	ColorReset        = "\033[0m"
	ColorRed          = "\033[31m"
	ColorTwitchPurple = "\033[38;2;145;70;255m" // Custom color for Twitch Purple
	ColorWhite        = "\033[97m"
	ColorPurple       = "\033[35m"
	ColorCyan         = "\033[36m"
)

// A regular expression to parse the username and message from an IRC PRIVMSG.
var chatMessageRegex = regexp.MustCompile(`^:(\w+)!.*?PRIVMSG #\w+ :(.+)$`)

// A regular expression to extract IRC tags.
var ircTagRegex = regexp.MustCompile(`^@([^ ]+) `)

// Channel IDs for PubSub topics. You need the user ID, not the name.
var CHANNEL_ID string
var SHOW_LOGS bool

func main() {
	// Load environment variables from the .env file.
	err := godotenv.Load()
	if err != nil {
		log.Println("Note: No .env file found. Falling back to system environment variables.")
	}

	// Read values from the environment.
	NICK = os.Getenv("TWITCH_NICK")
	OAUTH_TOKEN = os.Getenv("TWITCH_TOKEN")
	CHANNEL = os.Getenv("TWITCH_CHANNEL")
	CHANNEL_ID = os.Getenv("TWITCH_CHANNEL_ID")
	CLIENT_ID = os.Getenv("TWITCH_CLIENT_ID")
	APP_ACCESS_TOKEN = os.Getenv("TWITCH_APP_ACCESS_TOKEN")
	SHOW_LOGS = os.Getenv("SHOW_LOGS") == "true"

	var missingVars []string
	if NICK == "" {
		missingVars = append(missingVars, "TWITCH_NICK")
	}
	if OAUTH_TOKEN == "" {
		missingVars = append(missingVars, "TWITCH_TOKEN")
	}
	if CHANNEL == "" {
		missingVars = append(missingVars, "TWITCH_CHANNEL")
	}
	if CHANNEL_ID == "" {
		missingVars = append(missingVars, "TWITCH_CHANNEL_ID")
	}
	if CLIENT_ID == "" {
		missingVars = append(missingVars, "TWITCH_CLIENT_ID")
	}
	if APP_ACCESS_TOKEN == "" {
		missingVars = append(missingVars, "TWITCH_APP_ACCESS_TOKEN")
	}

	if len(missingVars) > 0 {
		log.Fatalf("Please set the following environment variables in your .env file: %s", strings.Join(missingVars, ", "))
	}

	// Use a channel to wait for a termination signal.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// We will run two separate goroutines for chat and PubSub.
	go connectToIRC()
	go connectToEventSub()

	// Wait for a termination signal to close the program.
	<-sigs
	fmt.Println("\nProgram terminated. Disconnecting...")
}

func connectToIRC() {
	conn, err := net.Dial("tcp", IRC_SERVER+":"+IRC_PORT)
	if err != nil {
		log.Fatalf("Error connecting to Twitch IRC: %v", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)

	if SHOW_LOGS {
		fmt.Println("-------------------- Twitch Chat --------------------")
		// Request IRCv3 tags capability to get user badges.
		fmt.Fprintf(conn, "CAP REQ :twitch.tv/tags\r\n")
		// The IRC connection requires the `oauth:` prefix.
		fmt.Fprintf(conn, "PASS oauth:%s\r\n", OAUTH_TOKEN)
		fmt.Fprintf(conn, "NICK %s\r\n", NICK)
		fmt.Fprintf(conn, "JOIN %s\r\n", CHANNEL)
		log.Printf("Joined IRC channel %s", CHANNEL)
		fmt.Println("-------------------------------------------------")
	} else {
		log.Printf("CLI active for channel %s", CHANNEL)
	}

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("IRC Connection lost or closed: %v", err)
			return
		}
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "PING") {
			fmt.Fprintf(conn, "PONG :tmi.twitch.tv\r\n")
		}

		parts := strings.SplitN(line, "PRIVMSG", 2)
		if len(parts) == 2 {
			tagString := ircTagRegex.FindStringSubmatch(parts[0])
			username := ""
			message := strings.TrimSpace(parts[1][strings.Index(parts[1], ":")+1:])

			if tagString != nil {
				tags := parseTags(tagString[1])
				username = tags["display-name"]
				if username == "" {
					username = tags["login"]
				}
				color := getColorByRole(tags["badges"])
				fmt.Printf(" [CHAT] %s[%s]%s: %s\n", color, username, ColorReset, message)
			} else {
				// Fallback for messages without tags
				matches := chatMessageRegex.FindStringSubmatch(line)
				if len(matches) == 3 {
					username = matches[1]
					fmt.Printf(" [CHAT] %s[%s]%s: %s\n", ColorTwitchPurple, username, ColorReset, message)
				}
			}
		}
	}
}

func getColorByRole(badges string) string {
	if strings.Contains(badges, "moderator") || strings.Contains(badges, "broadcaster") {
		return ColorRed
	}
	return ColorTwitchPurple
}

func parseTags(tagString string) map[string]string {
	tags := make(map[string]string)
	pairs := strings.Split(tagString, ";")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			tags[kv[0]] = kv[1]
		}
	}
	return tags
}

func connectToEventSub() {
	u := url.URL{Scheme: "wss", Host: "eventsub.wss.twitch.tv", Path: "/ws"}
	if SHOW_LOGS {
		log.Printf("Connecting to EventSub at %s", u.String())
	}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
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
				if SHOW_LOGS {
					log.Println("Read error:", err)
				}
				return
			}
			var msg map[string]interface{}
			if err := json.Unmarshal(message, &msg); err != nil {
				if SHOW_LOGS {
					log.Println("JSON unmarshal error:", err)
				}
				continue
			}

			metadata, ok := msg["metadata"].(map[string]any)
			if !ok {
				if SHOW_LOGS {
					log.Println("metadata not found in message")
				}
				continue
			}

			messageType, ok := metadata["message_type"].(string)
			if !ok {
				if SHOW_LOGS {
					log.Println("message_type not found in metadata")
				}
				continue
			}

			switch messageType {
			case "session_welcome":
				sessionID := msg["payload"].(map[string]any)["session"].(map[string]any)["id"].(string)
				if SHOW_LOGS {
					log.Println("Received session welcome. Session ID:", sessionID)
				}
				subscribeToEvents(sessionID)
			case "session_keepalive":
				if SHOW_LOGS {
					log.Println("Received keepalive message.")
				}
			case "notification":
				handleEventSubNotification(msg)
			case "revocation":
				if SHOW_LOGS {
					log.Println("Received revocation. Session revoked.")
				}
				return
			case "session_reconnect":
				if SHOW_LOGS {
					log.Println("Received reconnect message. Disconnecting and reconnecting...")
				}
				return
			default:
				if SHOW_LOGS {
					log.Printf("Received unhandled message type: %s", messageType)
				}
			}
		}
	}()

	<-done
	if SHOW_LOGS {
		log.Println("EventSub connection closed.")
	}
}

func subscribeToEvents(sessionID string) {
	subscriptionTypes := []string{
		"channel.subscribe",
		"channel.cheer",
		"channel.channel_points_custom_reward_redemption.add",
	}

	for _, eventType := range subscriptionTypes {
		data := map[string]any{
			"type":      eventType,
			"version":   "1",
			"condition": map[string]string{"broadcaster_user_id": CHANNEL_ID},
			"transport": map[string]string{"method": "websocket", "session_id": sessionID},
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			if SHOW_LOGS {
				log.Printf("Error marshalling JSON for %s: %v", eventType, err)
			}
			continue
		}

		client := &http.Client{}
		req, err := http.NewRequest("POST", API_URL, bytes.NewBuffer(jsonData))
		if err != nil {
			if SHOW_LOGS {
				log.Printf("Error creating request for %s: %v", eventType, err)
			}
			continue
		}

		req.Header.Add("Client-ID", CLIENT_ID)
		req.Header.Add("Authorization", "Bearer "+OAUTH_TOKEN)
		req.Header.Add("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			if SHOW_LOGS {
				log.Printf("Error making request for %s: %v", eventType, err)
			}
			continue
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			if SHOW_LOGS {
				log.Printf("Error reading response body for %s: %v", eventType, err)
			}
			continue
		}

		if resp.StatusCode == http.StatusAccepted || resp.StatusCode == http.StatusOK {
			if SHOW_LOGS {
				log.Printf("Successfully subscribed to %s", eventType)
			}
		} else {
			if SHOW_LOGS {
				log.Printf("Failed to subscribe to %s. Status: %s, Body: %s", eventType, resp.Status, string(bodyBytes))
			}
		}
	}

	if SHOW_LOGS {
		fmt.Println("----------------- Activity Feed -----------------")
		fmt.Println("Application is now ready to receive events.")
		fmt.Println("-------------------------------------------------")
	}
}

func handleEventSubNotification(msg map[string]interface{}) {
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
		fmt.Printf("%s [ACTIVITY] New Subscriber: %s!%s\n", ColorWhite, username, ColorReset)
	case "channel.cheer":
		username := event["user_name"].(string)
		bitsAmount := event["bits"].(float64)
		fmt.Printf("%s [ACTIVITY] %s cheered %d bits!%s\n", ColorPurple, username, int(bitsAmount), ColorReset)
	case "channel.channel_points_custom_reward_redemption.add":
		username := event["user_name"].(string)
		rewardTitle := event["reward"].(map[string]any)["title"].(string)
		rewardCost := event["reward"].(map[string]any)["cost"].(float64)
		fmt.Printf("%s [ACTIVITY] %s redeemed %d channel points for: %s%s\n", ColorCyan, username, int(rewardCost), rewardTitle, ColorReset)
	}
}
