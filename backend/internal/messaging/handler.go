package messaging

import (
	"context"
	"crypto/rand"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
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
	return &MessageHandler{
		host:        h,
		sent:        make(map[string]bool),
		subscribers: make([]chan Message, 0),
	}
}

func (m *MessageHandler) SendMessage(ctx context.Context, toPeerID string, content []byte) (string, error) {
	messageID := generateMessageID()

	// Parse peer ID
	peerID, err := peer.Decode(toPeerID)
	if err != nil {
		return "", fmt.Errorf("invalid peer ID: %w", err)
	}

	// Create message (for storage if needed)
	_ = Message{
		ID:        messageID,
		From:      m.host.ID().String(),
		Content:   content,
		Timestamp: time.Now().Unix(),
	}

	// Send via libp2p stream
	stream, err := m.host.NewStream(ctx, peerID, protocol.ID("/chat/1.0.0"))
	if err != nil {
		return "", fmt.Errorf("failed to create stream: %w", err)
	}
	defer stream.Close()

	// Write message (simplified - in real implementation, this would be encrypted)
	_, err = stream.Write(content)
	if err != nil {
		return "", fmt.Errorf("failed to write message: %w", err)
	}

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
	m.mutex.Unlock()

	// Start message handler if not already started
	go m.handleIncomingMessages(ctx)

	return msgChan
}

func (m *MessageHandler) handleIncomingMessages(ctx context.Context) {
	// Set up stream handler for incoming messages
	// For now, just a placeholder - in real implementation this would handle streams
	// m.host.SetStreamHandler(protocol.ID("/chat/1.0.0"), handler)
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
	// For now, return empty history
	// In real implementation, this would query IPFS storage
	return []Message{}, nil
}

func generateMessageID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)
}
