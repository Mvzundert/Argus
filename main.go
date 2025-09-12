package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"twitch-go/internal/chat"
	"twitch-go/internal/config"
	"twitch-go/internal/eventsub"
)

func main() {
	cfg := config.Load()

	// Use a channel to wait for a termination signal.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Run chat and EventSub concurrently.
	go chat.Connect(cfg)
	go eventsub.Run(cfg)

	// Wait for a termination signal to close the program.
	<-sigs
	fmt.Println("\nProgram terminated. Disconnecting...")
}