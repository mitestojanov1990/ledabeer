package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"ledabeer/backend/internal/api/websocket"
	"ledabeer/backend/internal/conversations"
	"ledabeer/backend/internal/user"
)

// ConversationHandlers handles conversation-related HTTP endpoints
type ConversationHandlers struct {
	conversationService *conversations.ConversationService
	userManager         *user.UserManager
	wsServer            *websocket.Server
}

// NewConversationHandlers creates a new ConversationHandlers instance
func NewConversationHandlers(conversationService *conversations.ConversationService, userManager *user.UserManager, wsServer *websocket.Server) *ConversationHandlers {
	return &ConversationHandlers{
		conversationService: conversationService,
		userManager:         userManager,
		wsServer:            wsServer,
	}
}

// CreateConversation handles creating a new conversation
func (h *ConversationHandlers) CreateConversation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract current user ID from JWT token (you'd need to implement this)
	currentUserID, err := h.extractUserIDFromToken(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Invalid or missing token")
		return
	}

	var req conversations.CreateConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Set creator ID from token
	req.CreatorID = currentUserID

	// Validate request
	if req.ParticipantID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Participant ID is required")
		return
	}

	if req.Type == "" {
		req.Type = "direct" // Default to direct conversation
	}

	// Create conversation
	response, err := h.conversationService.CreateConversation(&req)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetUserConversations handles getting all conversations for a user
func (h *ConversationHandlers) GetUserConversations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract current user ID from JWT token
	currentUserID, err := h.extractUserIDFromToken(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Invalid or missing token")
		return
	}

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50 // Default limit
	offset := 0 // Default offset

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get conversations
	req := &conversations.GetUserConversationsRequest{
		UserID: currentUserID,
		Limit:  limit,
		Offset: offset,
	}

	response, err := h.conversationService.GetUserConversations(req)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get conversations")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetConversation handles getting a specific conversation
func (h *ConversationHandlers) GetConversation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract current user ID from JWT token
	currentUserID, err := h.extractUserIDFromToken(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Invalid or missing token")
		return
	}

	// Extract conversation ID from URL path
	conversationID := r.URL.Query().Get("conversation_id")
	if conversationID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Conversation ID is required")
		return
	}

	// Get conversation
	conversation, err := h.conversationService.GetConversation(conversationID, currentUserID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "Conversation not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(conversation)
}

// AddMessage handles adding a message to a conversation
func (h *ConversationHandlers) AddMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract current user ID from JWT token
	currentUserID, err := h.extractUserIDFromToken(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Invalid or missing token")
		return
	}

	var req conversations.AddMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Set from user ID from token
	req.FromUserID = currentUserID

	// Validate request
	if req.ConversationID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Conversation ID is required")
		return
	}

	if req.Content == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Content is required")
		return
	}

	if req.Type == "" {
		req.Type = "text" // Default to text message
	}

	// Add message
	response, err := h.conversationService.AddMessage(&req)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Broadcast real-time message to all participants
	if h.wsServer != nil {
		messagePreview := &conversations.MessagePreview{
			ID:        response.MessageID,
			Content:   req.Content,
			FromUser:  req.FromUserID,
			FromName:  "User", // You'd get this from user service
			Timestamp: response.Timestamp,
			Type:      req.Type,
		}
		h.wsServer.BroadcastNewMessage(req.ConversationID, messagePreview)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// MarkAsRead handles marking messages as read
func (h *ConversationHandlers) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract current user ID from JWT token
	currentUserID, err := h.extractUserIDFromToken(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Invalid or missing token")
		return
	}

	var req struct {
		ConversationID string `json:"conversation_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validate request
	if req.ConversationID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Conversation ID is required")
		return
	}

	// Mark as read
	err = h.conversationService.MarkAsRead(req.ConversationID, currentUserID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Messages marked as read"})
}

// extractUserIDFromToken extracts user ID from JWT token in Authorization header
func (h *ConversationHandlers) extractUserIDFromToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("missing Authorization header")
	}

	// Extract token from "Bearer <token>"
	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
		return "", errors.New("invalid Authorization header format")
	}

	// For now, we'll extract from a custom header for testing
	// In production, you'd validate the JWT token and extract user ID from claims
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		return "", errors.New("user ID not found in token")
	}
	
	return userID, nil
}

// writeErrorResponse writes a JSON error response
func (h *ConversationHandlers) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := map[string]string{"error": message}
	json.NewEncoder(w).Encode(response)
}
