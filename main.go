package main

import (
	"fmt"
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

	// Start the web server in its own goroutine.
	go web.StartServer()

	// Use a channel to wait for a termination signal.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Run chat and Events concurrently.
	go chat.Connect(cfg)
	go events.Run(cfg)

	// Wait for a termination signal to close the program.
	<-sigs
	fmt.Println("\nProgram terminated. Disconnecting...")
}
