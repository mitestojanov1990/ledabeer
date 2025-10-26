package storage

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

type MockFullIPFSNode struct {
	host     host.Host
	config   *IPFSConfig
	content  map[string][]byte // CID -> content
	pinned   map[string]bool   // CID -> pinned status
	peers    []peer.ID
	mutex    sync.RWMutex
	closed   bool
	realMode bool
}

type MockFullIPFSConfig struct {
	RepoPath      string
	Bootstrap     []string
	APIPort       int
	GatewayPort   int
	SwarmPort     int
	EnableAPI     bool
	EnableGateway bool
}

func NewMockFullIPFSNode(ctx context.Context, cfg *MockFullIPFSConfig) (*MockFullIPFSNode, error) {
	// Create mock full IPFS node with enhanced features
	node := &MockFullIPFSNode{
		config:   &IPFSConfig{RepoPath: cfg.RepoPath, Bootstrap: cfg.Bootstrap},
		content:  make(map[string][]byte),
		pinned:   make(map[string]bool),
		peers:    make([]peer.ID, 0),
		realMode: true,
	}

	// Simulate connecting to bootstrap peers
	go func() {
		time.Sleep(100 * time.Millisecond)
		node.mutex.Lock()
		defer node.mutex.Unlock()

		// Add some dummy peers for testing
		if len(cfg.Bootstrap) > 0 {
			// Simulate peer connections
			for i := 0; i < 2; i++ {
				peerID := generateMockPeerID()
				node.peers = append(node.peers, peerID)
			}
		}
	}()

	return node, nil
}

func NewMockFullIPFSNodeWithHost(ctx context.Context, h host.Host, cfg *MockFullIPFSConfig) (*MockFullIPFSNode, error) {
	// Create mock full IPFS node using existing libp2p host
	node := &MockFullIPFSNode{
		host:     h,
		config:   &IPFSConfig{RepoPath: cfg.RepoPath, Bootstrap: cfg.Bootstrap},
		content:  make(map[string][]byte),
		pinned:   make(map[string]bool),
		peers:    make([]peer.ID, 0),
		realMode: true,
	}

	// Simulate connecting to bootstrap peers
	go func() {
		time.Sleep(100 * time.Millisecond)
		node.mutex.Lock()
		defer node.mutex.Unlock()

		// Add some dummy peers for testing
		if len(cfg.Bootstrap) > 0 {
			// Simulate peer connections
			for i := 0; i < 2; i++ {
				peerID := generateMockPeerID()
				node.peers = append(node.peers, peerID)
			}
		}
	}()

	return node, nil
}

func (n *MockFullIPFSNode) ID() string {
	if n.host != nil {
		return n.host.ID().String()
	}
	// Generate a real-looking IPFS node ID
	return generateRealIPFSNodeID()
}

func (n *MockFullIPFSNode) Peers() []peer.ID {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	return n.peers
}

func (n *MockFullIPFSNode) Close() error {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	n.closed = true
	return nil
}

// Content storage methods with enhanced IPFS features
func (n *MockFullIPFSNode) Add(ctx context.Context, data []byte) (string, error) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	// Generate real IPFS CID
	cid := generateRealCID(data)
	n.content[cid] = data

	// Simulate IPFS network propagation
	go func() {
		time.Sleep(50 * time.Millisecond)
		// In a real implementation, this would propagate to the IPFS network
	}()

	return cid, nil
}

func (n *MockFullIPFSNode) Get(ctx context.Context, cid string) ([]byte, error) {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	content, exists := n.content[cid]
	if !exists {
		return nil, fmt.Errorf("content not found: %s", cid)
	}
	return content, nil
}

func (n *MockFullIPFSNode) Pin(ctx context.Context, cid string) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	// Verify content exists
	if _, exists := n.content[cid]; !exists {
		return fmt.Errorf("content not found: %s", cid)
	}

	n.pinned[cid] = true

	// Simulate IPFS pinning
	go func() {
		time.Sleep(10 * time.Millisecond)
		// In a real implementation, this would pin in the IPFS network
	}()

	return nil
}

func (n *MockFullIPFSNode) Unpin(ctx context.Context, cid string) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	delete(n.pinned, cid)

	// Simulate IPFS unpinning
	go func() {
		time.Sleep(10 * time.Millisecond)
		// In a real implementation, this would unpin from the IPFS network
	}()

	return nil
}

func (n *MockFullIPFSNode) IsPinned(ctx context.Context, cid string) bool {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	return n.pinned[cid]
}

// Enhanced IPFS features
func (n *MockFullIPFSNode) GetStats(ctx context.Context) (map[string]interface{}, error) {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	return map[string]interface{}{
		"repo_size":    len(n.content) * 1024, // Simulate repo size
		"pinned_count": len(n.pinned),
		"peer_count":   len(n.peers),
		"mode":         "full-ipfs-mock",
	}, nil
}

func (n *MockFullIPFSNode) ConnectToPeer(ctx context.Context, peerID string) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	// Simulate connecting to a peer
	// In a real implementation, this would connect to the actual peer
	return nil
}

func (n *MockFullIPFSNode) PublishToPubSub(ctx context.Context, topic string, data []byte) error {
	// Simulate publishing to IPFS pubsub
	// In a real implementation, this would publish to the IPFS pubsub network
	return nil
}

func (n *MockFullIPFSNode) SubscribeToPubSub(ctx context.Context, topic string) (<-chan []byte, error) {
	// Simulate subscribing to IPFS pubsub
	// In a real implementation, this would subscribe to the IPFS pubsub network
	ch := make(chan []byte, 1)
	go func() {
		time.Sleep(100 * time.Millisecond)
		close(ch)
	}()
	return ch, nil
}
