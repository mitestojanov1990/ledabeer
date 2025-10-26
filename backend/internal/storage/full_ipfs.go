package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	shell "github.com/ipfs/go-ipfs-api"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

type FullIPFSNode struct {
	shell    *shell.Shell
	host     host.Host
	config   *IPFSConfig
	peers    []peer.ID
	mutex    sync.RWMutex
	closed   bool
	process  *exec.Cmd
	repoPath string
}

type FullIPFSConfig struct {
	RepoPath      string
	Bootstrap     []string
	APIPort       int
	GatewayPort   int
	SwarmPort     int
	EnableAPI     bool
	EnableGateway bool
}

func NewFullIPFSNode(ctx context.Context, cfg *FullIPFSConfig) (*FullIPFSNode, error) {
	// Initialize IPFS repo if needed
	repoPath := cfg.RepoPath
	if repoPath == "" {
		repoPath = filepath.Join(os.TempDir(), "ipfs-repo")
	}

	// Create repo directory
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create repo directory: %w", err)
	}

	// Initialize IPFS repo if it doesn't exist
	initCmd := exec.Command("ipfs", "init", "--repo", repoPath)
	if err := initCmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to initialize IPFS repo: %w", err)
	}

	// Start IPFS daemon
	daemonCmd := exec.Command("ipfs", "daemon", "--repo", repoPath)
	if err := daemonCmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start IPFS daemon: %w", err)
	}

	// Wait for daemon to start
	time.Sleep(2 * time.Second)

	// Connect to IPFS API
	shell := shell.NewShell("localhost:5001")

	node := &FullIPFSNode{
		shell:    shell,
		config:   &IPFSConfig{RepoPath: repoPath, Bootstrap: cfg.Bootstrap},
		peers:    make([]peer.ID, 0),
		process:  daemonCmd,
		repoPath: repoPath,
	}

	// Connect to bootstrap peers
	go func() {
		time.Sleep(1 * time.Second)
		node.mutex.Lock()
		defer node.mutex.Unlock()

		// Connect to bootstrap peers
		for range cfg.Bootstrap {
			// In a real implementation, this would connect to actual bootstrap nodes
			peerID := generateMockPeerID()
			node.peers = append(node.peers, peerID)
		}
	}()

	return node, nil
}

func NewFullIPFSNodeWithHost(ctx context.Context, h host.Host, cfg *FullIPFSConfig) (*FullIPFSNode, error) {
	// For now, use the same implementation as NewFullIPFSNode
	// In a real implementation, this would share the libp2p host
	return NewFullIPFSNode(ctx, cfg)
}

func (n *FullIPFSNode) ID() string {
	if n.shell == nil {
		return "unknown-ipfs-node-id"
	}

	id, err := n.shell.ID()
	if err != nil {
		return "unknown-ipfs-node-id"
	}

	return id.ID
}

func (n *FullIPFSNode) Peers() []peer.ID {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	return n.peers
}

func (n *FullIPFSNode) Close() error {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	n.closed = true

	// Stop IPFS daemon
	if n.process != nil {
		n.process.Process.Kill()
		n.process.Wait()
	}

	return nil
}

// Content storage methods using real IPFS
func (n *FullIPFSNode) Add(ctx context.Context, data []byte) (string, error) {
	if n.shell == nil {
		return "", fmt.Errorf("IPFS shell not initialized")
	}

	// Add data to IPFS
	cid, err := n.shell.Add(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to add to IPFS: %w", err)
	}

	return cid, nil
}

func (n *FullIPFSNode) Get(ctx context.Context, cid string) ([]byte, error) {
	if n.shell == nil {
		return nil, fmt.Errorf("IPFS shell not initialized")
	}

	// Get data from IPFS
	reader, err := n.shell.Cat(cid)
	if err != nil {
		return nil, fmt.Errorf("failed to get from IPFS: %w", err)
	}
	defer reader.Close()

	// Read all data
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	return data, nil
}

func (n *FullIPFSNode) Pin(ctx context.Context, cid string) error {
	if n.shell == nil {
		return fmt.Errorf("IPFS shell not initialized")
	}

	// Pin in IPFS
	err := n.shell.Pin(cid)
	if err != nil {
		return fmt.Errorf("failed to pin: %w", err)
	}

	return nil
}

func (n *FullIPFSNode) Unpin(ctx context.Context, cid string) error {
	if n.shell == nil {
		return fmt.Errorf("IPFS shell not initialized")
	}

	// Unpin in IPFS
	err := n.shell.Unpin(cid)
	if err != nil {
		return fmt.Errorf("failed to unpin: %w", err)
	}

	return nil
}

func (n *FullIPFSNode) IsPinned(ctx context.Context, cid string) bool {
	if n.shell == nil {
		return false
	}

	// Check if pinned
	pins, err := n.shell.Pins()
	if err != nil {
		return false
	}

	_, exists := pins[cid]
	return exists
}

// Helper function to check if IPFS is installed
func IsIPFSInstalled() bool {
	_, err := exec.LookPath("ipfs")
	return err == nil
}
