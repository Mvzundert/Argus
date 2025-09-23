package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"argus/chat"
	"argus/config"
	"argus/events"
	"argus/web"
)

func main() {
	cfg := config.Load()

	// Create a channel to handle web server errors.
	serverErr := make(chan error, 1)

	// Start the web server in its own goroutine.
	// You must handle the error returned by web.StartServer.
	go func() {
		if err := web.StartServer(cfg); err != nil {
			serverErr <- err
		}
	}()

	// Use a channel to wait for a termination signal.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Run chat and Events concurrently.
	go chat.Connect(cfg)
	go events.Run(cfg)

	// Wait for a termination signal or a server error.
	select {
	case err := <-serverErr:
		log.Fatalf("Server failed to start: %v", err)
	case sig := <-sigs:
		fmt.Printf("\nReceived signal: %s. Terminating...\n", sig)
	}
}

