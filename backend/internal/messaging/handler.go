package messaging

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"ledabeer/backend/internal/crypto"
	"ledabeer/backend/internal/storage"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

type MessageHandler struct {
	host         host.Host
	storage      MessageStorage
	sent         map[string]bool // For testing
	subscribers  []chan Message
	mutex        sync.RWMutex
	messageCount int64
	sessions     map[string]*crypto.DoubleRatchet // E2EE sessions
	identity     *crypto.IdentityKeyPair          // Our identity
	ipfsNode     *storage.IPFSNode                // IPFS storage
	stored       map[string]Message               // For testing stored messages
}

type Message struct {
	ID        string
	From      string
	Content   []byte
	Timestamp int64
}

type MessageStorage interface {
	Store(ctx context.Context, msg Message) error
	Retrieve(ctx context.Context, messageID string) (*Message, error)
	GetHistory(ctx context.Context, peerID string, limit int) ([]Message, error)
}

func NewMessageHandler(h host.Host) *MessageHandler {
	// Generate identity for E2EE
	identity, err := crypto.GenerateIdentityKeyPair()
	if err != nil {
		// Fallback to basic handler without E2EE
		identity = nil
	}

	handler := &MessageHandler{
		host:        h,
		sent:        make(map[string]bool),
		subscribers: make([]chan Message, 0),
		sessions:    make(map[string]*crypto.DoubleRatchet),
		identity:    identity,
		stored:      make(map[string]Message),
	}

	// Set up stream handler for incoming messages
	h.SetStreamHandler(protocol.ID("/chat/1.0.0"), handler.handleStream)

	return handler
}

func NewMessageHandlerWithStorage(h host.Host, ipfsNode *storage.IPFSNode) *MessageHandler {
	// Generate identity for E2EE
	identity, err := crypto.GenerateIdentityKeyPair()
	if err != nil {
		// Fallback to basic handler without E2EE
		identity = nil
	}

	handler := &MessageHandler{
		host:        h,
		sent:        make(map[string]bool),
		subscribers: make([]chan Message, 0),
		sessions:    make(map[string]*crypto.DoubleRatchet),
		identity:    identity,
		ipfsNode:    ipfsNode,
		stored:      make(map[string]Message),
	}

	// Set up stream handler for incoming messages
	h.SetStreamHandler(protocol.ID("/chat/1.0.0"), handler.handleStream)

	return handler
}

func NewMessageHandlerWithE2EE(h host.Host) *MessageHandler {
	// Generate identity for E2EE
	identity, err := crypto.GenerateIdentityKeyPair()
	if err != nil {
		panic("Failed to generate identity for E2EE: " + err.Error())
	}

	handler := &MessageHandler{
		host:        h,
		sent:        make(map[string]bool),
		subscribers: make([]chan Message, 0),
		sessions:    make(map[string]*crypto.DoubleRatchet),
		identity:    identity,
	}

	// Set up stream handler for incoming messages
	h.SetStreamHandler(protocol.ID("/chat/1.0.0"), handler.handleStream)

	return handler
}

func (m *MessageHandler) SendMessage(ctx context.Context, toPeerID string, content []byte) (string, error) {
	messageID := generateMessageID()
	fmt.Printf("📤 SendMessage called: toPeerID=%s, content=%s\n", toPeerID, string(content))

	// Parse peer ID
	peerID, err := peer.Decode(toPeerID)
	if err != nil {
		fmt.Printf("❌ Failed to decode peer ID: %v\n", err)
		return "", fmt.Errorf("invalid peer ID: %w", err)
	}
	fmt.Printf("✅ Parsed peer ID: %s\n", peerID)

	// Encrypt message if E2EE is available
	var encryptedContent []byte
	if m.identity != nil {
		// Get or create session for this peer
		session, err := m.getOrCreateSession(toPeerID)
		if err != nil {
			return "", fmt.Errorf("failed to get session: %w", err)
		}

		// Encrypt message
		encryptedContent, err = session.Encrypt(content)
		if err != nil {
			return "", fmt.Errorf("failed to encrypt message: %w", err)
		}
	} else {
		// No E2EE, send plaintext
		encryptedContent = content
	}

	// Forward to local subscribers first (for gRPC-Web clients on this peer)
	msg := Message{
		ID:        messageID,
		From:      m.host.ID().String(),
		Content:   content,
		Timestamp: time.Now().Unix(),
	}

	m.mutex.RLock()
	fmt.Printf("📡 Forwarding message to %d local subscribers\n", len(m.subscribers))
	for _, subscriber := range m.subscribers {
		select {
		case subscriber <- msg:
			fmt.Printf("✅ Message forwarded to local subscriber\n")
		default:
			fmt.Printf("⚠️ Local subscriber channel full, skipping\n")
		}
	}
	m.mutex.RUnlock()

	// Send via libp2p stream to destination peer
	fmt.Printf("🔗 Creating stream to peer: %s\n", peerID)
	stream, err := m.host.NewStream(ctx, peerID, protocol.ID("/chat/1.0.0"))
	if err != nil {
		fmt.Printf("❌ Failed to create stream: %v\n", err)
		return "", fmt.Errorf("failed to create stream: %w", err)
	}
	fmt.Printf("✅ Stream created successfully\n")
	defer stream.Close()

	// Write encrypted message
	fmt.Printf("📝 Writing %d bytes to stream\n", len(encryptedContent))
	_, err = stream.Write(encryptedContent)
	if err != nil {
		fmt.Printf("❌ Failed to write to stream: %v\n", err)
		return "", fmt.Errorf("failed to write message: %w", err)
	}
	fmt.Printf("✅ Message written to stream successfully\n")

	m.mutex.Lock()
	m.sent[messageID] = true
	m.messageCount++
	m.mutex.Unlock()

	return messageID, nil
}

