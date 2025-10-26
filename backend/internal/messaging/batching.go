package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

type BatchingConfig struct {
	BatchSize    int
	BatchTimeout time.Duration
	MaxBatchSize int
}

type BatchingStats struct {
	BatchesSent     int
	MessagesBatched int
	MessagesPending int
	BatchesDropped  int
}

type BatchedMessage struct {
	ID        string `json:"id"`
	To        string `json:"to"`
	Content   []byte `json:"content"`
	Timestamp int64  `json:"timestamp"`
}

type MessageBatch struct {
	Messages []BatchedMessage `json:"messages"`
	To       string           `json:"to"`
	SentAt   int64            `json:"sent_at"`
}

type MessageHandlerWithBatching struct {
	host     host.Host
	config   *BatchingConfig
	batches  map[peer.ID]*MessageBatch
	stats    *BatchingStats
	mutex    sync.RWMutex
	stopChan chan struct{}
}

func NewMessageHandlerWithBatching(h host.Host, config *BatchingConfig) *MessageHandlerWithBatching {
	handler := &MessageHandlerWithBatching{
		host:     h,
		config:   config,
		batches:  make(map[peer.ID]*MessageBatch),
		stats:    &BatchingStats{},
		stopChan: make(chan struct{}),
	}

	// Start batch processing goroutine
	go handler.batchProcessor()

	return handler
}

func (m *MessageHandlerWithBatching) SendMessage(ctx context.Context, toPeerID string, content []byte) (string, error) {
	peerID, err := peer.Decode(toPeerID)
	if err != nil {
		return "", fmt.Errorf("invalid peer ID: %w", err)
	}

	messageID := generateBatchingMessageID()
	message := BatchedMessage{
		ID:        messageID,
		To:        toPeerID,
		Content:   content,
		Timestamp: time.Now().Unix(),
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Get or create batch for this peer
	batch, exists := m.batches[peerID]
	if !exists {
		batch = &MessageBatch{
			Messages: make([]BatchedMessage, 0),
			To:       toPeerID,
		}
		m.batches[peerID] = batch
	}

	// Add message to batch
	batch.Messages = append(batch.Messages, message)
	m.stats.MessagesPending++

	// Check if batch should be sent
	if len(batch.Messages) >= m.config.BatchSize || len(batch.Messages) >= m.config.MaxBatchSize {
		go m.sendBatch(ctx, peerID, batch)
		delete(m.batches, peerID)
		m.stats.MessagesPending -= len(batch.Messages)
	}

	return messageID, nil
}

func (m *MessageHandlerWithBatching) sendBatch(ctx context.Context, peerID peer.ID, batch *MessageBatch) error {
	// Create stream to peer
	stream, err := m.host.NewStream(ctx, peerID, protocol.ID("/ledabeer/messaging"))
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}
	defer stream.Close()

	// Serialize batch
	batchData, err := json.Marshal(batch)
	if err != nil {
		return fmt.Errorf("failed to marshal batch: %w", err)
	}

	// Send batch
	_, err = stream.Write(batchData)
	if err != nil {
		return fmt.Errorf("failed to send batch: %w", err)
	}

	// Update stats
	m.mutex.Lock()
	m.stats.BatchesSent++
	m.stats.MessagesBatched += len(batch.Messages)
	m.mutex.Unlock()

	return nil
}

func (m *MessageHandlerWithBatching) batchProcessor() {
	ticker := time.NewTicker(m.config.BatchTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.flushPendingBatches()
		case <-m.stopChan:
			return
		}
	}
}

func (m *MessageHandlerWithBatching) flushPendingBatches() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	ctx := context.Background()
	for peerID, batch := range m.batches {
		if len(batch.Messages) > 0 {
			go m.sendBatch(ctx, peerID, batch)
			delete(m.batches, peerID)
			m.stats.MessagesPending -= len(batch.Messages)
		}
	}
}

func (m *MessageHandlerWithBatching) GetBatchingStats() BatchingStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return BatchingStats{
		BatchesSent:     m.stats.BatchesSent,
		MessagesBatched: m.stats.MessagesBatched,
		MessagesPending: m.stats.MessagesPending,
		BatchesDropped:  m.stats.BatchesDropped,
	}
}

func (m *MessageHandlerWithBatching) Close() {
	close(m.stopChan)

	// Flush any remaining batches
	m.flushPendingBatches()
}

// Helper function to generate message ID for batching
func generateBatchingMessageID() string {
	// Simple implementation for testing
	return fmt.Sprintf("batch_msg_%d", time.Now().UnixNano())
}
