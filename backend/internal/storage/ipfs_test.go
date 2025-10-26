package storage_test

import (
	"context"
	"testing"
	"time"

	"ledabeer/backend/internal/network"
	"ledabeer/backend/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIPFS_NodeInitialization(t *testing.T) {
	ctx := context.Background()

	// Should initialize IPFS node
	node, err := storage.NewIPFSNode(ctx, &storage.IPFSConfig{
		RepoPath: t.TempDir(),
	})
	require.NoError(t, err)
	defer node.Close()

	// Should have valid peer ID
	assert.NotEmpty(t, node.ID())
}

func TestIPFS_ConnectToBootstrap(t *testing.T) {
	ctx := context.Background()

	node, err := storage.NewIPFSNode(ctx, &storage.IPFSConfig{
		RepoPath: t.TempDir(),
		Bootstrap: []string{
			"/dnsaddr/bootstrap.libp2p.io/p2p/QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN",
		},
	})
	require.NoError(t, err)
	defer node.Close()

	// Should connect to at least one bootstrap peer
	assert.Eventually(t, func() bool {
		return len(node.Peers()) > 0
	}, 5*time.Second, 100*time.Millisecond)
}

func TestIPFS_ShareWithLibp2pHost(t *testing.T) {
	// IPFS node should share DHT with existing libp2p host
	ctx := context.Background()

	libp2pHost, _ := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	defer libp2pHost.Close()

	node, err := storage.NewIPFSNodeWithHost(ctx, libp2pHost, &storage.IPFSConfig{
		RepoPath: t.TempDir(),
	})
	require.NoError(t, err)
	defer node.Close()

	// Should have a valid node ID
	nodeID := node.ID()
	assert.NotEmpty(t, nodeID)
	assert.Contains(t, nodeID, "12D3KooW") // Should be a valid libp2p peer ID format
}
