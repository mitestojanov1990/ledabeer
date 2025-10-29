package conversations

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"ledabeer/backend/internal/user"
	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p/core/peer"
)

// ConversationService manages conversations between users
type ConversationService struct {
	userManager *user.UserManager
	// Map conversation ID to conversation
	conversations map[string]*Conversation
	// Map user ID to their conversation IDs
	userConversations map[string][]string
	mutex            sync.RWMutex
}

// NewConversationService creates a new ConversationService
func NewConversationService(userManager *user.UserManager) *ConversationService {
	return &ConversationService{
		userManager:       userManager,
		conversations:     make(map[string]*Conversation),
		userConversations: make(map[string][]string),
	}
}

// Conversation represents a conversation between users
type Conversation struct {
	ID           string                `json:"id"`
	Type         string                `json:"type"` // "direct" or "group"
	Name         string                `json:"name"` // For group chats
	Participants []ConversationMember  `json:"participants"`
	LastMessage  *MessagePreview       `json:"last_message,omitempty"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
	UnreadCount  map[string]int        `json:"unread_count"` // User ID -> unread count
}

// ConversationMember represents a member in a conversation
type ConversationMember struct {
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url"`
	JoinedAt    time.Time `json:"joined_at"`
	IsOnline    bool      `json:"is_online"`
}

// MessagePreview represents a preview of the last message
type MessagePreview struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	FromUser  string    `json:"from_user"`
	FromName  string    `json:"from_name"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"` // "text", "image", "file", etc.
}

// CreateConversationRequest represents a request to create a conversation
type CreateConversationRequest struct {
	CreatorID     string   `json:"creator_id"`
	ParticipantID string   `json:"participant_id"` // For direct conversations
	Type          string   `json:"type"`           // "direct" or "group"
	Name          string   `json:"name,omitempty"` // For group conversations
}

// CreateConversationResponse represents the response for creating a conversation
type CreateConversationResponse struct {
	Conversation *Conversation `json:"conversation"`
	Message      string        `json:"message"`
}

// GetUserConversationsRequest represents a request to get user conversations
type GetUserConversationsRequest struct {
	UserID string `json:"user_id"`
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
}

// GetUserConversationsResponse represents the response for getting user conversations
type GetUserConversationsResponse struct {
	Conversations []*Conversation `json:"conversations"`
	Total         int             `json:"total"`
	HasMore       bool            `json:"has_more"`
}

// AddMessageRequest represents a request to add a message to a conversation
type AddMessageRequest struct {
	ConversationID string `json:"conversation_id"`
	FromUserID     string `json:"from_user_id"`
	Content        string `json:"content"`
	Type           string `json:"type"` // "text", "image", "file", etc.
}

// AddMessageResponse represents the response for adding a message
type AddMessageResponse struct {
	MessageID      string    `json:"message_id"`
	ConversationID string    `json:"conversation_id"`
	Timestamp      time.Time `json:"timestamp"`
}

// CreateConversation creates a new conversation
func (cs *ConversationService) CreateConversation(req *CreateConversationRequest) (*CreateConversationResponse, error) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	// Convert string ID to peer.ID
	creatorPeerID, err := peer.Decode(req.CreatorID)
	if err != nil {
		return nil, fmt.Errorf("invalid creator ID: %w", err)
	}
	
	// Validate creator exists
	creatorInfo, err := cs.userManager.GetUserByPeerID(creatorPeerID)
	if err != nil {
		return nil, fmt.Errorf("creator validation failed: %w", err)
	}

	// Validate conversation type
	if req.Type != "direct" && req.Type != "group" {
		return nil, errors.New("conversation type must be 'direct' or 'group'")
	}

	var conversation *Conversation
	var conversationID string

	if req.Type == "direct" {
		// Convert participant ID to peer.ID
		participantPeerID, err := peer.Decode(req.ParticipantID)
		if err != nil {
			return nil, fmt.Errorf("invalid participant ID: %w", err)
		}
		
		// Validate participant exists
		participantInfo, err := cs.userManager.GetUserByPeerID(participantPeerID)
		if err != nil {
			return nil, fmt.Errorf("participant validation failed: %w", err)
		}

		// Check if direct conversation already exists
		existingConvID := cs.findDirectConversation(req.CreatorID, req.ParticipantID)
		if existingConvID != "" {
			conversation = cs.conversations[existingConvID]
			return &CreateConversationResponse{
				Conversation: conversation,
				Message:      "Conversation already exists",
			}, nil
		}

		// Create direct conversation
		conversationID = fmt.Sprintf("conv_%s_%s_%d", req.CreatorID, req.ParticipantID, time.Now().Unix())
		conversation = &Conversation{
			ID:   conversationID,
			Type: "direct",
			Name: fmt.Sprintf("%s & %s", creatorInfo.DisplayName, participantInfo.DisplayName),
			Participants: []ConversationMember{
				{
					UserID:      creatorInfo.UserID,
					Username:    creatorInfo.Username,
					DisplayName: creatorInfo.DisplayName,
					AvatarURL:   creatorInfo.AvatarURL,
					JoinedAt:    time.Now(),
					IsOnline:    creatorInfo.IsOnline,
				},
				{
					UserID:      participantInfo.UserID,
					Username:    participantInfo.Username,
					DisplayName: participantInfo.DisplayName,
					AvatarURL:   participantInfo.AvatarURL,
					JoinedAt:    time.Now(),
					IsOnline:    participantInfo.IsOnline,
				},
			},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			UnreadCount: make(map[string]int),
		}
	} else {
		// Group conversation logic would go here
		return nil, errors.New("group conversations not yet implemented")
	}

	// Store conversation
	cs.conversations[conversationID] = conversation

	// Add conversation to user's conversation list
	cs.userConversations[req.CreatorID] = append(cs.userConversations[req.CreatorID], conversationID)
	cs.userConversations[req.ParticipantID] = append(cs.userConversations[req.ParticipantID], conversationID)

	return &CreateConversationResponse{
		Conversation: conversation,
		Message:      "Conversation created successfully",
	}, nil
}

