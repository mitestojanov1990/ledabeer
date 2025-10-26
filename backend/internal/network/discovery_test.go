package network_test

import (
	"context"
	"testing"
	"time"

	"ledabeer/backend/internal/network"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/stretchr/testify/assert"
)

func TestDiscovery_WithBootstrapPeers(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create bootstrap node
	bootstrap, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	assert.NoError(t, err)
	defer bootstrap.Close()

	// Create client node with bootstrap peer
	bootstrapAddr := bootstrap.Addrs()[0].String() + "/p2p/" + bootstrap.ID().String()
	client, err := network.NewHost(ctx, &network.Config{
		ListenAddrs:    []string{"/ip4/127.0.0.1/tcp/0"},
		BootstrapPeers: []string{bootstrapAddr},
	})
	assert.NoError(t, err)
	defer client.Close()

	// Client should be able to find bootstrap peer
	time.Sleep(2 * time.Second) // Allow discovery time

	// Check if client knows about bootstrap peer
	peerInfo := client.Peerstore().PeerInfo(bootstrap.ID())
	assert.NotEmpty(t, peerInfo.Addrs, "Client should know bootstrap peer addresses")
}

func TestDiscovery_WithoutBootstrapPeers(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create isolated node
	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	assert.NoError(t, err)
	defer host.Close()

	// Should start successfully even without bootstrap peers
	assert.NotNil(t, host)
	assert.NotEmpty(t, host.ID())
}

func TestDiscovery_MultipleNodes(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Create 3 nodes
	nodes := make([]host.Host, 3)
	for i := 0; i < 3; i++ {
		host, err := network.NewHost(ctx, &network.Config{
			ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
		})
		assert.NoError(t, err)
		nodes[i] = host
		defer host.Close()
	}

	// Connect nodes in a chain: 0 -> 1 -> 2
	for i := 0; i < len(nodes)-1; i++ {
		peerInfo := nodes[i+1].Peerstore().PeerInfo(nodes[i+1].ID())
		err := nodes[i].Connect(ctx, peerInfo)
		assert.NoError(t, err)
	}

	// Wait for discovery
	time.Sleep(3 * time.Second)

	// All nodes should be connected
	for i := 0; i < len(nodes); i++ {
		connectedPeers := nodes[i].Network().Peers()
		assert.GreaterOrEqual(t, len(connectedPeers), 1, "Node %d should have at least 1 peer", i)
	}
}
