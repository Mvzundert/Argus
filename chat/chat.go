package chat

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strings"

	"golang.org/x/term"

	"twitch-go/colors"
	"twitch-go/config"
)

// --- Configuration ---
// Your Twitch username or a guest account. "justinfan123" is a valid guest.
var NICK string

// Your Twitch OAuth token. Get one from the Twitch Dev Console with required scopes.
var OAUTH_TOKEN string

// Your Twitch Client ID.
var CLIENT_ID string

// The channel to connect to, including the # prefix.
var CHANNEL string

// A regular expression to parse the username and message from an IRC PRIVMSG.
var chatMessageRegex = regexp.MustCompile(`^:(\w+)!.*?PRIVMSG #\w+ :(.+)$`)

// A regular expression to extract IRC tags.
var ircTagRegex = regexp.MustCompile(`^@([^ ]+) `)

// A regular expression to strip ANSI codes.
var ansiStripRegex = regexp.MustCompile("[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?(?:[a-zA-Z\\d]+(?:;[a-zA-Z\\d]*)*)?[a-zA-Z])|(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?(?:[a-zA-Z\\d]+(?:;[a-zA-Z\\d]*)*)?[a-zA-Z\u0080-\u009F])")

// --- IRC Chat Configuration ---
const (
	IRC_SERVER = "irc.chat.twitch.tv"
	IRC_PORT   = "6667"
)

func Connect(cfg config.Config) {
	conn, err := net.Dial("tcp", IRC_SERVER+":"+IRC_PORT)
	if err != nil {
		log.Fatalf("Error connecting to Twitch IRC: %v", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)

	fmt.Println("\n-------------------- Twitch Chat --------------------")
	// Request IRCv3 tags capability to get user badges.
	fmt.Fprintf(conn, "CAP REQ :twitch.tv/tags\r\n")
	// The IRC connection requires the `oauth:` prefix.
	fmt.Fprintf(conn, "PASS oauth:%s\r\n", cfg.OAuthToken)
	fmt.Fprintf(conn, "NICK %s\r\n", cfg.Nick)
	fmt.Fprintf(conn, "JOIN %s\r\n", cfg.Channel)
	log.Printf("Joined IRC channel %s", cfg.Channel)

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

				// Get the terminal width and wrap the message
				width, _, err := term.GetSize(int(os.Stdout.Fd()))
				if err == nil && width > 0 {
					prefix := fmt.Sprintf(" [CHAT] %s[%s]%s: ", color, username, colors.ColorReset)
					prefixLen := len(stripAnsiCodes(prefix))
					wrappedMessage := wrapMessage(message, width-prefixLen, prefixLen)
					fmt.Printf("%s\n", prefix+wrappedMessage)
				} else {
					fmt.Printf(" [CHAT] %s[%s]%s: %s\n", color, username, colors.ColorReset, message)
				}
			} else {
				// Fallback for messages without tags
				matches := chatMessageRegex.FindStringSubmatch(line)
				if len(matches) == 3 {
					username = matches[1]
					fmt.Printf(" [CHAT] %s[%s]%s: %s\n", colors.ColorTwitchPurple, username, colors.ColorReset, message)
				}
			}
		}
	}
}

func wrapMessage(message string, width int, prefixLen int) string {
	var builder strings.Builder
	words := strings.Fields(message)
	if len(words) == 0 {
		return ""
	}

	currentLineLen := 0
	for i, word := range words {
		if currentLineLen+len(word)+1 > width {
			builder.WriteString("\n" + strings.Repeat(" ", prefixLen))
			currentLineLen = 0
		}
		builder.WriteString(word)
		if i < len(words)-1 {
			builder.WriteString(" ")
			currentLineLen += len(word) + 1
		} else {
			currentLineLen += len(word)
		}
	}
	return builder.String()
}

func stripAnsiCodes(str string) string {
	return ansiStripRegex.ReplaceAllString(str, "")
}

func getColorByRole(badges string) string {
	if strings.Contains(badges, "moderator") || strings.Contains(badges, "broadcaster") {
		return colors.ColorRed
	}
	return colors.ColorTwitchPurple
}

func parseTags(tagString string) map[string]string {
	tags := make(map[string]string)
	pairs := strings.SplitSeq(tagString, ";")
	for pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			tags[kv[0]] = kv[1]
		}
	}
	return tags
}
