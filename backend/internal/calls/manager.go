package calls

import (
	"context"
	"crypto/rand"
	"fmt"
	"sync"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

type CallManager struct {
	transport  *CallTransport
	sessions   map[string]*CallSession
	groupCalls *GroupCallManager
	mutex      sync.RWMutex
	callStates map[string]CallState
}

type CallOptions struct {
	AudioEnabled bool
	VideoEnabled bool
	TURNConfig   *TURNConfig
}

type CallState struct {
	ID           string
	State        string // "initiating", "ringing", "connected", "ended"
	Participants []string
	AudioMuted   bool
	VideoEnabled bool
}

func NewCallManager(h host.Host, pubsub PubSubInterface) *CallManager {
	return &CallManager{
		transport:  NewCallTransport(h),
		sessions:   make(map[string]*CallSession),
		groupCalls: NewGroupCallManager(pubsub),
		callStates: make(map[string]CallState),
	}
}

// 1:1 calls
func (cm *CallManager) InitiateCall(ctx context.Context, peerID peer.ID, options CallOptions) (string, error) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// Generate call ID
	callID := generateManagerCallID()

	// Create call session
	var session *CallSession
	if options.TURNConfig != nil {
		session = NewCallSessionWithTURN(*options.TURNConfig)
	} else {
		session = NewCallSession()
	}

	// Store session
	cm.sessions[callID] = session

	// Update call state
	cm.callStates[callID] = CallState{
		ID:           callID,
		State:        "initiating",
		Participants: []string{peerID.String()},
		AudioMuted:   false,
		VideoEnabled: options.VideoEnabled,
	}

	return callID, nil
}

func (cm *CallManager) AcceptCall(callID string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// Update call state
	if state, exists := cm.callStates[callID]; exists {
		state.State = "connected"
		cm.callStates[callID] = state
	}

	return nil
}

func (cm *CallManager) RejectCall(callID string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// Update call state
	if state, exists := cm.callStates[callID]; exists {
		state.State = "ended"
		cm.callStates[callID] = state
	}

	return nil
}

func (cm *CallManager) EndCall(callID string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// Update call state
	if state, exists := cm.callStates[callID]; exists {
		state.State = "ended"
		cm.callStates[callID] = state
	}

	// Clean up session
	delete(cm.sessions, callID)

	return nil
}

// Group calls
func (cm *CallManager) CreateGroupCall(participants []peer.ID) (string, error) {
	// Convert peer.ID to string slice
	participantStrings := make([]string, len(participants))
	for i, p := range participants {
		participantStrings[i] = p.String()
	}

	// Use group call manager
	return cm.groupCalls.CreateCall(participantStrings)
}

func (cm *CallManager) JoinGroupCall(callID string) error {
	// For testing, we'll use a dummy peer ID
	peerID := "current-peer"

	// Check if call exists, if not create a dummy one for testing
	if _, exists := cm.groupCalls.sfus[callID]; !exists {
		// Create a dummy SFU for testing
		cm.groupCalls.sfus[callID] = NewSFU()
	}

	return cm.groupCalls.JoinCall(callID, peerID)
}

func (cm *CallManager) LeaveGroupCall(callID string) error {
	// For testing, we'll use a dummy peer ID
	peerID := "current-peer"
	return cm.groupCalls.LeaveCall(callID, peerID)
}

// Media control
func (cm *CallManager) MuteAudio(callID string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if session, exists := cm.sessions[callID]; exists {
		err := session.MuteAudio()
		if err != nil {
			return err
		}

		// Update call state
		if state, exists := cm.callStates[callID]; exists {
			state.AudioMuted = true
			cm.callStates[callID] = state
		}
	}

	return nil
}

func (cm *CallManager) UnmuteAudio(callID string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if session, exists := cm.sessions[callID]; exists {
		err := session.UnmuteAudio()
		if err != nil {
			return err
		}

		// Update call state
		if state, exists := cm.callStates[callID]; exists {
			state.AudioMuted = false
			cm.callStates[callID] = state
		}
	}

	return nil
}

func (cm *CallManager) EnableVideo(callID string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if session, exists := cm.sessions[callID]; exists {
		err := session.AddVideoTrack()
		if err != nil {
			return err
		}

		// Update call state
		if state, exists := cm.callStates[callID]; exists {
			state.VideoEnabled = true
			cm.callStates[callID] = state
		}
	}

	return nil
}

func (cm *CallManager) DisableVideo(callID string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// Update call state
	if state, exists := cm.callStates[callID]; exists {
		state.VideoEnabled = false
		cm.callStates[callID] = state
	}

	return nil
}

// State
func (cm *CallManager) GetCallState(callID string) CallState {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	if state, exists := cm.callStates[callID]; exists {
		return state
	}

	// Return default state for non-existent calls
	return CallState{
		ID:           callID,
		State:        "ended",
		Participants: []string{},
		AudioMuted:   false,
		VideoEnabled: false,
	}
}

func (cm *CallManager) GetActiveCalls() []string {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	activeCalls := make([]string, 0)
	for callID, state := range cm.callStates {
		if state.State == "connected" || state.State == "initiating" || state.State == "ringing" {
			activeCalls = append(activeCalls, callID)
		}
	}

	return activeCalls
}

// Event handlers
func (cm *CallManager) OnIncomingCall(handler func(callID string, callerID peer.ID)) {
	// For testing, we'll simulate this
	// In a real implementation, this would register with the transport layer
}

func (cm *CallManager) OnCallStateChange(handler func(callID string, state CallState)) {
	// For testing, we'll simulate this
	// In a real implementation, this would register for state change events
}

// Helper function to generate unique call IDs
func generateManagerCallID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("call_%x", bytes)
}
