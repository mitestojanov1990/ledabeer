package storage

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

type Message struct {
	From             string
	To               string
	EncryptedContent []byte
	Timestamp        time.Time
	TTL              time.Duration
}

type MessageStore struct {
	ipfs  *IPFSNode
	index map[string][]string // peer ID -> CID list
	mutex sync.RWMutex
}

func NewMessageStore(ipfs *IPFSNode) *MessageStore {
	return &MessageStore{
		ipfs:  ipfs,
		index: make(map[string][]string),
	}
}

func (ms *MessageStore) StoreMessage(ctx context.Context, msg *Message) (string, error) {
	// Serialize message
	data, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	// Store in IPFS
	cid, err := ms.ipfs.Add(ctx, data)
	if err != nil {
		return "", err
	}

	// Add CID to peer's index
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	ms.index[msg.To] = append(ms.index[msg.To], cid)

	// Pin for TTL duration if specified
	if msg.TTL > 0 {
		ms.ipfs.Pin(ctx, cid)

		// Schedule unpinning after TTL
		go func() {
			time.Sleep(msg.TTL)
			ms.ipfs.Unpin(ctx, cid)
		}()
	}

	return cid, nil
}

func (ms *MessageStore) GetMessagesFor(ctx context.Context, peerID string) ([]*Message, error) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	// Get CIDs from index
	cids, exists := ms.index[peerID]
	if !exists {
		return []*Message{}, nil
	}

	// Retrieve messages from IPFS
	var messages []*Message
	for _, cid := range cids {
		data, err := ms.ipfs.Get(ctx, cid)
		if err != nil {
			continue // Skip failed retrievals (may be expired)
		}

		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			continue // Skip invalid messages
		}

		// Check if message has expired
		if msg.TTL > 0 && time.Since(msg.Timestamp) > msg.TTL {
			continue // Skip expired messages
		}

		messages = append(messages, &msg)
	}

	// Clear index after retrieval
	delete(ms.index, peerID)

	return messages, nil
}

func (ms *MessageStore) Close() error {
	return ms.ipfs.Close()
}
