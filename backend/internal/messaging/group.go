package messaging

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
)

// GroupManager manages group encryption using Sender Keys protocol
type GroupManager struct {
	host     host.Host
	ctx      context.Context
	groups   map[string]*Group
	mu       sync.RWMutex
	pubsub   *PubSub
	messages map[string][]string // groupID -> messages
}

// Group represents a group with its encryption state
type Group struct {
	ID         string
	Members    map[string]bool
	SenderKeys map[string][]byte // memberID -> sender key
	GroupKey   []byte
	KeyVersion int
	mu         sync.RWMutex
}

// GroupMessage represents an encrypted group message
type GroupMessage struct {
	GroupID    string `json:"group_id"`
	SenderID   string `json:"sender_id"`
	KeyVersion int    `json:"key_version"`
	Data       []byte `json:"data"`
	Timestamp  int64  `json:"timestamp"`
}

// NewGroupManager creates a new group manager
func NewGroupManager(ctx context.Context, h host.Host) (*GroupManager, error) {
	// Create pubsub instance
	ps, err := NewPubSub(ctx, h)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub: %w", err)
	}

	return &GroupManager{
		host:     h,
		ctx:      ctx,
		groups:   make(map[string]*Group),
		pubsub:   ps,
		messages: make(map[string][]string),
	}, nil
}

// CreateGroup creates a new group with the specified members
func (gm *GroupManager) CreateGroup(groupID string, memberIDs []string) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	// Check if group already exists
	if _, exists := gm.groups[groupID]; exists {
		return fmt.Errorf("group %s already exists", groupID)
	}

	// Create group
	group := &Group{
		ID:         groupID,
		Members:    make(map[string]bool),
		SenderKeys: make(map[string][]byte),
		KeyVersion: 1,
	}

	// Add all members
	for _, memberID := range memberIDs {
		group.Members[memberID] = true
		// Generate sender key for each member
		senderKey := make([]byte, 32)
		rand.Read(senderKey)
		group.SenderKeys[memberID] = senderKey
	}

	// Generate initial group key
	group.GroupKey = make([]byte, 32)
	rand.Read(group.GroupKey)

	gm.groups[groupID] = group

	// Subscribe to group topic
	topic := fmt.Sprintf("group-%s", groupID)
	_, err := gm.pubsub.Subscribe(topic)
	if err != nil {
		return fmt.Errorf("failed to subscribe to group topic: %w", err)
	}

	// Set up message handler for group messages
	gm.pubsub.SetMessageHandler(func(topicName string, data []byte) {
		if topicName == topic {
			gm.handleGroupMessage(groupID, data)
		}
	})

	// Initialize messages storage for this group
	gm.messages[groupID] = make([]string, 0)

	return nil
}

// EncryptGroupMessage encrypts a message for the group
func (gm *GroupManager) EncryptGroupMessage(groupID string, plaintext []byte) ([]byte, error) {
	gm.mu.RLock()
	group, exists := gm.groups[groupID]
	gm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("group %s not found", groupID)
	}

	// Check if current user is member
	myID := gm.host.ID().String()
	if !group.Members[myID] {
		return nil, fmt.Errorf("not a member of group %s", groupID)
	}

	// Create group message
	msg := &GroupMessage{
		GroupID:    groupID,
		SenderID:   myID,
		KeyVersion: group.KeyVersion,
		Data:       plaintext, // For now, send plaintext (will be encrypted in real implementation)
		Timestamp:  time.Now().Unix(),
	}

	// Serialize message
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize message: %w", err)
	}

	// Publish to group topic
	topic := fmt.Sprintf("group-%s", groupID)
	err = gm.pubsub.Publish(topic, msgBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to publish message: %w", err)
	}

	return msgBytes, nil
}

// DecryptGroupMessage decrypts a group message
func (gm *GroupManager) DecryptGroupMessage(groupID string, encryptedMsg []byte) ([]byte, error) {
	gm.mu.RLock()
	group, exists := gm.groups[groupID]
	gm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("group %s not found", groupID)
	}

	// Check if current user is member
	myID := gm.host.ID().String()
	if !group.Members[myID] {
		return nil, fmt.Errorf("not a member of group %s", groupID)
	}

	// Deserialize message
	var msg GroupMessage
	err := json.Unmarshal(encryptedMsg, &msg)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize message: %w", err)
	}

	// Verify group ID
	if msg.GroupID != groupID {
		return nil, fmt.Errorf("message group ID mismatch")
	}

	// Check if sender is member
	if !group.Members[msg.SenderID] {
		return nil, fmt.Errorf("sender %s is not a member of group %s", msg.SenderID, groupID)
	}

	// For now, return the data as-is (no encryption in this simplified version)
	return msg.Data, nil
}

