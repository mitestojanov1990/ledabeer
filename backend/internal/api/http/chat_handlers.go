package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/ledabeer/backend/internal/auth"
	"github.com/ledabeer/backend/internal/chat"
)

// ChatHandlers handles chat-related HTTP endpoints
type ChatHandlers struct {
	chatService *chat.ChatService
}

// NewChatHandlers creates a new ChatHandlers instance
func NewChatHandlers(chatService *chat.ChatService) *ChatHandlers {
	return &ChatHandlers{
		chatService: chatService,
	}
}

// CreateConversation handles creating a new conversation
func (h *ChatHandlers) CreateConversation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract current user ID from JWT token (you'd need to implement this)
	currentUserID, err := h.extractUserIDFromToken(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid or missing token"})
		return
	}

	var req chat.CreateConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}

	// Validate request
	if req.ParticipantID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Participant ID is required"})
		return
	}

	if req.Type != "direct" && req.Type != "group" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Type must be 'direct' or 'group'"})
		return
	}

	// Create conversation based on type
	var response *chat.CreateConversationResponse
	if req.Type == "direct" {
		response, err = h.chatService.CreateDirectConversation(currentUserID, req.ParticipantID)
	} else {
		// Group conversation logic would go here
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotImplemented)
		json.NewEncoder(w).Encode(map[string]string{"error": "Group conversations not yet implemented"})
		return
	}

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetUserConversations handles getting all conversations for a user
func (h *ChatHandlers) GetUserConversations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract current user ID from JWT token
	currentUserID, err := h.extractUserIDFromToken(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid or missing token"})
		return
	}

	// Get user conversations
	conversations, err := h.chatService.GetUserConversations(currentUserID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get conversations"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"conversations": conversations,
		"total":         len(conversations),
	})
}

// GetConversationParticipants handles getting participants of a conversation
func (h *ChatHandlers) GetConversationParticipants(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract current user ID from JWT token
	currentUserID, err := h.extractUserIDFromToken(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid or missing token"})
		return
	}

	// Extract conversation ID from URL path
	conversationID := r.URL.Query().Get("conversation_id")
	if conversationID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Conversation ID is required"})
		return
	}

	// Validate user has access to conversation
	if err := h.chatService.ValidateConversationAccess(currentUserID, conversationID); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Access denied"})
		return
	}

	// Get participants
	participants, err := h.chatService.GetConversationParticipants(conversationID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get participants"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"participants": participants,
		"total":        len(participants),
	})
}

// extractUserIDFromToken extracts user ID from JWT token in Authorization header
func (h *ChatHandlers) extractUserIDFromToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("missing authorization header")
	}

	// Extract token from "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid authorization header format")
	}

	token := parts[1]

	// In a real implementation, you would validate the JWT token here
	// and extract the user ID from the claims
	// For now, we'll return a placeholder
	// This should be implemented using the auth.TokenManager

	return "", errors.New("token validation not implemented")
}
