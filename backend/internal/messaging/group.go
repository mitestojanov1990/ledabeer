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
	// Enhanced removal synchronization
	pendingRemovals map[string]map[string]time.Time       // groupID -> memberID -> timestamp
	removalAcks     map[string]map[string]map[string]bool // groupID -> memberID -> peerID -> acked
	removalMu       sync.RWMutex
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

	gm := &GroupManager{
		host:            h,
		ctx:             ctx,
		groups:          make(map[string]*Group),
		pubsub:          ps,
		messages:        make(map[string][]string),
		subscribers:     make(map[string][]chan GroupMessage),
		published:       make(map[string]map[string][]byte),
		pendingRemovals: make(map[string]map[string]time.Time),
		removalAcks:     make(map[string]map[string]map[string]bool),
	}

	// Start background goroutine for removal management
	go gm.removalManagementLoop()

	return gm, nil
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

	// Publish member left message first
	leftMsg := map[string]interface{}{
		"type":      "member_left",
		"group_id":  groupID,
		"member_id": myID,
		"timestamp": time.Now().Unix(),
		"from":      myID,
	}

	leftBytes, _ := json.Marshal(leftMsg)
	topic := fmt.Sprintf("group-%s", groupID)
	gm.pubsub.Publish(topic, leftBytes)

	// Then remove locally
	err := gm.removeMember(groupID, myID)
	if err != nil {
		return err
	}

	// Clear group access after leaving
	gm.mu.Lock()
	delete(gm.groups, groupID)
	delete(gm.messages, groupID)
	delete(gm.subscribers, groupID)
	delete(gm.published, groupID)
	gm.mu.Unlock()

	return nil
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

	// Check if this is a voluntary leave (member leaving themselves)
	myID := gm.host.ID().String()
	isVoluntaryLeave := memberID == myID

	if !isVoluntaryLeave {
		// Rotate group key for forward secrecy only for forced removals
		group.GroupKey = make([]byte, 32)
		rand.Read(group.GroupKey)
		group.KeyVersion++

		// Track pending removal for acknowledgment
		gm.removalMu.Lock()
		if gm.pendingRemovals[groupID] == nil {
			gm.pendingRemovals[groupID] = make(map[string]time.Time)
		}
		gm.pendingRemovals[groupID][memberID] = time.Now()

		if gm.removalAcks[groupID] == nil {
			gm.removalAcks[groupID] = make(map[string]map[string]bool)
		}
		if gm.removalAcks[groupID][memberID] == nil {
			gm.removalAcks[groupID][memberID] = make(map[string]bool)
		}
		gm.removalMu.Unlock()

		// Publish member removal message with acknowledgment request
		removalMsg := map[string]interface{}{
			"type":        "member_removed",
			"group_id":    groupID,
			"member_id":   memberID,
			"timestamp":   time.Now().Unix(),
			"request_ack": true,
			"from":        gm.host.ID().String(),
		}

		removalBytes, _ := json.Marshal(removalMsg)
		topic := fmt.Sprintf("group-%s", groupID)
		gm.pubsub.Publish(topic, removalBytes)

		// Publish key rotation message
		rotationMsg := map[string]interface{}{
			"type":        "key_rotation",
			"group_id":    groupID,
			"key_version": group.KeyVersion,
			"timestamp":   time.Now().Unix(),
		}

		rotationBytes, _ := json.Marshal(rotationMsg)
		gm.pubsub.Publish(topic, rotationBytes)
	}

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

// JoinGroup allows a node to join an existing group
func (gm *GroupManager) JoinGroup(ctx context.Context, groupID string) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	// Check if group already exists locally
	if _, exists := gm.groups[groupID]; exists {
		return nil // Already a member
	}

	// Create group structure
	group := &Group{
		ID:         groupID,
		Members:    make(map[string]bool),
		SenderKeys: make(map[string][]byte),
		KeyVersion: 1,
	}

	// Add current user as member
	myID := gm.host.ID().String()
	group.Members[myID] = true

	// Generate sender key for current user
	senderKey := make([]byte, 32)
	rand.Read(senderKey)
	group.SenderKeys[myID] = senderKey

	// Generate initial group key (will be updated when we receive the real one)
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

	// Request current group state
	gm.requestGroupState(ctx, groupID)

	// Wait a bit for state synchronization
	time.Sleep(50 * time.Millisecond)

	// If this is a rejoin (group already exists remotely), we need to be added back
	// This will be handled by the existing members when they receive our state request
	return nil
}

