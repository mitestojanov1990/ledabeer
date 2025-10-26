package api

import (
	"fmt"
	"net"
	"net/http"
	"sync"

	grpcapi "ledabeer/backend/internal/api/grpc"
	"ledabeer/backend/internal/api/websocket"

	pb "ledabeer/backend/pkg/proto"

	"google.golang.org/grpc"
)

type Gateway struct {
	grpcServer *grpc.Server
	httpServer *http.Server
	wsServer   *websocket.Server
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

func NewGateway(cfg *Config, msgService *grpcapi.MessageService, mediaService *grpcapi.MediaService, callService *grpcapi.CallService, wsServer *websocket.Server) *Gateway {
	grpcServer := grpc.NewServer()

	// Register services
	pb.RegisterMessageServiceServer(grpcServer, msgService)
	pb.RegisterMediaServiceServer(grpcServer, mediaService)
	pb.RegisterCallServiceServer(grpcServer, callService)

	return &Gateway{
		grpcServer: grpcServer,
		wsServer:   wsServer,
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
