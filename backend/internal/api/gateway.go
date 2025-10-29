package api

import (
	"fmt"
	"net"
	"net/http"
	"sync"

	grpcapi "ledabeer/backend/internal/api/grpc"
	"ledabeer/backend/internal/api/websocket"

	pb "ledabeer/backend/pkg/proto"

	"github.com/libp2p/go-libp2p/core/host"
	"google.golang.org/grpc"
)

type Gateway struct {
	grpcServer *grpc.Server
	httpServer *http.Server
	wsServer   *websocket.Server
	host       host.Host
	config     *Config
	running    bool
	mutex      sync.RWMutex
	grpcAddr   string
	httpAddr   string
}

type Config struct {
	GRPCPort int
	HTTPPort int
}

func NewGateway(cfg *Config, h host.Host, msgService *grpcapi.MessageService, mediaService *grpcapi.MediaService, callService *grpcapi.CallService, peerService *grpcapi.PeerService, groupService *grpcapi.GroupService, wsServer *websocket.Server) *Gateway {
	grpcServer := grpc.NewServer()

	// Register services
	pb.RegisterMessageServiceServer(grpcServer, msgService)
	pb.RegisterMediaServiceServer(grpcServer, mediaService)
	pb.RegisterCallServiceServer(grpcServer, callService)
	pb.RegisterPeerServiceServer(grpcServer, peerService)
	pb.RegisterGroupServiceServer(grpcServer, groupService)

	return &Gateway{
		grpcServer: grpcServer,
		wsServer:   wsServer,
		host:       h,
		config:     cfg,
	}
}

func (g *Gateway) Start() error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// Start gRPC server
	go g.startGRPC()

	// Start HTTP/WebSocket server
	go g.startHTTP()

	g.running = true
	return nil
}

func (g *Gateway) Stop() error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if g.grpcServer != nil {
		g.grpcServer.GracefulStop()
	}

	if g.httpServer != nil {
		g.httpServer.Close()
	}

	g.running = false
	return nil
}

func (g *Gateway) IsRunning() bool {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.running
}

func (g *Gateway) GRPCAddr() string {
	return g.grpcAddr
}

func (g *Gateway) HTTPAddr() string {
	return g.httpAddr
}

func (g *Gateway) GetGRPCServer() *grpc.Server {
	return g.grpcServer
}

func (g *Gateway) startGRPC() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", g.config.GRPCPort))
	if err != nil {
		return err
	}

	g.grpcAddr = lis.Addr().String()
	return g.grpcServer.Serve(lis)
}

func (g *Gateway) startHTTP() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", g.wsServer.HandleConnection)
	mux.HandleFunc("/peer-id", g.handlePeerID)
	mux.HandleFunc("/peers", g.handlePeers)

	// Create listener first to get actual address
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", g.config.HTTPPort))
	if err != nil {
		return err
	}

	server := &http.Server{
		Handler: mux,
	}

	g.httpServer = server
	g.httpAddr = lis.Addr().String()

	return server.Serve(lis)
}

func (g *Gateway) handlePeerID(w http.ResponseWriter, r *http.Request) {
	if g.host == nil {
		http.Error(w, "Host not available", http.StatusInternalServerError)
		return
	}

	peerID := g.host.ID().String()
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(peerID))
}

func (g *Gateway) handlePeers(w http.ResponseWriter, r *http.Request) {
	if g.host == nil {
		http.Error(w, "Host not available", http.StatusInternalServerError)
		return
	}

	// Get all connected peers
	peers := g.host.Network().Peers()

	// Build response
	response := fmt.Sprintf("Connected peers: %d\n", len(peers))
	for _, peerID := range peers {
		addrs := g.host.Peerstore().Addrs(peerID)
		response += fmt.Sprintf("Peer: %s\n", peerID)
		for _, addr := range addrs {
			response += fmt.Sprintf("  Address: %s\n", addr)
		}
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
}
