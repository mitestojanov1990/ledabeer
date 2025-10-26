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

	// Create call session with real WebRTC peer connection
	session, err := NewCallSessionWithWebRTC(audioEnabled, videoEnabled)
	if err != nil {
		return "", fmt.Errorf("failed to create WebRTC session: %w", err)
	}

	cm.mutex.Lock()
	cm.sessions[callID] = session
	cm.mutex.Unlock()

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

// CreateCallSession creates a new call session for testing
func (cm *CallManager) CreateCallSession(callID, peerID string) *CallSession {
	session := &CallSession{
		participants: make(map[string]bool),
		state:        StateInitiating,
		rtpHandler:   &RTPHandler{},
	}

	cm.mutex.Lock()
	cm.sessions[callID] = session
	cm.mutex.Unlock()

	return session
}

func (cm *CallManager) HandleSignaling(ctx context.Context, callID, msgType, sdp, candidate string) (*pb.SignalingMessage, error) {
	cm.mutex.RLock()
	session, exists := cm.sessions[callID]
	cm.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("call session %s not found", callID)
	}

	switch msgType {
	case "offer":
		// Process real SDP offer
		answer, err := session.CreateAnswer(&SDP{Type: "offer", SDP: sdp})
		if err != nil {
			return nil, fmt.Errorf("failed to create answer: %w", err)
		}
		return &pb.SignalingMessage{
			Type: "answer",
			Sdp:  answer.SDP,
		}, nil

	case "answer":
		// Process SDP answer
		err := session.SetRemoteDescription(&SDP{Type: "answer", SDP: sdp})
		if err != nil {
			return nil, fmt.Errorf("failed to set remote description: %w", err)
		}
		return nil, nil

	case "ice-candidate":
		// Process real ICE candidate
		err := session.AddICECandidate(candidate)
		if err != nil {
			return nil, fmt.Errorf("failed to add ICE candidate: %w", err)
		}

		// Return our own ICE candidate
		candidates, err := session.GatherCandidates()
		if err != nil {
			return nil, fmt.Errorf("failed to gather ICE candidates: %w", err)
		}

		if len(candidates) > 0 {
			// For now, return a dummy candidate
			return &pb.SignalingMessage{
				Type:      "ice-candidate",
				Candidate: "candidate:1 1 UDP 2113667326 192.168.1.100 54400 typ host",
			}, nil
		}
		return nil, nil

	default:
		return nil, fmt.Errorf("unknown message type: %s", msgType)
	}
}
