package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"ledabeer/backend/internal/network"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a new host
	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/0.0.0.0/tcp/4001"},
	})
	if err != nil {
		log.Fatal("Failed to create host:", err)
	}
	defer host.Close()

	// Print host information
	fmt.Printf("Node started with ID: %s\n", host.ID())
	fmt.Printf("Listening on addresses:\n")
	for _, addr := range host.Addrs() {
		fmt.Printf("  %s/p2p/%s\n", addr, host.ID())
	}

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("\nShutting down...")
}
