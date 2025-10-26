package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"ledabeer/backend/internal/api"
	"ledabeer/backend/internal/api/grpc"
	"ledabeer/backend/internal/api/websocket"
	"ledabeer/backend/internal/auth"
	"ledabeer/backend/internal/calls"
	"ledabeer/backend/internal/logging"
	"ledabeer/backend/internal/media"
	"ledabeer/backend/internal/messaging"
	"ledabeer/backend/internal/network"
	"ledabeer/backend/internal/storage"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize logging
	logConfig, err := logging.ConfigFromEnv()
	if err != nil {
		fmt.Printf("Failed to load logging config: %v\n", err)
		os.Exit(1)
	}
	logger := logging.NewLogger(logConfig)
	logging.SetDefault(logger)

	logger.Info("Starting Ledabeer backend")

	// Create libp2p host
	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/0.0.0.0/tcp/4001"},
	})
	if err != nil {
		logger.Error("Failed to create host", logging.String("error", err.Error()))
		os.Exit(1)
	}
	defer host.Close()

	// Initialize IPFS
	ipfsNode, err := storage.NewIPFSNode(ctx, &storage.IPFSConfig{})
	if err != nil {
		logger.Error("Failed to initialize IPFS", logging.String("error", err.Error()))
		os.Exit(1)
	}
	defer ipfsNode.Close()

	// Initialize real handlers
	msgHandler := messaging.NewMessageHandler(host)
	groupManager, err := messaging.NewGroupManager(ctx, host)
	if err != nil {
		logger.Error("Failed to create group manager", logging.String("error", err.Error()))
		os.Exit(1)
	}
	mediaHandler := media.NewMediaHandler(ipfsNode)
	callManager := calls.NewCallManager(host)

	// Create real gRPC services
	msgService := grpc.NewMessageService(msgHandler, groupManager)
	mediaService := grpc.NewMediaService(mediaHandler)
	callService := grpc.NewCallService(callManager)

	// Create WebSocket server
	auth := auth.NewAuthenticator()
	wsServer := websocket.NewServer(auth)

	// Create API gateway with real services
	gateway := api.NewGateway(&api.Config{
		GRPCPort: 50051,
		HTTPPort: 8080,
	}, msgService, mediaService, callService, wsServer)

	// Logging is already integrated in the services

	// Start gateway
	if err := gateway.Start(); err != nil {
		logger.Error("Failed to start gateway", logging.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("Backend started successfully",
		logging.String("peer_id", host.ID().String()),
		logging.String("grpc_addr", gateway.GRPCAddr()),
		logging.String("http_addr", gateway.HTTPAddr()),
	)

	// Print host information
	fmt.Printf("Node started with ID: %s\n", host.ID())
	fmt.Printf("Listening on addresses:\n")
	for _, addr := range host.Addrs() {
		fmt.Printf("  %s/p2p/%s\n", addr, host.ID())
	}
	fmt.Printf("gRPC server: %s\n", gateway.GRPCAddr())
	fmt.Printf("HTTP server: %s\n", gateway.HTTPAddr())

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	logger.Info("Shutting down...")
	gateway.Stop()
	fmt.Println("\nShutting down...")
}