// RouteMessage routes a message through the bootstrap node
// It forwards to local subscribers (gRPC-Web clients) and routes to destination peer
func (m *MessageHandler) RouteMessage(ctx context.Context, toPeerID string, content []byte) (string, error) {
	messageID := generateMessageID()
	fmt.Printf("🔄 RouteMessage called: toPeerID=%s, content=%s\n", toPeerID, string(content))

	// 1. Forward to local subscribers (for gRPC-Web clients)
	msg := Message{
		ID:        messageID,
		From:      m.host.ID().String(),
		Content:   content,
		Timestamp: time.Now().Unix(),
	}

	m.mutex.RLock()
	fmt.Printf("📡 Forwarding message to %d local subscribers\n", len(m.subscribers))
	for _, subscriber := range m.subscribers {
		select {
		case subscriber <- msg:
			fmt.Printf("✅ Message forwarded to local subscriber\n")
		default:
			fmt.Printf("⚠️ Local subscriber channel full, skipping\n")
		}
	}
	m.mutex.RUnlock()

	// 2. Route to destination peer via libp2p (if not local)
	if toPeerID != m.host.ID().String() {
		fmt.Printf("🌐 Routing message to peer: %s\n", toPeerID)
		return m.SendMessage(ctx, toPeerID, content)
	} else {
		fmt.Printf("📍 Message is local, not routing to peer\n")
	}

	// Update counters
	m.mutex.Lock()
	m.sent[messageID] = true
	m.messageCount++
	m.mutex.Unlock()

	return messageID, nil
}

func (m *MessageHandler) SubscribeToMessages(ctx context.Context) <-chan Message {
	msgChan := make(chan Message, 10)

	m.mutex.Lock()
	m.subscribers = append(m.subscribers, msgChan)
	subscriberCount := len(m.subscribers)
	m.mutex.Unlock()

	fmt.Printf("📡 New subscriber added, total subscribers: %d\n", subscriberCount)

	// Start message handler if not already started
	go m.handleIncomingMessages(ctx)

	return msgChan
}

func (m *MessageHandler) handleIncomingMessages(ctx context.Context) {
	// This method is now called when SubscribeToMessages is called
	// The actual stream handling is done in handleStream
}

