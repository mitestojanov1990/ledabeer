package calls

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"ledabeer/backend/internal/user"
	"github.com/libp2p/go-libp2p/core/peer"
)

// CallService handles voice/video calls for authenticated users
type CallService struct {
	userManager *user.UserManager
	// Active calls map: callID -> CallSession
	activeCalls map[string]*CallSession
	mutex       sync.RWMutex
}

// NewCallService creates a new CallService
func NewCallService(userManager *user.UserManager) *CallService {
	return &CallService{
		userManager: userManager,
		activeCalls: make(map[string]*CallSession),
	}
}

// CallSession represents an active call session
type CallSession struct {
	CallID        string           `json:"call_id"`
	CallType      string           `json:"call_type"` // "voice" or "video"
	InitiatorID   string           `json:"initiator_id"`
	Participants  []CallParticipant `json:"participants"`
	Status        string           `json:"status"` // "ringing", "active", "ended"
	StartedAt     time.Time        `json:"started_at"`
	EndedAt       *time.Time       `json:"ended_at,omitempty"`
	WebRTCConfig  *WebRTCConfig    `json:"webrtc_config"`
}

// CallParticipant represents a participant in a call
type CallParticipant struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	PeerID      string `json:"peer_id"`
	IsOnline    bool   `json:"is_online"`
	IsMuted     bool   `json:"is_muted"`
	IsVideoOn   bool   `json:"is_video_on"`
	JoinedAt    time.Time `json:"joined_at"`
}

// WebRTCConfig contains WebRTC configuration for the call
type WebRTCConfig struct {
	ICEServers []ICEServer `json:"ice_servers"`
	STUNServers []string   `json:"stun_servers"`
	TURNServers []TURNServer `json:"turn_servers"`
}

// ICEServer represents an ICE server configuration
type ICEServer struct {
	URLs       []string `json:"urls"`
	Username   string   `json:"username,omitempty"`
	Credential string   `json:"credential,omitempty"`
}

// TURNServer represents a TURN server configuration
type TURNServer struct {
	URLs       []string `json:"urls"`
	Username   string   `json:"username"`
	Credential string   `json:"credential"`
}

// InitiateCallRequest represents a request to initiate a call
type InitiateCallRequest struct {
	InitiatorID  string   `json:"initiator_id"`
	RecipientIDs []string `json:"recipient_ids"`
	CallType     string   `json:"call_type"` // "voice" or "video"
}

// InitiateCallResponse represents the response for initiating a call
type InitiateCallResponse struct {
	CallID       string        `json:"call_id"`
	CallType     string        `json:"call_type"`
	Participants []CallParticipant `json:"participants"`
	WebRTCConfig *WebRTCConfig `json:"webrtc_config"`
	Status       string        `json:"status"`
}

// JoinCallRequest represents a request to join a call
type JoinCallRequest struct {
	CallID   string `json:"call_id"`
	UserID   string `json:"user_id"`
	PeerID   string `json:"peer_id"`
}

// JoinCallResponse represents the response for joining a call
type JoinCallResponse struct {
	CallID       string        `json:"call_id"`
	Status       string        `json:"status"`
	Participants []CallParticipant `json:"participants"`
	WebRTCConfig *WebRTCConfig `json:"webrtc_config"`
}

// EndCallRequest represents a request to end a call
type EndCallRequest struct {
	CallID string `json:"call_id"`
	UserID string `json:"user_id"`
}

// EndCallResponse represents the response for ending a call
type EndCallResponse struct {
	CallID    string     `json:"call_id"`
	Status    string     `json:"status"`
	EndedAt   time.Time  `json:"ended_at"`
	Duration  int64      `json:"duration"` // in seconds
}

// InitiateCall initiates a new call between users
func (cs *CallService) InitiateCall(req *InitiateCallRequest) (*InitiateCallResponse, error) {
	// Validate call type
	if req.CallType != "voice" && req.CallType != "video" {
		return nil, errors.New("call type must be 'voice' or 'video'")
	}

	// Convert initiator ID to peer.ID
	initiatorPeerID, err := peer.Decode(req.InitiatorID)
	if err != nil {
		return nil, fmt.Errorf("invalid initiator ID: %w", err)
	}
	
	// Validate initiator exists
	initiatorInfo, err := cs.userManager.GetUserByPeerID(initiatorPeerID)
	if err != nil {
		return nil, fmt.Errorf("initiator validation failed: %w", err)
	}

	// Validate recipients exist and are online
	var participants []CallParticipant
	for _, recipientID := range req.RecipientIDs {
		// Convert recipient ID to peer.ID
		recipientPeerID, err := peer.Decode(recipientID)
		if err != nil {
			return nil, fmt.Errorf("invalid recipient ID %s: %w", recipientID, err)
		}
		
		recipientInfo, err := cs.userManager.GetUserByPeerID(recipientPeerID)
		if err != nil {
			return nil, fmt.Errorf("recipient validation failed for %s: %w", recipientID, err)
		}

		if !recipientInfo.IsOnline {
			return nil, fmt.Errorf("recipient %s is offline", recipientID)
		}

		participants = append(participants, CallParticipant{
			UserID:      recipientInfo.UserID,
			Username:    recipientInfo.Username,
			DisplayName: recipientInfo.DisplayName,
			PeerID:      recipientInfo.PeerID.String(),
			IsOnline:    recipientInfo.IsOnline,
			IsMuted:     false,
			IsVideoOn:   req.CallType == "video",
			JoinedAt:    time.Now(),
		})
	}

	// Add initiator to participants
	participants = append(participants, CallParticipant{
		UserID:      initiatorInfo.UserID,
		Username:    initiatorInfo.Username,
		DisplayName: initiatorInfo.DisplayName,
		PeerID:      initiatorInfo.PeerID.String(),
		IsOnline:    initiatorInfo.IsOnline,
		IsMuted:     false,
		IsVideoOn:   req.CallType == "video",
		JoinedAt:    time.Now(),
	})

	// Generate call ID
	callID := fmt.Sprintf("call_%s_%d", req.InitiatorID, time.Now().Unix())

	// Create call session
	callSession := &CallSession{
		CallID:       callID,
		CallType:     req.CallType,
		InitiatorID:  req.InitiatorID,
		Participants: participants,
		Status:       "ringing",
		StartedAt:    time.Now(),
		WebRTCConfig: cs.getWebRTCConfig(),
	}

	// Store call session
	cs.mutex.Lock()
	cs.activeCalls[callID] = callSession
	cs.mutex.Unlock()

	// Notify participants about the incoming call
	// In a real implementation, this would send WebRTC signaling messages

	return &InitiateCallResponse{
		CallID:       callID,
		CallType:     req.CallType,
		Participants: participants,
		WebRTCConfig: callSession.WebRTCConfig,
		Status:       "ringing",
	}, nil
}

