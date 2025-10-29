package api

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"

	grpcapi "ledabeer/backend/internal/api/grpc"
	"ledabeer/backend/internal/api/websocket"
	"ledabeer/backend/internal/auth"
	"ledabeer/backend/internal/conversations"
	"ledabeer/backend/internal/messaging"
	"ledabeer/backend/internal/user"

	pb "ledabeer/backend/pkg/proto"

	"github.com/libp2p/go-libp2p/core/host"
	"google.golang.org/grpc"
)

// AuthHandlers handles authentication HTTP endpoints
type AuthHandlers struct {
	authenticator *auth.Authenticator
}

// NewAuthHandlers creates a new AuthHandlers instance
func NewAuthHandlers(authenticator *auth.Authenticator) *AuthHandlers {
	return &AuthHandlers{
		authenticator: authenticator,
	}
}

// RegisterUser handles user registration
func (h *AuthHandlers) RegisterUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req auth.UserRegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}

	// Basic validation
	if req.Username == "" || req.Email == "" || req.Password == "" || req.DisplayName == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Missing required fields"})
		return
	}

	// Register user
	user, err := h.authenticator.RegisterUser(&req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// LoginUser handles user login
func (h *AuthHandlers) LoginUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req auth.UserLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}

	// Basic validation
	if req.Username == "" || req.Password == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Missing username or password"})
		return
	}

	// Login user
	tokens, user, err := h.authenticator.LoginUser(&req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	response := map[string]interface{}{
		"user":  user,
		"token": tokens,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// RefreshToken handles token refresh
func (h *AuthHandlers) RefreshToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}

	if req.RefreshToken == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Missing refresh token"})
		return
	}

	// Refresh token
	tokens, err := h.authenticator.RefreshToken(req.RefreshToken)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tokens)
}

// GetCurrentUser returns the current authenticated user
func (h *AuthHandlers) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Missing Authorization header"})
		return
	}

	// Validate token and get user
	user, err := h.authenticator.ValidateToken(authHeader)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user.ToResponse())
}

// LogoutUser handles user logout (client-side token removal)
func (h *AuthHandlers) LogoutUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// For JWT tokens, logout is handled client-side by removing the token
	// In a more sophisticated system, you might maintain a blacklist of tokens
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}

type Gateway struct {
	grpcServer          *grpc.Server
	httpServer          *http.Server
	wsServer            *websocket.Server
	host                host.Host
	msgHandler          *messaging.MessageHandler
	authenticator       *auth.Authenticator
	userManager         *user.UserManager
	conversationService *conversations.ConversationService
	config              *Config
	running             bool
	mutex               sync.RWMutex
	grpcAddr            string
	httpAddr            string
}

type Config struct {
	GRPCPort int
	HTTPPort int
}

func NewGateway(cfg *Config, h host.Host, msgHandler *messaging.MessageHandler, msgService *grpcapi.MessageService, mediaService *grpcapi.MediaService, callService *grpcapi.CallService, peerService *grpcapi.PeerService, groupService *grpcapi.GroupService, wsServer *websocket.Server, authenticator *auth.Authenticator) *Gateway {
	grpcServer := grpc.NewServer()

	// Register services
	pb.RegisterMessageServiceServer(grpcServer, msgService)
	pb.RegisterMediaServiceServer(grpcServer, mediaService)
	pb.RegisterCallServiceServer(grpcServer, callService)
	pb.RegisterPeerServiceServer(grpcServer, peerService)
	pb.RegisterGroupServiceServer(grpcServer, groupService)

	// Initialize user manager and conversation service
	userManager := user.NewUserManager()
	conversationService := conversations.NewConversationService(userManager)

	return &Gateway{
		grpcServer:          grpcServer,
		wsServer:            wsServer,
		host:                h,
		msgHandler:          msgHandler,
		authenticator:       authenticator,
		userManager:         userManager,
		conversationService: conversationService,
		config:              cfg,
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

	// WebSocket endpoint
	mux.HandleFunc("/ws", g.wsServer.HandleConnection)

	// Legacy peer endpoints
	mux.HandleFunc("/peer-id", g.handlePeerID)
	mux.HandleFunc("/peers", g.handlePeers)

	// API endpoints
	mux.HandleFunc("/api/peers", g.handleAPIPeers)
	mux.HandleFunc("/api/send-message", g.handleAPISendMessage)

	// Authentication endpoints
	authHandlers := NewAuthHandlers(g.authenticator)
	mux.HandleFunc("/api/auth/register", authHandlers.RegisterUser)
	mux.HandleFunc("/api/auth/login", authHandlers.LoginUser)
	mux.HandleFunc("/api/auth/refresh", authHandlers.RefreshToken)
	mux.HandleFunc("/api/auth/me", authHandlers.GetCurrentUser)
	mux.HandleFunc("/api/auth/logout", authHandlers.LogoutUser)

	// User search endpoints
	userSearchHandlers := NewUserSearchHandlers(g.userManager)
	mux.HandleFunc("/api/users/search", userSearchHandlers.SearchUsers)
	mux.HandleFunc("/api/users/find-by-email", userSearchHandlers.FindUserByEmail)
	mux.HandleFunc("/api/users/find-by-username", userSearchHandlers.FindUserByUsername)

	// Conversation endpoints
	conversationHandlers := NewConversationHandlers(g.conversationService, g.userManager)
	mux.HandleFunc("/api/conversations", conversationHandlers.CreateConversation)
	mux.HandleFunc("/api/conversations/list", conversationHandlers.GetUserConversations)
	mux.HandleFunc("/api/conversations/get", conversationHandlers.GetConversation)
	mux.HandleFunc("/api/conversations/messages", conversationHandlers.AddMessage)
	mux.HandleFunc("/api/conversations/mark-read", conversationHandlers.MarkAsRead)

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

func (g *Gateway) handleAPIPeers(w http.ResponseWriter, r *http.Request) {
	// Call the gRPC service internally
	peerService := grpcapi.NewPeerService(g.host)
	ctx := r.Context()

	response, err := peerService.GetPeers(ctx, &pb.GetPeersRequest{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert to JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Simple JSON response
	json := `{"peers":[`
	for i, peer := range response.Peers {
		if i > 0 {
			json += ","
		}
		json += fmt.Sprintf(`{"id":"%s","name":"%s","online":%t,"addresses":[`, peer.Id, peer.Name, peer.Online)
		for j, addr := range peer.Addresses {
			if j > 0 {
				json += ","
			}
			json += fmt.Sprintf(`"%s"`, addr)
		}
		json += "]}"
	}
	json += "]}"

	w.Write([]byte(json))
}

func (g *Gateway) handleAPISendMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		To      string `json:"to"`
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}

	// Call the gRPC service internally
	// We need to create a groupManager, but for now we'll use nil
	messageService := grpcapi.NewMessageService(g.msgHandler, nil)
	ctx := r.Context()

	response, err := messageService.SendMessage(ctx, &pb.SendMessageRequest{
		ToPeerId: req.To,
		Content:  []byte(req.Content),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"messageId": response.MessageId,
	})
}