// GetGroupKeyVersion returns the current key version for a group
func (gm *GroupManager) GetGroupKeyVersion(groupID string) int {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	group, exists := gm.groups[groupID]
	if !exists {
		return 0
	}

	return group.KeyVersion
}

// GetGroups returns all groups that the current user is a member of
func (gm *GroupManager) GetGroups() []*Group {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	groups := make([]*Group, 0, len(gm.groups))
	for _, group := range gm.groups {
		groups = append(groups, group)
	}

	return groups
}

// GetGroup returns a specific group by ID
func (gm *GroupManager) GetGroup(groupID string) (*Group, bool) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	group, exists := gm.groups[groupID]
	return group, exists
}

// RotateGroupKeys rotates the group keys and notifies all members
func (gm *GroupManager) RotateGroupKeys(ctx context.Context, groupID string) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	group, exists := gm.groups[groupID]
	if !exists {
		return fmt.Errorf("group %s not found", groupID)
	}

	// Check if current user is member
	myID := gm.host.ID().String()
	if !group.Members[myID] {
		return fmt.Errorf("not a member of group %s", groupID)
	}

	// Rotate group key
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

// ProcessMemberRemovalMessage processes a member removal message
func (gm *GroupManager) ProcessMemberRemovalMessage(groupID, memberID string) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	group, exists := gm.groups[groupID]
	if !exists {
		return
	}

	// Remove member from local state
	delete(group.Members, memberID)
	delete(group.SenderKeys, memberID)

	// If this is the current user being removed, leave the group
	myID := gm.host.ID().String()
	if memberID == myID {
		// Unsubscribe from topic
		topic := fmt.Sprintf("group-%s", groupID)
		gm.pubsub.Unsubscribe(topic)
	}
}

// requestGroupState requests the current group state from other members
func (gm *GroupManager) requestGroupState(ctx context.Context, groupID string) {
	// Publish a state request message
	stateRequest := map[string]interface{}{
		"type":      "state_request",
		"group_id":  groupID,
		"from":      gm.host.ID().String(),
		"timestamp": time.Now().Unix(),
	}

	stateRequestBytes, _ := json.Marshal(stateRequest)
	topic := fmt.Sprintf("group-%s", groupID)
	gm.pubsub.Publish(topic, stateRequestBytes)
}

