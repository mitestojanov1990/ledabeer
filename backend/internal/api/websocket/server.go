package websocket

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"ledabeer/backend/internal/auth"
	"ledabeer/backend/internal/conversations"

	"github.com/gorilla/websocket"
)

type Server struct {
	upgrader           websocket.Upgrader
	clients            map[string]*Client
	conversationService *conversations.ConversationService
	mutex              sync.RWMutex
	auth               *auth.Authenticator
}

type Client struct {
	conn     *websocket.Conn
	userID   string
	peerID   string
	send     chan []byte
}

type RealtimeMessage struct {
	Type           string                 `json:"type"`
	ConversationID string                 `json:"conversation_id,omitempty"`
	Message        *conversations.MessagePreview `json:"message,omitempty"`
	Conversation   *conversations.Conversation   `json:"conversation,omitempty"`
	UserID         string                 `json:"user_id,omitempty"`
	IsOnline       bool                   `json:"is_online,omitempty"`
}

func NewServer(auth *auth.Authenticator, conversationService *conversations.ConversationService) *Server {
	return &Server{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		clients:            make(map[string]*Client),
		conversationService: conversationService,
		auth:               auth,
	}
}

func (s *Server) HandleConnection(w http.ResponseWriter, r *http.Request) {
	// Extract token from query parameter
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
	}

	// Validate token and get user ID
	user, err := s.auth.ValidateAccessToken(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &Client{
		conn:   conn,
		userID: user.ID,
		send:   make(chan []byte, 256),
	}

	// Register client
	s.mutex.Lock()
	s.clients[user.ID] = client
	s.mutex.Unlock()

	go client.readPump(s)
	go client.writePump()

	// Cleanup on disconnect
	defer func() {
		s.mutex.Lock()
		delete(s.clients, user.ID)
		s.mutex.Unlock()
		conn.Close()
	}()
}

func (s *Server) BroadcastMessage(peerID string, data []byte) {
	s.mutex.RLock()
	client, exists := s.clients[peerID]
	s.mutex.RUnlock()

	if exists {
		// Create a message event
		msg := map[string]interface{}{
			"type":    "message",
			"content": string(data),
		}

		msgBytes, _ := json.Marshal(msg)

		select {
		case client.send <- msgBytes:
		default:
			close(client.send)
			s.mutex.Lock()
			delete(s.clients, peerID)
			s.mutex.Unlock()
		}
	}
}

// BroadcastNewMessage broadcasts a new message to all participants in a conversation
func (s *Server) BroadcastNewMessage(conversationID string, message *conversations.MessagePreview) {
	// Get conversation to find participants
	conversation, err := s.conversationService.GetConversation(conversationID, "")
	if err != nil {
		return
	}

	realtimeMsg := RealtimeMessage{
		Type:           "new_message",
		ConversationID: conversationID,
		Message:        message,
	}

	data, err := json.Marshal(realtimeMsg)
	if err != nil {
		return
	}

	// Send to all participants
	for _, participant := range conversation.Participants {
		s.BroadcastMessage(participant.UserID, data)
	}
}

// BroadcastConversationUpdate broadcasts conversation updates to all participants
func (s *Server) BroadcastConversationUpdate(conversation *conversations.Conversation) {
	realtimeMsg := RealtimeMessage{
		Type:         "conversation_update",
		Conversation: conversation,
	}

	data, err := json.Marshal(realtimeMsg)
	if err != nil {
		return
	}

	// Send to all participants
	for _, participant := range conversation.Participants {
		s.BroadcastMessage(participant.UserID, data)
	}
}

// BroadcastUserStatusUpdate broadcasts user online/offline status
func (s *Server) BroadcastUserStatusUpdate(userID string, isOnline bool) {
	realtimeMsg := RealtimeMessage{
		Type:     "user_status_update",
		UserID:   userID,
		IsOnline: isOnline,
	}

	data, err := json.Marshal(realtimeMsg)
	if err != nil {
		return
	}

	// Broadcast to all connected clients
	s.mutex.RLock()
	for _, client := range s.clients {
		select {
		case client.send <- data:
		default:
		}
	}
	s.mutex.RUnlock()
}

func (c *Client) readPump(server *Server) {
	defer func() {
		server.mutex.Lock()
		delete(server.clients, c.peerID)
		server.mutex.Unlock()
		c.conn.Close()
	}()

	for {
		var msg map[string]interface{}
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			break
		}

		// Handle authentication
		if msgType, ok := msg["type"].(string); ok && msgType == "auth" {
			if peerID, ok := msg["peer_id"].(string); ok {
				c.peerID = peerID
				server.mutex.Lock()
				server.clients[peerID] = c
				server.mutex.Unlock()

				// Send auth success
				authResp := map[string]string{
					"type": "auth_success",
				}
				c.conn.WriteJSON(authResp)
			}
		}
	}
}

func (c *Client) writePump() {
	defer c.conn.Close()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Send message as JSON
			var msg map[string]interface{}
			json.Unmarshal(message, &msg)
			c.conn.WriteJSON(msg)
		}
	}
}
