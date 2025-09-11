package main

import (
	"bufio"
	"encoding/json"
	"os/signal"

	"fmt"
	"log"
	"net"
	"net/url"
	"os"

	//"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	//	"github.com/gorilla/websocket"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

// --- Configuration ---
// This application now uses a .env file to load configuration.
// Please create a file named .env in the same directory as this script
// and fill it with your credentials. See the example.env file provided.

var (
	// Your Twitch username or a guest account. "justinfan123" is a valid guest.
	NICK string
	// Your Twitch OAuth token. Get one from https://dev.twitch.tv/console.
	OAUTH_TOKEN string
	// The channel to connect to, including the # prefix.
	CHANNEL string
)

// Configuration for the activity feed.
const (
	PUBSUB_URL           = "wss://pubsub-edge.twitch.tv"
	PUBSUB_TOPIC_SUB     = "channel-subscribe-events-v1.%s"
	PUBSUB_TOPIC_BITS    = "channel-bits-events-v2.%s"
	PUBSUB_TOPIC_POINTS  = "channel-points-channel-v1.%s"
	PUBSUB_PING_INTERVAL = 4 * time.Minute
)

// Configuration for irc chat
const (
	SERVER = "irc.chat.twitch.tv"
	PORT   = "6667"
)

// A regular expression to parse the username and message from an IRC PRIVMSG.
var chatMessageRegex = regexp.MustCompile(`^:(\w+)!.*?PRIVMSG #\w+ :(.+)$`)

// A regular expression to extract user ID from the response.
var userIDRegex = regexp.MustCompile(`^@.*user-id=(\d+).*$`)

// Channel IDs are for the Pub Sub topics, you need the ID and not the name.
var CHANNEL_ID string

func main() {
	// Load environment variables from the .env file.
	err := godotenv.Load()
	if err != nil {
		log.Println("Note: No .env file found.")
	}

	// Read values from the environment.
	NICK = os.Getenv("TWITCH_NICK")
	OAUTH_TOKEN = os.Getenv("TWITCH_TOKEN")
	CHANNEL = os.Getenv("TWITCH_CHANNEL")
	CHANNEL_ID = os.Getenv("TWITCH_CHANNEL_ID")

	// Check for required configuration.
	if NICK == "" || OAUTH_TOKEN == "" || CHANNEL == "" || CHANNEL_ID == "" {
		log.Println("Please set TWITCH_NICK, TWITCH_TOKEN, and TWITCH_CHANNEL in your .env file or as environment variables.")
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go connectToIRC()
	go connectToPubSub()

	<-sigs
	fmt.Println("\nProgram Terminating, Disconnecting....")
}

func connectToIRC() {
	// Establish a TCP connection to the IRC server.
	conn, err := net.Dial("tcp", SERVER+":"+PORT)
	if err != nil {
		log.Fatalf("Error connecting to Twitch: %v", err)
	}
	defer conn.Close()

	// Use a bufio.Reader for efficient line-by-line reading.
	reader := bufio.NewReader(conn)

	log.Printf("Connected to Twitch IRC server: %s", conn.RemoteAddr().String())
	fmt.Println("----------------- Twitch Chat -----------------")
	// Send authentication details to the server.
	fmt.Fprintf(conn, "PASS %s\r\n", OAUTH_TOKEN)
	fmt.Fprintf(conn, "NICK %s\r\n", NICK)
	fmt.Fprintf(conn, "JOIN %s\r\n", CHANNEL)

	// log.Printf("Joining channel %s...", CHANNEL)

	for {
		// Read a line from the connection. A line is terminated by \r\n.
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Connection lost or closed: %v", err)
			return
		}
		line = strings.TrimSpace(line)

		// Handle PING/PONG to keep the connection alive.
		if strings.HasPrefix(line, "PING") {
			fmt.Fprintf(conn, "PONG :tmi.twitch.tv\r\n")
			// log.Println("PING received, PONG sent.")
			continue
		}

		// Parse the line for a chat message.
		matches := chatMessageRegex.FindStringSubmatch(line)
		if len(matches) == 3 {
			// matches[1] is the username, matches[2] is the message.
			username := matches[1]
			message := matches[2]
			fmt.Printf("[%s]: %s\n", username, message)
		}
	}
}

func connectToPubSub() {
	url := url.URL{Scheme: "wss", Host: "pubsub-edge.twitch.tv", Path: ""}
	log.Printf("Connecting to PubSub at %s", url.String())

	conn, _, err := websocket.DefaultDialer.Dial(url.String(), nil)

	if err != nil {
		log.Fatal("Websocket connection error:", err)
	}
	defer conn.Close()

	done := make(chan struct{})

	// Handle the incoming messages.
	go func() {
		defer conn.Close()
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("Read error:", err)
				return
			}
			var msg map[string]interface{}
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Println("JSON Unmarshal error: ", err)
				continue
			}
			handlePubSubMessage(msg)
		}
	}()

	topics := []string{
		fmt.Sprintf(PUBSUB_TOPIC_SUB, CHANNEL_ID),
		fmt.Sprintf(PUBSUB_TOPIC_BITS, CHANNEL_ID),
		fmt.Sprintf(PUBSUB_TOPIC_POINTS, CHANNEL_ID),
	}
	subscribe(conn, topics, OAUTH_TOKEN)

	// Need to keep the connection alive
	ticker := time.NewTicker(PUBSUB_PING_INTERVAL)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			ping(conn)
		}
	}
}

