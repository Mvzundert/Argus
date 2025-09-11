package chat

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"regexp"
	"strings"

	"twitch-go/internal/colors"
	"twitch-go/internal/config"
)

const (
	ircServer = "irc.chat.twitch.tv"
	ircPort   = "6667"
)

var (
	chatMessageRegex = regexp.MustCompile(`^:(\w+)!.*?PRIVMSG #\w+ :(.+)$`)
	ircTagRegex      = regexp.MustCompile(`^@([^ ]+) `)
)

// Connect establishes an IRC connection to Twitch chat and prints messages.
func Connect(cfg config.Config) {
	conn, err := net.Dial("tcp", ircServer+":"+ircPort)
	if err != nil {
		log.Fatalf("Error connecting to Twitch IRC: %v", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// Request IRCv3 tags capability to get user badges.
	fmt.Fprintf(conn, "CAP REQ :twitch.tv/tags\r\n")
	// The IRC connection requires the `oauth:` prefix.
	fmt.Fprintf(conn, "PASS oauth:%s\r\n", cfg.OAuthToken)
	fmt.Fprintf(conn, "NICK %s\r\n", cfg.Nick)
	fmt.Fprintf(conn, "JOIN %s\r\n", cfg.Channel)

	if cfg.ShowLogs {
		fmt.Println("-------------------- Twitch Chat --------------------")
		log.Printf("Joined IRC channel %s", cfg.Channel)
		fmt.Println("-------------------------------------------------")
	} else {
		log.Printf("CLI active for channel %s", cfg.Channel)
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
				fmt.Printf(" [CHAT] %s[%s]%s: %s\n", color, username, colors.ColorReset, message)
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

func getColorByRole(badges string) string {
	if strings.Contains(badges, "moderator") || strings.Contains(badges, "broadcaster") {
		return colors.ColorRed
	}
	return colors.ColorTwitchPurple
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
