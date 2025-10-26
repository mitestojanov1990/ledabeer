package calls

import (
	"context"
	"fmt"
	"sync"

	pb "ledabeer/backend/pkg/proto"

	"github.com/libp2p/go-libp2p/core/host"
)

type CallManager struct {
	host     host.Host
	sessions map[string]*CallSession
	mutex    sync.RWMutex
}

type CallState int

const (
	StateInitiating CallState = iota
	StateRinging
	StateConnected
	StateEnded
)

func NewCallManager(h host.Host) *CallManager {
	return &CallManager{
		host:     h,
		sessions: make(map[string]*CallSession),
	}
}

func (cm *CallManager) InitiateCall(ctx context.Context, toPeerID string, audioEnabled, videoEnabled bool) (string, error) {
	callID := generateCallID()

	// Create call session using existing CallSession
	session := NewCallSession()

	cm.mutex.Lock()
	cm.sessions[callID] = session
	cm.mutex.Unlock()

	// In real implementation, this would initiate WebRTC call
	// For now, just return the call ID
	return callID, nil
}

func (cm *CallManager) AnswerCall(ctx context.Context, callID string, accept bool) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	_, exists := cm.sessions[callID]
	if !exists {
		return fmt.Errorf("call session %s not found", callID)
	}

	// In real implementation, this would handle the call answer
	// For now, just return success
	return nil
}

func (cm *CallManager) EndCall(ctx context.Context, callID string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	_, exists := cm.sessions[callID]
	if !exists {
		return fmt.Errorf("call session %s not found", callID)
	}

	// Clean up session
	delete(cm.sessions, callID)

	return nil
}

func (cm *CallManager) GetCallSession(callID string) *CallSession {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.sessions[callID]
}

func (cm *CallManager) HandleSignaling(ctx context.Context, callID, msgType, sdp, candidate string) (*pb.SignalingMessage, error) {
	// Handle WebRTC signaling
	// In real implementation, this would process SDP offers/answers and ICE candidates

	switch msgType {
	case "offer":
		// Process SDP offer
		return &pb.SignalingMessage{
			Type: "answer",
			Sdp:  "mock-sdp-answer",
		}, nil
	case "answer":
		// Process SDP answer
		return nil, nil
	case "ice-candidate":
		// Process ICE candidate
		return &pb.SignalingMessage{
			Type:      "ice-candidate",
			Candidate: "mock-ice-candidate",
		}, nil
	default:
		return nil, fmt.Errorf("unknown message type: %s", msgType)
	}
}

// Add methods to CallSession for compatibility
func (cs *CallSession) GetState() CallState {
	// For now, return a default state
	// In real implementation, this would track the actual call state
	return StateConnected
}

func (cs *CallSession) GetParticipants() []string {
	// For now, return empty participants
	// In real implementation, this would track actual participants
	return []string{}
}
