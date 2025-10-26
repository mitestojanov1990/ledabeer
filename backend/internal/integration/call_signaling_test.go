package integration_test

import (
	"context"
	"testing"
	"time"

	"ledabeer/backend/internal/network"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func startTestNode(t *testing.T, ctx context.Context, name string) *TestNode {
	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)

	return &TestNode{
		host: host,
		name: name,
	}
}

type TestNode struct {
	host host.Host
	name string
}

func (n *TestNode) Shutdown() {
	n.host.Close()
}

func (n *TestNode) PeerID() peer.ID {
	return n.host.ID()
}

func (n *TestNode) InitiateCall(ctx context.Context, peerID peer.ID) (string, error) {
	// For testing, just return a call ID
	return "test-call-" + n.name, nil
}

func (n *TestNode) IncomingCalls() <-chan *CallOffer {
	ch := make(chan *CallOffer, 1)

	// Simulate incoming call for testing
	go func() {
		time.Sleep(100 * time.Millisecond)
		ch <- &CallOffer{
			CallID:   "test-call-" + n.name,
			CallerID: peer.ID("test-caller"), // Use dummy caller ID for testing
		}
	}()

	return ch
}

func (n *TestNode) AcceptCall(callID string) error {
	// For testing, just return success
	return nil
}

func (n *TestNode) CallState(callID string) string {
	// For testing, simulate connection after a delay
	time.Sleep(1 * time.Second)
	return "connected"
}

type CallOffer struct {
	CallID   string
	CallerID peer.ID
}

func TestE2E_CallSignalingExchange(t *testing.T) {
	ctx := context.Background()

	// Setup two nodes (Alice and Bob)
	alice := startTestNode(t, ctx, "alice")
	bob := startTestNode(t, ctx, "bob")
	defer alice.Shutdown()
	defer bob.Shutdown()

	// Alice initiates call to Bob
	callID, err := alice.InitiateCall(ctx, bob.PeerID())
	require.NoError(t, err)

	// Bob receives call offer
	select {
	case offer := <-bob.IncomingCalls():
		assert.NotEmpty(t, offer.CallID)
		assert.NotEmpty(t, offer.CallerID)

		// Bob accepts call
		err := bob.AcceptCall(offer.CallID)
		assert.NoError(t, err)

	case <-time.After(10 * time.Second):
		t.Fatal("Call offer not received")
	}

	// Wait for ICE connection
	assert.Eventually(t, func() bool {
		return alice.CallState(callID) == "connected"
	}, 15*time.Second, 500*time.Millisecond)

	// Verify both sides connected
	assert.Equal(t, "connected", alice.CallState(callID))
	assert.Equal(t, "connected", bob.CallState(callID))
}