// RemoveMember removes a member from the group and rotates keys
func (gm *GroupManager) RemoveMember(groupID string, memberID string) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	group, exists := gm.groups[groupID]
	if !exists {
		return fmt.Errorf("group %s not found", groupID)
	}

	// Check if member exists
	if !group.Members[memberID] {
		return fmt.Errorf("member %s not in group %s", memberID, groupID)
	}

	// Remove member
	delete(group.Members, memberID)
	delete(group.SenderKeys, memberID)

	// Rotate group key for forward secrecy
	group.GroupKey = make([]byte, 32)
	rand.Read(group.GroupKey)
	group.KeyVersion++

	// Publish key rotation message
	rotationMsg := map[string]interface{}{
		"type":        "key_rotation",
		"group_id":    groupID,
		"key_version": group.KeyVersion,
		"timestamp":   time.Now().Unix(),
	}

	rotationBytes, _ := json.Marshal(rotationMsg)
	topic := fmt.Sprintf("group-%s", groupID)
	gm.pubsub.Publish(topic, rotationBytes)

	return nil
}

// UpdateGroupFromRotation handles key rotation messages
func (gm *GroupManager) UpdateGroupFromRotation(groupID string, keyVersion int) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	group, exists := gm.groups[groupID]
	if !exists {
		return
	}

	// Update key version if it's newer
	if keyVersion > group.KeyVersion {
		group.KeyVersion = keyVersion
		// Generate new group key
		group.GroupKey = make([]byte, 32)
		rand.Read(group.GroupKey)
	}
}

// RemoveGroupForTesting removes a group (for testing purposes)
func (gm *GroupManager) RemoveGroupForTesting(groupID string) {
	gm.mu.Lock()
	defer gm.mu.Unlock()
	delete(gm.groups, groupID)
}

// IsGroupMember checks if a member is in the group
func (gm *GroupManager) IsGroupMember(groupID string, memberID string) bool {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	group, exists := gm.groups[groupID]
	if !exists {
		return false
	}

	return group.Members[memberID]
}

// SendGroupMessage sends a message to the group
func (gm *GroupManager) SendGroupMessage(groupID string, message string) error {
	_, err := gm.EncryptGroupMessage(groupID, []byte(message))
	return err
}

// GetGroupMessages returns messages for a group
func (gm *GroupManager) GetGroupMessages(groupID string) []string {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	messages, exists := gm.messages[groupID]
	if !exists {
		return []string{}
	}

	return messages
}

// AddMember adds a member to the group
func (gm *GroupManager) AddMember(groupID string, memberID string) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	group, exists := gm.groups[groupID]
	if !exists {
		return fmt.Errorf("group %s not found", groupID)
	}

	// Add member
	group.Members[memberID] = true

	// Generate sender key for new member
	senderKey := make([]byte, 32)
	rand.Read(senderKey)
	group.SenderKeys[memberID] = senderKey

	// Publish member addition message
	additionMsg := map[string]interface{}{
		"type":      "member_added",
		"group_id":  groupID,
		"member_id": memberID,
		"timestamp": time.Now().Unix(),
	}

	additionBytes, _ := json.Marshal(additionMsg)
	topic := fmt.Sprintf("group-%s", groupID)
	gm.pubsub.Publish(topic, additionBytes)

	return nil
}

// LeaveGroup removes the current user from the group
func (gm *GroupManager) LeaveGroup(groupID string) error {
	myID := gm.host.ID().String()
	return gm.RemoveMember(groupID, myID)
}

// handleGroupMessage handles incoming group messages
func (gm *GroupManager) handleGroupMessage(groupID string, data []byte) {
	// Deserialize message
	var msg GroupMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return
	}

	// Store message
	gm.mu.Lock()
	if gm.messages[groupID] == nil {
		gm.messages[groupID] = make([]string, 0)
	}
	gm.messages[groupID] = append(gm.messages[groupID], string(msg.Data))
	gm.mu.Unlock()
}

// encryptWithGroupKey encrypts data using the group key (simplified)
func (gm *GroupManager) encryptWithGroupKey(groupKey []byte, data []byte) ([]byte, error) {
	// Simplified encryption using XOR (in real implementation, use proper encryption)
	encrypted := make([]byte, len(data))
	keyHash := sha256.Sum256(groupKey)

	for i := 0; i < len(data); i++ {
		encrypted[i] = data[i] ^ keyHash[i%32]
	}

	return encrypted, nil
}

// decryptWithGroupKey decrypts data using the group key (simplified)
func (gm *GroupManager) decryptWithGroupKey(groupKey []byte, data []byte) ([]byte, error) {
	// XOR is symmetric, so decryption is the same as encryption
	return gm.encryptWithGroupKey(groupKey, data)
}
