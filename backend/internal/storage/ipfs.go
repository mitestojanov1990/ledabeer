package storage

import (
	"context"
	"crypto/rand"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

type IPFSConfig struct {
	RepoPath  string
	Bootstrap []string
}

type IPFSNode struct {
	host    host.Host
	config  *IPFSConfig
	content map[string][]byte // CID -> content
	pinned  map[string]bool   // CID -> pinned status
	peers   []peer.ID
	mutex   sync.RWMutex
	closed  bool
}

func NewIPFSNode(ctx context.Context, cfg *IPFSConfig) (*IPFSNode, error) {
	// For testing, create a mock IPFS node
	// In a real implementation, this would initialize a full IPFS node
	node := &IPFSNode{
		config:  cfg,
		content: make(map[string][]byte),
		pinned:  make(map[string]bool),
		peers:   make([]peer.ID, 0),
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

func NewIPFSNodeWithHost(ctx context.Context, h host.Host, cfg *IPFSConfig) (*IPFSNode, error) {
	// Create IPFS node using existing libp2p host
	node := &IPFSNode{
		host:    h,
		config:  cfg,
		content: make(map[string][]byte),
		pinned:  make(map[string]bool),
		peers:   make([]peer.ID, 0),
	}

	// Simulate connecting to bootstrap peers
	go func() {
		time.Sleep(1 * time.Second)
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

func (n *IPFSNode) ID() string {
	if n.host != nil {
		return n.host.ID().String()
	}
	// Generate a mock ID for testing
	return "mock-ipfs-node-id"
}

func (n *IPFSNode) Peers() []peer.ID {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	return n.peers
}

func (n *IPFSNode) Close() error {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	n.closed = true
	return nil
}

// Content storage methods
func (n *IPFSNode) Add(ctx context.Context, data []byte) (string, error) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	// Generate a mock CID for testing
	cid := generateMockCID()
	n.content[cid] = data
	return cid, nil
}

func (n *IPFSNode) Get(ctx context.Context, cid string) ([]byte, error) {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	content, exists := n.content[cid]
	if !exists {
		return nil, fmt.Errorf("content not found: %s", cid)
	}
	return content, nil
}

func (n *IPFSNode) Pin(ctx context.Context, cid string) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	// Verify content exists
	if _, exists := n.content[cid]; !exists {
		return fmt.Errorf("content not found: %s", cid)
	}

	n.pinned[cid] = true
	return nil
}

func (n *IPFSNode) Unpin(ctx context.Context, cid string) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	delete(n.pinned, cid)
	return nil
}

func (n *IPFSNode) IsPinned(ctx context.Context, cid string) bool {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	return n.pinned[cid]
}

// Helper function to generate mock peer IDs for testing
func generateMockPeerID() peer.ID {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return peer.ID(bytes)
}

// Helper function to generate mock CIDs for testing
func generateMockCID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("Qm%s", string(bytes))
}
