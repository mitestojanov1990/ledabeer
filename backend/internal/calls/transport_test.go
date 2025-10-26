package calls_test

import (
	"context"
	"testing"
	"time"

	"ledabeer/backend/internal/calls"
	"ledabeer/backend/internal/network"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/pion/webrtc/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestHosts(t *testing.T) (host.Host, host.Host) {
	ctx := context.Background()

	host1, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)

	host2, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)

	// Connect hosts
	host2Addr := host2.Addrs()[0].String() + "/p2p/" + host2.ID().String()
	peerInfo, err := network.ParseAddrInfo(host2Addr)
	require.NoError(t, err)
	err = host1.Connect(ctx, *peerInfo)
	require.NoError(t, err)

	return host1, host2
}

func TestSignalingTransport_SendOffer(t *testing.T) {
	// Setup two hosts
	host1, host2 := setupTestHosts(t)
	defer host1.Close()
	defer host2.Close()

	// Setup receiver
	receiver := calls.NewCallTransport(host2)
	receiver.HandleOffer(func(offer *calls.SDP) (*calls.SDP, error) {
		return &calls.SDP{Type: "answer", SDP: "response"}, nil
	})

	// Send encrypted offer from host1 to host2
	caller := calls.NewCallTransport(host1)
	offer := &calls.SDP{Type: "offer", SDP: "test"}
	err := caller.SendOffer(context.Background(), host2.ID(), offer)
	assert.NoError(t, err)
}

func TestSignalingTransport_ReceiveAnswer(t *testing.T) {
	// Caller should receive answer after sending offer
	host1, host2 := setupTestHosts(t)
	defer host1.Close()
	defer host2.Close()

	caller := calls.NewCallTransport(host1)
	callee := calls.NewCallTransport(host2)

	// Setup answer handler
	callee.HandleOffer(func(offer *calls.SDP) (*calls.SDP, error) {
		return &calls.SDP{Type: "answer", SDP: "response"}, nil
	})

	// Send offer
	offer := &calls.SDP{Type: "offer", SDP: "test"}
	err := caller.SendOffer(context.Background(), host2.ID(), offer)
	require.NoError(t, err)

	// Receive answer
	select {
	case answer := <-caller.AnswerChannel():
		assert.Equal(t, "answer", answer.Type)
	case <-time.After(5 * time.Second):
		t.Fatal("No answer received")
	}
}

func TestSignalingTransport_ICETrickle(t *testing.T) {
	// Should stream ICE candidates as they're discovered
	host1, host2 := setupTestHosts(t)
	defer host1.Close()
	defer host2.Close()

	transport := calls.NewCallTransport(host1)
	candidates := make(chan *webrtc.ICECandidate, 10)

	transport.OnICECandidate(func(c *webrtc.ICECandidate) {
		candidates <- c
	})

	// Simulate candidate discovery
	transport.AddICECandidate(host2.ID(), &webrtc.ICECandidate{})

	// Verify received
	select {
	case c := <-candidates:
		assert.NotNil(t, c)
	case <-time.After(2 * time.Second):
		t.Fatal("Candidate not received")
	}
}
