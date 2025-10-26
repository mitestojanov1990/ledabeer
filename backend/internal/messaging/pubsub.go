package messaging

import (
	"context"
	"fmt"
	"regexp"
	"sync"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
)

// PubSub wraps libp2p pubsub functionality
type PubSub struct {
	ps         *pubsub.PubSub
	host       host.Host
	ctx        context.Context
	topics     map[string]*pubsub.Topic
	subs       map[string]*pubsub.Subscription
	mu         sync.RWMutex
	msgHandler func(string, []byte)
	// Message deduplication
	seenMessages map[string]time.Time
	dedupMu      sync.RWMutex
	// Topic validation
	topicRegex *regexp.Regexp
}

// NewPubSub creates a new PubSub instance
func NewPubSub(ctx context.Context, h host.Host) (*PubSub, error) {
	// Create GossipSub instance
	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub: %w", err)
	}

	// Compile topic validation regex (alphanumeric, hyphens, underscores)
	topicRegex, err := regexp.Compile(`^[a-zA-Z0-9_-]+$`)
	if err != nil {
		return nil, fmt.Errorf("failed to compile topic regex: %w", err)
	}

	psub := &PubSub{
		ps:           ps,
		host:         h,
		ctx:          ctx,
		topics:       make(map[string]*pubsub.Topic),
		subs:         make(map[string]*pubsub.Subscription),
		seenMessages: make(map[string]time.Time),
		topicRegex:   topicRegex,
	}

	// Start cleanup goroutine for message deduplication
	go psub.cleanupSeenMessages()

	return psub, nil
}

// Subscribe subscribes to a topic
func (p *PubSub) Subscribe(topicName string) (*pubsub.Subscription, error) {
	// Validate topic name
	if err := p.validateTopic(topicName); err != nil {
		return nil, err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if already subscribed
	if _, exists := p.subs[topicName]; exists {
		return p.subs[topicName], nil
	}

	// Get or create topic
	topic, err := p.ps.Join(topicName)
	if err != nil {
		return nil, fmt.Errorf("failed to join topic %s: %w", topicName, err)
	}

	// Subscribe to topic
	sub, err := topic.Subscribe()
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to topic %s: %w", topicName, err)
	}

	// Store references
	p.topics[topicName] = topic
	p.subs[topicName] = sub

	// Start message handler goroutine
	go p.handleMessages(topicName, sub)

	return sub, nil
}

// Publish publishes a message to a topic
func (p *PubSub) Publish(topicName string, data []byte) error {
	p.mu.RLock()
	topic, exists := p.topics[topicName]
	p.mu.RUnlock()

	if !exists {
		// Join topic if not already joined
		var err error
		topic, err = p.ps.Join(topicName)
		if err != nil {
			return fmt.Errorf("failed to join topic %s: %w", topicName, err)
		}

		p.mu.Lock()
		p.topics[topicName] = topic
		p.mu.Unlock()
	}

	// Publish message
	return topic.Publish(p.ctx, data)
}

// Unsubscribe unsubscribes from a topic
func (p *PubSub) Unsubscribe(topicName string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Close subscription
	if sub, exists := p.subs[topicName]; exists {
		sub.Cancel()
		delete(p.subs, topicName)
	}

	// Leave topic
	if topic, exists := p.topics[topicName]; exists {
		topic.Close()
		delete(p.topics, topicName)
	}

	return nil
}

// IsSubscribed checks if currently subscribed to a topic
func (p *PubSub) IsSubscribed(topicName string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	_, exists := p.subs[topicName]
	return exists
}

// SetMessageHandler sets the message handler callback
func (p *PubSub) SetMessageHandler(handler func(string, []byte)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.msgHandler = handler
}

// handleMessages processes incoming messages for a subscription
func (p *PubSub) handleMessages(topicName string, sub *pubsub.Subscription) {
	for {
		msg, err := sub.Next(p.ctx)
		if err != nil {
			// Context cancelled or subscription closed
			return
		}

		// Check for message deduplication
		msgID := fmt.Sprintf("%s:%s", topicName, string(msg.Data))
		if p.isDuplicateMessage(msgID) {
			continue // Skip duplicate message
		}

		// Call message handler if set
		p.mu.RLock()
		handler := p.msgHandler
		p.mu.RUnlock()

		if handler != nil {
			handler(topicName, msg.Data)
		}
	}
}

// validateTopic validates topic name format
func (p *PubSub) validateTopic(topicName string) error {
	if topicName == "" {
		return fmt.Errorf("topic name cannot be empty")
	}

	if len(topicName) > 100 {
		return fmt.Errorf("topic name too long (max 100 characters)")
	}

	if !p.topicRegex.MatchString(topicName) {
		return fmt.Errorf("invalid topic name format (only alphanumeric, hyphens, underscores allowed)")
	}

	return nil
}

// isDuplicateMessage checks if message was already seen
func (p *PubSub) isDuplicateMessage(msgID string) bool {
	p.dedupMu.Lock()
	defer p.dedupMu.Unlock()

	_, seen := p.seenMessages[msgID]
	if !seen {
		p.seenMessages[msgID] = time.Now()
	}

	return seen
}

// cleanupSeenMessages periodically cleans up old message IDs
func (p *PubSub) cleanupSeenMessages() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.dedupMu.Lock()
			cutoff := time.Now().Add(-10 * time.Minute)
			for msgID, timestamp := range p.seenMessages {
				if timestamp.Before(cutoff) {
					delete(p.seenMessages, msgID)
				}
			}
			p.dedupMu.Unlock()
		}
	}
}
