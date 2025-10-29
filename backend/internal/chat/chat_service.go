package chat

import (
	"errors"
	"fmt"
	"time"

	"ledabeer/backend/internal/auth"
)

// ChatService handles chat-related operations with authentication
type ChatService struct {
	userService *auth.UserService
	// Add other dependencies like message service, peer service, etc.
}

// NewChatService creates a new ChatService
func NewChatService(userService *auth.UserService) *ChatService {
	return &ChatService{
		userService: userService,
	}
}

// CreateConversationRequest represents a request to create a new conversation
type CreateConversationRequest struct {
	ParticipantID string `json:"participant_id" validate:"required"`
	Type          string `json:"type" validate:"required,oneof=direct group"` // direct or group
}

// CreateConversationResponse represents the response for creating a conversation
type CreateConversationResponse struct {
	ConversationID string                `json:"conversation_id"`
	Type           string                `json:"type"`
	Participants   []ConversationMember  `json:"participants"`
	CreatedAt      time.Time             `json:"created_at"`
	LastMessage    *MessagePreview       `json:"last_message,omitempty"`
}

// ConversationMember represents a member in a conversation
type ConversationMember struct {
	UserID     string `json:"user_id"`
	Username   string `json:"username"`
	DisplayName string `json:"display_name"`
	AvatarURL  string `json:"avatar_url"`
	IsOnline   bool   `json:"is_online"`
	JoinedAt   time.Time `json:"joined_at"`
}

// MessagePreview represents a preview of the last message
type MessagePreview struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	FromUser  string    `json:"from_user"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
}

// CreateDirectConversation creates a direct conversation between two users
func (cs *ChatService) CreateDirectConversation(currentUserID, participantID string) (*CreateConversationResponse, error) {
	// Validate that both users exist and are active
	if err := cs.userService.ValidateUserExists(currentUserID); err != nil {
		return nil, fmt.Errorf("current user validation failed: %w", err)
	}

	if err := cs.userService.ValidateUserExists(participantID); err != nil {
		return nil, fmt.Errorf("participant user validation failed: %w", err)
	}

	// Prevent self-conversation
	if currentUserID == participantID {
		return nil, errors.New("cannot create conversation with yourself")
	}

	// Get user display info for both users
	currentUserInfo, err := cs.userService.GetUserDisplayInfo(currentUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current user info: %w", err)
	}

	participantUserInfo, err := cs.userService.GetUserDisplayInfo(participantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participant user info: %w", err)
	}

	// Generate conversation ID (in a real system, this would be stored in database)
	conversationID := fmt.Sprintf("conv_%s_%s_%d", currentUserID, participantID, time.Now().Unix())

	// Create conversation response
	response := &CreateConversationResponse{
		ConversationID: conversationID,
		Type:           "direct",
		Participants: []ConversationMember{
			{
				UserID:      currentUserInfo.ID,
				Username:    currentUserInfo.Username,
				DisplayName: currentUserInfo.DisplayName,
				AvatarURL:   currentUserInfo.AvatarURL,
				IsOnline:    false, // This would be determined by peer status
				JoinedAt:    time.Now(),
			},
			{
				UserID:      participantUserInfo.ID,
				Username:    participantUserInfo.Username,
				DisplayName: participantUserInfo.DisplayName,
				AvatarURL:   participantUserInfo.AvatarURL,
				IsOnline:    false, // This would be determined by peer status
				JoinedAt:    time.Now(),
			},
		},
		CreatedAt: time.Now(),
	}

	return response, nil
}

// GetConversationParticipants gets the participants of a conversation
func (cs *ChatService) GetConversationParticipants(conversationID string) ([]ConversationMember, error) {
	// In a real system, this would query the database
	// For now, return empty slice
	return []ConversationMember{}, nil
}

// ValidateConversationAccess validates if a user can access a conversation
func (cs *ChatService) ValidateConversationAccess(userID, conversationID string) error {
	// In a real system, this would check if the user is a participant
	// For now, always return success
	return nil
}

// GetUserConversations gets all conversations for a user
func (cs *ChatService) GetUserConversations(userID string) ([]ConversationSummary, error) {
	// In a real system, this would query the database
	// For now, return empty slice
	return []ConversationSummary{}, nil
}

// ConversationSummary represents a summary of a conversation
type ConversationSummary struct {
	ConversationID string                `json:"conversation_id"`
	Type           string                `json:"type"`
	Name           string                `json:"name"`           // For group chats
	Participants   []ConversationMember  `json:"participants"`
	LastMessage    *MessagePreview       `json:"last_message,omitempty"`
	UnreadCount    int                   `json:"unread_count"`
	UpdatedAt      time.Time             `json:"updated_at"`
}