// GetUserConversations gets all conversations for a user
func (cs *ConversationService) GetUserConversations(req *GetUserConversationsRequest) (*GetUserConversationsResponse, error) {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	// Convert user ID to peer.ID
	userPeerID, err := peer.Decode(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}
	
	// Validate user exists
	_, err = cs.userManager.GetUserByPeerID(userPeerID)
	if err != nil {
		return nil, fmt.Errorf("user validation failed: %w", err)
	}

	// Set default values
	if req.Limit <= 0 {
		req.Limit = 50
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	// Get user's conversation IDs
	conversationIDs, exists := cs.userConversations[req.UserID]
	if !exists {
		return &GetUserConversationsResponse{
			Conversations: []*Conversation{},
			Total:         0,
			HasMore:       false,
		}, nil
	}

	// Get conversations
	var conversations []*Conversation
	for _, convID := range conversationIDs {
		if conv, exists := cs.conversations[convID]; exists {
			conversations = append(conversations, conv)
		}
	}

	// Sort by updated time (most recent first)
	// In a real implementation, you'd use a proper sorting algorithm
	// For now, we'll just return them as-is

	// Apply pagination
	total := len(conversations)
	start := req.Offset
	end := start + req.Limit
	if end > total {
		end = total
	}

	if start >= total {
		conversations = []*Conversation{}
	} else {
		conversations = conversations[start:end]
	}

	return &GetUserConversationsResponse{
		Conversations: conversations,
		Total:         total,
		HasMore:       end < total,
	}, nil
}

// AddMessage adds a message to a conversation
func (cs *ConversationService) AddMessage(req *AddMessageRequest) (*AddMessageResponse, error) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	// Validate conversation exists
	conversation, exists := cs.conversations[req.ConversationID]
	if !exists {
		return nil, errors.New("conversation not found")
	}

	// Validate user is participant
	isParticipant := false
	for _, participant := range conversation.Participants {
		if participant.UserID == req.FromUserID {
			isParticipant = true
			break
		}
	}
	if !isParticipant {
		return nil, errors.New("user is not a participant in this conversation")
	}

	// Create message preview
	messageID := uuid.New().String()
	messagePreview := &MessagePreview{
		ID:        messageID,
		Content:   req.Content,
		FromUser:  req.FromUserID,
		FromName:  cs.getParticipantName(conversation, req.FromUserID),
		Timestamp: time.Now(),
		Type:      req.Type,
	}

	// Update conversation
	conversation.LastMessage = messagePreview
	conversation.UpdatedAt = time.Now()

	// Increment unread count for other participants
	for _, participant := range conversation.Participants {
		if participant.UserID != req.FromUserID {
			conversation.UnreadCount[participant.UserID]++
		}
	}

	return &AddMessageResponse{
		MessageID:      messageID,
		ConversationID: req.ConversationID,
		Timestamp:      messagePreview.Timestamp,
	}, nil
}

// MarkAsRead marks messages as read for a user in a conversation
func (cs *ConversationService) MarkAsRead(conversationID, userID string) error {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	conversation, exists := cs.conversations[conversationID]
	if !exists {
		return errors.New("conversation not found")
	}

	// Reset unread count for user
	conversation.UnreadCount[userID] = 0
	conversation.UpdatedAt = time.Now()

	return nil
}

// GetConversation gets a specific conversation
func (cs *ConversationService) GetConversation(conversationID, userID string) (*Conversation, error) {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	conversation, exists := cs.conversations[conversationID]
	if !exists {
		return nil, errors.New("conversation not found")
	}

	// Validate user is participant
	isParticipant := false
	for _, participant := range conversation.Participants {
		if participant.UserID == userID {
			isParticipant = true
			break
		}
	}
	if !isParticipant {
		return nil, errors.New("user is not a participant in this conversation")
	}

	return conversation, nil
}

// findDirectConversation finds existing direct conversation between two users
func (cs *ConversationService) findDirectConversation(user1ID, user2ID string) string {
	for convID, conv := range cs.conversations {
		if conv.Type == "direct" && len(conv.Participants) == 2 {
			participantIDs := make(map[string]bool)
			for _, p := range conv.Participants {
				participantIDs[p.UserID] = true
			}
			if participantIDs[user1ID] && participantIDs[user2ID] {
				return convID
			}
		}
	}
	return ""
}

// getParticipantName gets the display name of a participant
func (cs *ConversationService) getParticipantName(conversation *Conversation, userID string) string {
	for _, participant := range conversation.Participants {
		if participant.UserID == userID {
			return participant.DisplayName
		}
	}
	return "Unknown User"
}
