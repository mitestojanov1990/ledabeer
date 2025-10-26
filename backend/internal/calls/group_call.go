package calls

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
)

type GroupCallManager struct {
	pubsub PubSubInterface // Interface for pubsub
	sfus   map[string]*SFU
	mutex  sync.RWMutex
	events chan CallEvent
}

type CallEvent struct {
	Type         string // "participant_joined", "participant_left", "call_created"
	CallID       string
	Participant  string
	Participants []string
}

type PubSubInterface interface {
	Publish(topic string, data []byte) error
	Subscribe(topic string, handler func([]byte)) error
}

func NewGroupCallManager(ps PubSubInterface) *GroupCallManager {
	return &GroupCallManager{
		pubsub: ps,
		sfus:   make(map[string]*SFU),
		events: make(chan CallEvent, 10),
	}
}

func (gcm *GroupCallManager) CreateCall(participants []string) (string, error) {
	// Generate call ID
	callID := generateCallID()

	// Create SFU
	sfu := NewSFU()
	gcm.mutex.Lock()
	gcm.sfus[callID] = sfu
	gcm.mutex.Unlock()

	// Publish call_created event to group topic
	if gcm.pubsub != nil {
		event := CallEvent{
			Type:         "call_created",
			CallID:       callID,
			Participants: participants,
		}
		gcm.publishEvent(event)
	}

	return callID, nil
}

func (gcm *GroupCallManager) JoinCall(callID, peerID string) error {
	gcm.mutex.Lock()
	defer gcm.mutex.Unlock()

	// Add to SFU
	sfu, exists := gcm.sfus[callID]
	if !exists {
		return fmt.Errorf("call %s not found", callID)
	}

	sfu.AddParticipant(peerID)

	// Broadcast participant_joined
	event := CallEvent{
		Type:        "participant_joined",
		CallID:      callID,
		Participant: peerID,
	}
	gcm.publishEvent(event)

	return nil
}

func (gcm *GroupCallManager) LeaveCall(callID, peerID string) error {
	gcm.mutex.Lock()
	defer gcm.mutex.Unlock()

	// Remove from SFU
	sfu, exists := gcm.sfus[callID]
	if !exists {
		return fmt.Errorf("call %s not found", callID)
	}

	sfu.RemoveParticipant(peerID)

	// Broadcast participant_left
	event := CallEvent{
		Type:        "participant_left",
		CallID:      callID,
		Participant: peerID,
	}
	gcm.publishEvent(event)

	return nil
}

func (gcm *GroupCallManager) GetParticipants(callID string) []string {
	gcm.mutex.RLock()
	defer gcm.mutex.RUnlock()

	sfu, exists := gcm.sfus[callID]
	if !exists {
		return nil
	}

	// Return participant IDs
	var participants []string
	for peerID := range sfu.participants {
		participants = append(participants, peerID)
	}

	return participants
}

func (gcm *GroupCallManager) OnEvent(handler func(CallEvent)) {
	go func() {
		for event := range gcm.events {
			handler(event)
		}
	}()
}

func (gcm *GroupCallManager) publishEvent(event CallEvent) {
	select {
	case gcm.events <- event:
	default:
		// Channel full, drop event
	}
}

func generateCallID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