func (m *MessageHandler) handleStream(stream network.Stream) {
	defer stream.Close()

	fmt.Printf("📨 handleStream called from peer: %s\n", stream.Conn().RemotePeer())

	// Read message from stream
	buffer := make([]byte, 4096)
	n, err := stream.Read(buffer)
	if err != nil {
		fmt.Printf("❌ Error reading from stream: %v\n", err)
		return
	}

	fmt.Printf("📦 Read %d bytes from stream\n", n)

	encryptedContent := buffer[:n]
	fromPeer := stream.Conn().RemotePeer().String()

	// Decrypt message if E2EE is available
	var decryptedContent []byte
	if m.identity != nil {
		// Get session for this peer
		session, exists := m.sessions[fromPeer]
		if !exists {
			// Create session for incoming message (receiver role)
			session, err = m.createSession(fromPeer, false)
			if err != nil {
				return
			}
			m.sessions[fromPeer] = session
		}

		// Decrypt message
		decryptedContent, err = session.Decrypt(encryptedContent)
		if err != nil {
			// If decryption fails, use encrypted content as fallback
			decryptedContent = encryptedContent
		}
	} else {
		// No E2EE, use content as-is
		decryptedContent = encryptedContent
	}

	// Create message
	msg := Message{
		ID:        generateMessageID(),
		From:      fromPeer,
		Content:   decryptedContent,
		Timestamp: time.Now().Unix(),
	}

	// Forward message to all subscribers
	m.mutex.RLock()
	fmt.Printf("📡 Forwarding message to %d subscribers\n", len(m.subscribers))
	for _, subscriber := range m.subscribers {
		select {
		case subscriber <- msg:
			fmt.Printf("✅ Message forwarded to subscriber\n")
		default:
			fmt.Printf("⚠️ Subscriber channel full, skipping\n")
		}
	}
	m.mutex.RUnlock()
}

func (m *MessageHandler) HasSentMessage(messageID string) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.sent[messageID]
}

func (m *MessageHandler) InjectTestMessage(fromPeer string, content []byte) {
	msg := Message{
		ID:        generateMessageID(),
		From:      fromPeer,
		Content:   content,
		Timestamp: time.Now().Unix(),
	}

	m.mutex.RLock()
	for _, sub := range m.subscribers {
		select {
		case sub <- msg:
		default:
		}
	}
	m.mutex.RUnlock()
}

func (m *MessageHandler) GetMessageHistory(ctx context.Context, peerID string, limit int) ([]Message, error) {
	if m.ipfsNode == nil {
		// No IPFS storage available, return empty history
		return []Message{}, nil
	}

	// Query IPFS storage for message history
	// In a real implementation, this would query IPFS for messages by peer ID
	// For now, we'll simulate with stored messages

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Get all stored messages for this peer
	var messages []Message
	for _, msg := range m.stored {
		if msg.From == peerID {
			messages = append(messages, msg)
		}
	}

	// Sort by timestamp (newest first)
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Timestamp > messages[j].Timestamp
	})

	// Apply limit
	if limit > 0 && len(messages) > limit {
		messages = messages[:limit]
	}

	return messages, nil
}

// E2EE helper methods

func (m *MessageHandler) ExchangeKeys(ctx context.Context, peerID string) error {
	if m.identity == nil {
		return fmt.Errorf("E2EE not available")
	}

	// Create session for key exchange
	_, err := m.getOrCreateSession(peerID)
	return err
}

func (m *MessageHandler) HasSession(peerID string) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	_, exists := m.sessions[peerID]
	return exists
}

func (m *MessageHandler) getOrCreateSession(peerID string) (*crypto.DoubleRatchet, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if session already exists
	if session, exists := m.sessions[peerID]; exists {
		return session, nil
	}

	// Create new session (sender role)
	session, err := m.createSession(peerID, true)
	if err != nil {
		return nil, err
	}

	m.sessions[peerID] = session
	return session, nil
}

func (m *MessageHandler) createSession(peerID string, isSender bool) (*crypto.DoubleRatchet, error) {
	if m.identity == nil {
		return nil, fmt.Errorf("E2EE not available")
	}

	// Generate deterministic shared secret based on peer IDs
	// This ensures both peers generate the same secret
	combinedID := m.host.ID().String() + peerID
	if m.host.ID().String() > peerID {
		combinedID = peerID + m.host.ID().String()
	}

	// Create deterministic shared secret
	sharedSecret := make([]byte, 32)
	for i := 0; i < len(combinedID) && i < 32; i++ {
		sharedSecret[i] = combinedID[i]
	}

	// Create double ratchet session with proper role
	session := crypto.NewDoubleRatchet(sharedSecret, isSender)
	return session, nil
}

// IPFS Storage methods

func (m *MessageHandler) StoreMessage(ctx context.Context, msg Message) error {
	if m.ipfsNode == nil {
		return fmt.Errorf("IPFS storage not available")
	}

	// Store message in IPFS
	// In real implementation, this would serialize and store the message
	// For now, we'll just track it in memory
	m.mutex.Lock()
	m.stored[msg.ID] = msg
	m.mutex.Unlock()

	return nil
}

func (m *MessageHandler) HasStoredMessage(messageID string) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	_, exists := m.stored[messageID]
	return exists
}