// JoinCall allows a user to join an existing call
func (cs *CallService) JoinCall(req *JoinCallRequest) (*JoinCallResponse, error) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	// Get call session
	callSession, exists := cs.activeCalls[req.CallID]
	if !exists {
		return nil, errors.New("call not found")
	}

	// Check if call is still active
	if callSession.Status != "ringing" && callSession.Status != "active" {
		return nil, errors.New("call is not active")
	}

	// Convert user ID to peer.ID
	userPeerID, err := peer.Decode(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}
	
	// Validate user exists
	userInfo, err := cs.userManager.GetUserByPeerID(userPeerID)
	if err != nil {
		return nil, fmt.Errorf("user validation failed: %w", err)
	}

	// Check if user is already in the call
	for _, participant := range callSession.Participants {
		if participant.UserID == userInfo.UserID {
			// User is already in the call, return current status
			return &JoinCallResponse{
				CallID:       req.CallID,
				Status:       callSession.Status,
				Participants: callSession.Participants,
				WebRTCConfig: callSession.WebRTCConfig,
			}, nil
		}
	}

	// Add user to call
	newParticipant := CallParticipant{
		UserID:      userInfo.UserID,
		Username:    userInfo.Username,
		DisplayName: userInfo.DisplayName,
		PeerID:      userInfo.PeerID.String(),
		IsOnline:    userInfo.IsOnline,
		IsMuted:     false,
		IsVideoOn:   callSession.CallType == "video",
		JoinedAt:    time.Now(),
	}

	callSession.Participants = append(callSession.Participants, newParticipant)
	callSession.Status = "active"

	return &JoinCallResponse{
		CallID:       req.CallID,
		Status:       callSession.Status,
		Participants: callSession.Participants,
		WebRTCConfig: callSession.WebRTCConfig,
	}, nil
}

// EndCall ends an active call
func (cs *CallService) EndCall(req *EndCallRequest) (*EndCallResponse, error) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	// Get call session
	callSession, exists := cs.activeCalls[req.CallID]
	if !exists {
		return nil, errors.New("call not found")
	}

	// Check if user is authorized to end the call
	if callSession.InitiatorID != req.UserID {
		// Check if user is a participant
		isParticipant := false
		for _, participant := range callSession.Participants {
			if participant.UserID == req.UserID {
				isParticipant = true
				break
			}
		}
		if !isParticipant {
			return nil, errors.New("unauthorized to end call")
		}
	}

	// End the call
	now := time.Now()
	callSession.Status = "ended"
	callSession.EndedAt = &now

	// Calculate duration
	duration := now.Sub(callSession.StartedAt).Seconds()

	// Remove from active calls
	delete(cs.activeCalls, req.CallID)

	return &EndCallResponse{
		CallID:   req.CallID,
		Status:   "ended",
		EndedAt:  now,
		Duration: int64(duration),
	}, nil
}

// GetCallSession gets the current status of a call
func (cs *CallService) GetCallSession(callID string) (*CallSession, error) {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	callSession, exists := cs.activeCalls[callID]
	if !exists {
		return nil, errors.New("call not found")
	}

	return callSession, nil
}

// GetUserActiveCalls gets all active calls for a user
func (cs *CallService) GetUserActiveCalls(userID string) ([]*CallSession, error) {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	var userCalls []*CallSession
	for _, callSession := range cs.activeCalls {
		// Check if user is a participant
		for _, participant := range callSession.Participants {
			if participant.UserID == userID {
				userCalls = append(userCalls, callSession)
				break
			}
		}
	}

	return userCalls, nil
}

// getWebRTCConfig returns WebRTC configuration
func (cs *CallService) getWebRTCConfig() *WebRTCConfig {
	return &WebRTCConfig{
		ICEServers: []ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
			{
				URLs: []string{"stun:stun1.l.google.com:19302"},
			},
		},
		STUNServers: []string{
			"stun:stun.l.google.com:19302",
			"stun:stun1.l.google.com:19302",
		},
		TURNServers: []TURNServer{
			// In a real implementation, you'd have actual TURN servers
		},
	}
}
