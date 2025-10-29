package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ledabeer/backend/internal/api"
	"ledabeer/backend/internal/api/grpc"
	"ledabeer/backend/internal/api/websocket"
	"ledabeer/backend/internal/auth"
	"ledabeer/backend/internal/conversations"
	"ledabeer/backend/internal/logging"
	"ledabeer/backend/internal/media"
	"ledabeer/backend/internal/messaging"
	"ledabeer/backend/internal/network"
	"ledabeer/backend/internal/user"
	"ledabeer/backend/internal/storage"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Parse command line arguments
	var (
		isBootstrap   = flag.Bool("bootstrap", false, "Run as bootstrap node")
		listenAddr    = flag.String("listen", "/ip4/0.0.0.0/tcp/4001", "Listen address")
		bootstrapHost = flag.String("bootstrap-host", "", "Bootstrap host for peer discovery")
		bootstrapPort = flag.String("bootstrap-port", "4001", "Bootstrap port for peer discovery")
	)
	flag.Parse()

	// Initialize logging
	logConfig, err := logging.ConfigFromEnv()
	if err != nil {
		fmt.Printf("Failed to load logging config: %v\n", err)
		os.Exit(1)
	}
	logger := logging.NewLogger(logConfig)
	logging.SetDefault(logger)

	if *isBootstrap {
		logger.Info("Starting Ledabeer backend as bootstrap node")
	} else {
		logger.Info("Starting Ledabeer backend as peer node")
	}

	// Create libp2p host
	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs:   []string{*listenAddr},
		BootstrapHost: *bootstrapHost,
		BootstrapPort: *bootstrapPort,
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
	// Create real gRPC services
	msgService := grpc.NewMessageService(msgHandler, groupManager)
	mediaService := grpc.NewMediaService(mediaHandler)
	peerService := grpc.NewPeerService(host)
	groupService := grpc.NewGroupService(groupManager)

	// Create authentication and conversation services
	auth := auth.NewAuthenticator(nil) // Use memory repository for now
	userManager := user.NewUserManager()
	conversationService := conversations.NewConversationService(userManager)
	
	// Create WebSocket server
	wsServer := websocket.NewServer(auth, conversationService)

	// Create API gateway with real services
	gateway := api.NewGateway(&api.Config{
		GRPCPort: 50051,
		HTTPPort: 8080,
	}, host, msgHandler, msgService, mediaService, peerService, groupService, wsServer, auth)

	// Logging is already integrated in the services

	// Start gateway
	if err := gateway.Start(); err != nil {
		logger.Error("Failed to start gateway", logging.String("error", err.Error()))
		os.Exit(1)
	}

	// Wait a moment for servers to start
	time.Sleep(100 * time.Millisecond)

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
