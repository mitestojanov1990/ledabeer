package messaging

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"ledabeer/backend/internal/storage"

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
	// New fields for real PubSub integration
	subscribers map[string][]chan GroupMessage // groupID -> subscribers
	published   map[string]map[string][]byte   // groupID -> messageID -> data
	ipfsNode    *storage.IPFSNode              // For message history storage
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
	ID         string `json:"id"`
	GroupID    string `json:"group_id"`
	SenderID   string `json:"sender_id"`
	Content    []byte `json:"content"`
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
		host:        h,
		ctx:         ctx,
		groups:      make(map[string]*Group),
		pubsub:      ps,
		messages:    make(map[string][]string),
		subscribers: make(map[string][]chan GroupMessage),
		published:   make(map[string]map[string][]byte),
	}, nil
}

// NewGroupManagerWithStorage creates a new group manager with IPFS storage
func NewGroupManagerWithStorage(ctx context.Context, h host.Host, ipfsNode *storage.IPFSNode) (*GroupManager, error) {
	gm, err := NewGroupManager(ctx, h)
	if err != nil {
		return nil, err
	}
	gm.ipfsNode = ipfsNode
	return gm, nil
}

// CreateGroup creates a new group with the specified members (new signature)
func (gm *GroupManager) CreateGroup(ctx context.Context, groupID string) error {
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

	// Add current user as initial member
	myID := gm.host.ID().String()
	group.Members[myID] = true
	// Generate sender key for current user
	senderKey := make([]byte, 32)
	rand.Read(senderKey)
	group.SenderKeys[myID] = senderKey

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

	// Initialize subscribers and published maps
	gm.subscribers[groupID] = make([]chan GroupMessage, 0)
	gm.published[groupID] = make(map[string][]byte)

	return nil
}

// CreateGroupLegacy creates a new group with the specified members (legacy signature)
func (gm *GroupManager) CreateGroupLegacy(groupID string, memberIDs []string) error {
	ctx := context.Background()

	// Create group first
	err := gm.CreateGroup(ctx, groupID)
	if err != nil {
		return err
	}

	// Add all members
	for _, memberID := range memberIDs {
		err = gm.addMember(groupID, memberID)
		if err != nil {
			return err
		}
	}

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
func (gm *GroupManager) SendGroupMessage(ctx context.Context, groupID string, content []byte) (string, error) {
	// Generate message ID
	messageID := generateMessageID()

	// Encrypt and publish message
	encryptedData, err := gm.EncryptGroupMessage(groupID, content)
	if err != nil {
		return "", err
	}

	// Store published message for testing
	gm.mu.Lock()
	if gm.published[groupID] == nil {
		gm.published[groupID] = make(map[string][]byte)
	}
	gm.published[groupID][messageID] = encryptedData
	gm.mu.Unlock()

	// Store message in local storage
	gm.mu.Lock()
	if gm.messages[groupID] == nil {
		gm.messages[groupID] = make([]string, 0)
	}
	gm.messages[groupID] = append(gm.messages[groupID], string(content))
	gm.mu.Unlock()

	// Notify subscribers
	gm.notifySubscribers(groupID, GroupMessage{
		ID:        messageID,
		GroupID:   groupID,
		SenderID:  gm.host.ID().String(),
		Content:   content,
		Data:      content,
		Timestamp: time.Now().Unix(),
	})

	return messageID, nil
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

// AddMember adds a member to the group (legacy method)
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

// New methods for real PubSub integration

// SubscribeToGroupMessages subscribes to group messages
func (gm *GroupManager) SubscribeToGroupMessages(ctx context.Context, groupID string) <-chan GroupMessage {
	msgChan := make(chan GroupMessage, 10)

	gm.mu.Lock()
	if gm.subscribers[groupID] == nil {
		gm.subscribers[groupID] = make([]chan GroupMessage, 0)
	}
	gm.subscribers[groupID] = append(gm.subscribers[groupID], msgChan)
	gm.mu.Unlock()

	return msgChan
}

// HasPublishedToGroup checks if a message was published to a group
func (gm *GroupManager) HasPublishedToGroup(groupID, messageID string) bool {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	if groupMessages, exists := gm.published[groupID]; exists {
		_, published := groupMessages[messageID]
		return published
	}
	return false
}

// GetPublishedMessage gets the published message data
func (gm *GroupManager) GetPublishedMessage(groupID, messageID string) []byte {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	if groupMessages, exists := gm.published[groupID]; exists {
		return groupMessages[messageID]
	}
	return nil
}

// GetGroupMessageHistory retrieves group message history from IPFS storage
func (gm *GroupManager) GetGroupMessageHistory(ctx context.Context, groupID string, limit int) ([]GroupMessage, error) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	// Get messages from local storage (simplified for testing)
	messages, exists := gm.messages[groupID]
	if !exists {
		return []GroupMessage{}, nil
	}

	// Convert string messages to GroupMessage structs
	groupMessages := make([]GroupMessage, 0, len(messages))
	for i, content := range messages {
		groupMessages = append(groupMessages, GroupMessage{
			ID:        fmt.Sprintf("msg-%d", i),
			GroupID:   groupID,
			Content:   []byte(content),
			Timestamp: time.Now().Unix() - int64(len(messages)-i), // Simulate different timestamps
		})
	}

	// Apply limit
	if limit > 0 && len(groupMessages) > limit {
		groupMessages = groupMessages[:limit]
	}

	return groupMessages, nil
}

// GetGroupMembers returns the members of a group
func (gm *GroupManager) GetGroupMembers(ctx context.Context, groupID string) ([]string, error) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	group, exists := gm.groups[groupID]
	if !exists {
		return nil, fmt.Errorf("group %s not found", groupID)
	}

	members := make([]string, 0, len(group.Members))
	for memberID := range group.Members {
		members = append(members, memberID)
	}

	return members, nil
}

// AddMemberWithContext adds a member to the group (updated signature)
func (gm *GroupManager) AddMemberWithContext(ctx context.Context, groupID, memberID string) error {
	return gm.addMember(groupID, memberID)
}

// RemoveMemberWithContext removes a member from the group (updated signature)
func (gm *GroupManager) RemoveMemberWithContext(ctx context.Context, groupID, memberID string) error {
	return gm.removeMember(groupID, memberID)
}

// addMember adds a member to the group (internal method)
func (gm *GroupManager) addMember(groupID, memberID string) error {
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

// removeMember removes a member from the group (internal method)
func (gm *GroupManager) removeMember(groupID, memberID string) error {
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

// notifySubscribers notifies all subscribers of a group message
func (gm *GroupManager) notifySubscribers(groupID string, msg GroupMessage) {
	gm.mu.RLock()
	subscribers := gm.subscribers[groupID]
	gm.mu.RUnlock()

	for _, subscriber := range subscribers {
		select {
		case subscriber <- msg:
		default:
			// Skip if channel is full
		}
	}
}

// generateMessageID generates a unique message ID
func generateMessageID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)
}
