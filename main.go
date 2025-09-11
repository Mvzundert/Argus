package main

import (
	"bufio"
	"os/signal"

	// "encoding/json"
	"fmt"
	"log"
	"net"

	// "net/url"
	"os"
	// "os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	//	"github.com/gorilla/websocket"
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
