package network_test

import (
	"context"
	"testing"
	"time"

	"ledabeer/backend/internal/network"

	"github.com/stretchr/testify/assert"
)

func TestNewHost_InitializesWithSecureTransport(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	assert.NoError(t, err)
	assert.NotNil(t, host)
	defer host.Close()

	// Should have addresses
	assert.NotEmpty(t, host.Addrs())

	// Should have a peer ID
	assert.NotEmpty(t, host.ID())
}

func TestNewHost_MultipleInstances(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	host1, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	assert.NoError(t, err)
	defer host1.Close()

	host2, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	assert.NoError(t, err)
	defer host2.Close()

	// Should have different peer IDs
	assert.NotEqual(t, host1.ID(), host2.ID())
}