func subscribe(conn *websocket.Conn, topics []string, token string) {
	fmt.Println("----------------- Twitch Activity -----------------")
	request := map[string]interface{}{
		"type":  "LISTEN",
		"NONCE": fmt.Sprintf("%d", time.Now().UnixNano()),
		"data": map[string]interface{}{
			"topics":     topics,
			"auth_token": token,
		},
	}
	if err := conn.WriteJSON(request); err != nil {
		log.Println(" Subscribe failed: ", err)
	} else {
		log.Println(" Subscribe to topics: ")
	}
}

func ping(conn *websocket.Conn) {
	request := map[string]string{"type": "PING"}
	if err := conn.WriteJSON(request); err != nil {
		log.Println(" Ping failed: ", err)
	}
}

func handlePubSubMessage(msg map[string]interface{}) {
	if msg["type"] == "MESSAGE" {
		data, ok := msg["data"].(map[string]interface{})
		if !ok {
			return
		}

		messageData, ok := data["message"].(string)
		if !ok {
			return
		}

		var pubsubEvent map[string]interface{}
		if err := json.Unmarshal([]byte(messageData), &pubsubEvent); err != nil {
			log.Println(" Error unmarshaling PubSub event message: ", err)
			return
		}

		switch pubsubEvent["type"] {
		case "sub_message":
			// Handles all subscriber events
			subscriptionData := pubsubEvent["data"].(map[string]interface{})
			username := subscriptionData["username"].(string)
			fmt.Printf("[FEED]: New Subscriber: %s!\n", username)
		case "bits_event":
			// Handles all bit events
			bitsData := pubsubEvent["data"].(map[string]interface{})
			username := bitsData["username"].(string)
			bitsAmount := bitsData["bits_used"].(float64)
			fmt.Printf("[FEED]: %s cheered %d bits!!\n", username, int(bitsAmount))
		case "reward_redeemded":
			// Handles all point redeem events
			redemptionData := pubsubEvent["data"].(map[string]interface{})["redemption"].(map[string]interface{})
			user := redemptionData["user"].(map[string]interface{})
			reward := redemptionData["reward"].(map[string]interface{})
			username := user["login"].(string)
			rewardTitle := reward["title"].(string)
			fmt.Printf("[FEED] %s redeemed channel points for: %s\n", username, rewardTitle)
		}
	}
}