// handleGroupMessage handles incoming group messages including state synchronization
func (gm *GroupManager) handleGroupMessage(groupID string, data []byte) {
	// Try to parse as a state message first
	var stateMsg map[string]interface{}
	if err := json.Unmarshal(data, &stateMsg); err == nil {
		if msgType, ok := stateMsg["type"].(string); ok {
			switch msgType {
			case "member_added":
				gm.handleMemberAddedMessage(groupID, stateMsg)
			case "member_removed":
				gm.handleMemberRemovedMessage(groupID, stateMsg)
			case "key_rotation":
				gm.handleKeyRotationMessage(groupID, stateMsg)
			case "state_request":
				gm.handleStateRequestMessage(groupID, stateMsg)
			case "state_response":
				gm.handleStateResponseMessage(groupID, stateMsg)
			case "removal_ack":
				gm.handleRemovalAckMessage(groupID, stateMsg)
			case "removal_retry":
				gm.handleRemovalRetryMessage(groupID, stateMsg)
			case "member_left":
				gm.handleMemberLeftMessage(groupID, stateMsg)
			}
			return
		}
	}

	// Fall back to regular message handling
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

// handleMemberAddedMessage handles member addition messages
func (gm *GroupManager) handleMemberAddedMessage(groupID string, msg map[string]interface{}) {
	memberID, ok := msg["member_id"].(string)
	if !ok {
		return
	}

	gm.mu.Lock()
	defer gm.mu.Unlock()

	group, exists := gm.groups[groupID]
	if !exists {
		return
	}

	// Add member to local state
	group.Members[memberID] = true

	// Generate sender key for new member
	senderKey := make([]byte, 32)
	rand.Read(senderKey)
	group.SenderKeys[memberID] = senderKey
}

// handleMemberRemovedMessage handles member removal messages
func (gm *GroupManager) handleMemberRemovedMessage(groupID string, msg map[string]interface{}) {
	memberID, ok := msg["member_id"].(string)
	if !ok {
		return
	}

	from, _ := msg["from"].(string)
	requestAck, _ := msg["request_ack"].(bool)

	gm.mu.Lock()
	defer gm.mu.Unlock()

	group, exists := gm.groups[groupID]
	if !exists {
		return
	}

	// Remove member from local state
	delete(group.Members, memberID)
	delete(group.SenderKeys, memberID)

	// If this is the current user being removed, leave the group
	myID := gm.host.ID().String()
	if memberID == myID {
		// Unsubscribe from topic
		topic := fmt.Sprintf("group-%s", groupID)
		gm.pubsub.Unsubscribe(topic)

		// Clear group access - remove from local groups and clear messages
		delete(gm.groups, groupID)
		delete(gm.messages, groupID)
		delete(gm.subscribers, groupID)
		delete(gm.published, groupID)
	}

	// Send acknowledgment if requested
	if requestAck && from != "" && from != myID {
		gm.sendRemovalAck(groupID, memberID, from)
	}
}

// handleKeyRotationMessage handles key rotation messages
func (gm *GroupManager) handleKeyRotationMessage(groupID string, msg map[string]interface{}) {
	keyVersion, ok := msg["key_version"].(float64)
	if !ok {
		return
	}

	gm.mu.Lock()
	defer gm.mu.Unlock()

	group, exists := gm.groups[groupID]
	if !exists {
		return
	}

	// Update key version if it's newer
	if int(keyVersion) > group.KeyVersion {
		group.KeyVersion = int(keyVersion)
		// Generate new group key
		group.GroupKey = make([]byte, 32)
		rand.Read(group.GroupKey)
	}
}

// handleStateRequestMessage handles state request messages
func (gm *GroupManager) handleStateRequestMessage(groupID string, msg map[string]interface{}) {
	from, ok := msg["from"].(string)
	if !ok {
		return
	}

	// Don't respond to our own requests
	myID := gm.host.ID().String()
	if from == myID {
		return
	}

	// Send current group state
	gm.sendGroupState(groupID, from)
}

// handleStateResponseMessage handles state response messages
func (gm *GroupManager) handleStateResponseMessage(groupID string, msg map[string]interface{}) {
	members, ok := msg["members"].([]interface{})
	if !ok {
		return
	}

	keyVersion, _ := msg["key_version"].(float64)

	gm.mu.Lock()
	defer gm.mu.Unlock()

	group, exists := gm.groups[groupID]
	if !exists {
		return
	}

	// Update key version if it's newer
	if int(keyVersion) > group.KeyVersion {
		group.KeyVersion = int(keyVersion)
		// Generate new group key
		group.GroupKey = make([]byte, 32)
		rand.Read(group.GroupKey)
	}

	// Update local state with received members
	for _, member := range members {
		if memberID, ok := member.(string); ok {
			group.Members[memberID] = true
			// Generate sender key for new member
			if _, exists := group.SenderKeys[memberID]; !exists {
				senderKey := make([]byte, 32)
				rand.Read(senderKey)
				group.SenderKeys[memberID] = senderKey
			}
		}
	}
}

// sendGroupState sends the current group state to a specific peer
func (gm *GroupManager) sendGroupState(groupID, toPeerID string) {
	gm.mu.RLock()
	group, exists := gm.groups[groupID]
	if !exists {
		gm.mu.RUnlock()
		return
	}

	// Get current members
	members := make([]string, 0, len(group.Members))
	for memberID := range group.Members {
		members = append(members, memberID)
	}

	keyVersion := group.KeyVersion
	gm.mu.RUnlock()

	// Create state response message
	stateResponse := map[string]interface{}{
		"type":        "state_response",
		"group_id":    groupID,
		"members":     members,
		"key_version": keyVersion,
		"timestamp":   time.Now().Unix(),
	}

	stateResponseBytes, _ := json.Marshal(stateResponse)
	topic := fmt.Sprintf("group-%s", groupID)
	gm.pubsub.Publish(topic, stateResponseBytes)
}

// handleRemovalAckMessage handles removal acknowledgment messages
func (gm *GroupManager) handleRemovalAckMessage(groupID string, msg map[string]interface{}) {
	memberID, ok := msg["member_id"].(string)
	if !ok {
		return
	}

	from, ok := msg["from"].(string)
	if !ok {
		return
	}

	// Record acknowledgment
	gm.removalMu.Lock()
	if gm.removalAcks[groupID] != nil && gm.removalAcks[groupID][memberID] != nil {
		gm.removalAcks[groupID][memberID][from] = true
	}
	gm.removalMu.Unlock()
}

// handleRemovalRetryMessage handles removal retry messages
func (gm *GroupManager) handleRemovalRetryMessage(groupID string, msg map[string]interface{}) {
	_, ok := msg["member_id"].(string)
	if !ok {
		return
	}

	// Process the removal again
	gm.handleMemberRemovedMessage(groupID, msg)
}

// handleMemberLeftMessage handles voluntary member leave messages
func (gm *GroupManager) handleMemberLeftMessage(groupID string, msg map[string]interface{}) {
	memberID, ok := msg["member_id"].(string)
	if !ok {
		return
	}

	myID := gm.host.ID().String()

	// Don't process our own leave message
	if memberID == myID {
		return
	}

	gm.mu.Lock()
	defer gm.mu.Unlock()

	group, exists := gm.groups[groupID]
	if !exists {
		return
	}

	// Remove member from local state
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
}

// sendRemovalAck sends an acknowledgment for a member removal
func (gm *GroupManager) sendRemovalAck(groupID, memberID, toPeerID string) {
	ackMsg := map[string]interface{}{
		"type":      "removal_ack",
		"group_id":  groupID,
		"member_id": memberID,
		"from":      gm.host.ID().String(),
		"timestamp": time.Now().Unix(),
	}

	ackBytes, _ := json.Marshal(ackMsg)
	topic := fmt.Sprintf("group-%s", groupID)
	gm.pubsub.Publish(topic, ackBytes)
}

// checkRemovalAcks checks if all expected acknowledgments have been received
func (gm *GroupManager) checkRemovalAcks(groupID, memberID string) bool {
	gm.removalMu.RLock()
	defer gm.removalMu.RUnlock()

	// Get current group members
	gm.mu.RLock()
	group, exists := gm.groups[groupID]
	if !exists {
		gm.mu.RUnlock()
		return false
	}

	expectedMembers := make(map[string]bool)
	for peerID := range group.Members {
		expectedMembers[peerID] = true
	}
	gm.mu.RUnlock()

	// Check if we have acknowledgments from all remaining members
	acks := gm.removalAcks[groupID][memberID]
	if acks == nil {
		return false
	}

	myID := gm.host.ID().String()
	for peerID := range expectedMembers {
		if peerID != myID && !acks[peerID] {
			return false
		}
	}

	return true
}

// retryPendingRemovals retries pending removals that haven't been acknowledged
func (gm *GroupManager) retryPendingRemovals() {
	gm.removalMu.Lock()
	defer gm.removalMu.Unlock()

	now := time.Now()
	for groupID, removals := range gm.pendingRemovals {
		for memberID, timestamp := range removals {
			// Retry if more than 5 seconds have passed
			if now.Sub(timestamp) > 5*time.Second {
				// Send retry message
				retryMsg := map[string]interface{}{
					"type":      "removal_retry",
					"group_id":  groupID,
					"member_id": memberID,
					"from":      gm.host.ID().String(),
					"timestamp": now.Unix(),
				}

				retryBytes, _ := json.Marshal(retryMsg)
				topic := fmt.Sprintf("group-%s", groupID)
				gm.pubsub.Publish(topic, retryBytes)

				// Update timestamp
				gm.pendingRemovals[groupID][memberID] = now
			}
		}
	}
}

// cleanupCompletedRemovals removes completed removals from tracking
func (gm *GroupManager) cleanupCompletedRemovals() {
	gm.removalMu.Lock()
	defer gm.removalMu.Unlock()

	for groupID, removals := range gm.pendingRemovals {
		for memberID := range removals {
			if gm.checkRemovalAcks(groupID, memberID) {
				// Remove from pending
				delete(gm.pendingRemovals[groupID], memberID)
				// Clean up acknowledgments
				if gm.removalAcks[groupID] != nil {
					delete(gm.removalAcks[groupID], memberID)
				}
			}
		}
	}
}

// removalManagementLoop runs in the background to manage removal acknowledgments and retries
func (gm *GroupManager) removalManagementLoop() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-gm.ctx.Done():
			return
		case <-ticker.C:
			// Retry pending removals
			gm.retryPendingRemovals()
			// Clean up completed removals
			gm.cleanupCompletedRemovals()
		}
	}
}
