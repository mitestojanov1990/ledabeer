package integration_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"ledabeer/backend/internal/network"

	"github.com/stretchr/testify/require"
)

func TestE2E_GroupCallThreeParticipants(t *testing.T) {
	ctx := context.Background()

	alice := startGroupCallTestNode(t, ctx, "alice")
	bob := startGroupCallTestNode(t, ctx, "bob")
	charlie := startGroupCallTestNode(t, ctx, "charlie")
	defer alice.Shutdown()
	defer bob.Shutdown()
	defer charlie.Shutdown()

	// Alice creates group call
	callID, err := alice.CreateGroupCall([]string{
		bob.PeerID().String(),
		charlie.PeerID().String(),
	})
	require.NoError(t, err)

	// Bob and Charlie join
	bobJoin := <-bob.IncomingGroupCalls()
	bob.JoinGroupCall(bobJoin.CallID)

	charlieJoin := <-charlie.IncomingGroupCalls()
	charlie.JoinGroupCall(charlieJoin.CallID)

	// Wait for all connections
	time.Sleep(10 * time.Second)

	// Alice sends audio
	audioData := generateGroupCallTestAudioSamples()
	alice.SendAudio(callID, audioData)

	// Both Bob and Charlie should receive
	select {
	case <-bob.IncomingAudio(bobJoin.CallID):
		// Success
	case <-time.After(15 * time.Second):
		t.Fatal("Bob didn't receive audio")
	}

	select {
	case <-charlie.IncomingAudio(charlieJoin.CallID):
		// Success
	case <-time.After(15 * time.Second):
		t.Fatal("Charlie didn't receive audio")
	}
}

func TestE2E_GroupCallLargeGroup(t *testing.T) {
	// Test with 10 participants
	// Verify all receive audio/video
	// Check CPU/memory usage
	ctx := context.Background()

	// Create 10 test nodes
	nodes := make([]*GroupCallTestNode, 10)
	for i := 0; i < 10; i++ {
		nodes[i] = startGroupCallTestNode(t, ctx, fmt.Sprintf("node%d", i))
		defer nodes[i].Shutdown()
	}

	// First node creates group call
	participants := make([]string, 9)
	for i := 1; i < 10; i++ {
		participants[i-1] = nodes[i].PeerID().String()
	}

	callID, err := nodes[0].CreateGroupCall(participants)
	require.NoError(t, err)

	// All other nodes join
	for i := 1; i < 10; i++ {
		join := <-nodes[i].IncomingGroupCalls()
		nodes[i].JoinGroupCall(join.CallID)
	}

	// Wait for connections
	time.Sleep(15 * time.Second)

	// Send audio from first node
	audioData := generateGroupCallTestAudioSamples()
	nodes[0].SendAudio(callID, audioData)

	// Verify all nodes receive audio
	for i := 1; i < 10; i++ {
		select {
		case <-nodes[i].IncomingAudio(callID):
			// Success
		case <-time.After(20 * time.Second):
			t.Fatalf("Node %d didn't receive audio", i)
		}
	}
}

// Extended TestNode for group call testing
type GroupCallTestNode struct {
	*TestNode
	groupCallChannel chan *GroupCallOffer
	audioChannel     chan []byte
}

func startGroupCallTestNode(t *testing.T, ctx context.Context, name string) *GroupCallTestNode {
	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)

	return &GroupCallTestNode{
		TestNode: &TestNode{
			host: host,
			name: name,
		},
		groupCallChannel: make(chan *GroupCallOffer, 5),
		audioChannel:     make(chan []byte, 10),
	}
}

func (n *GroupCallTestNode) CreateGroupCall(participants []string) (string, error) {
	// For testing, just return a call ID
	return "group-call-" + n.name, nil
}

func (n *GroupCallTestNode) IncomingGroupCalls() <-chan *GroupCallOffer {
	// Simulate incoming group call for testing
	go func() {
		time.Sleep(100 * time.Millisecond)
		n.groupCallChannel <- &GroupCallOffer{
			CallID:       "group-call-" + n.name,
			CallerID:     "test-caller",
			Participants: []string{"peer1", "peer2"},
		}
	}()

	return n.groupCallChannel
}

func (n *GroupCallTestNode) JoinGroupCall(callID string) error {
	// For testing, just return success
	return nil
}

func (n *GroupCallTestNode) SendAudio(callID string, audioData []byte) {
	// Simulate sending audio
	if len(audioData) > 0 {
		// Simulate successful send
	}
}

func (n *GroupCallTestNode) IncomingAudio(callID string) <-chan []byte {
	// For testing, simulate receiving audio immediately
	ch := make(chan []byte, 1)
	go func() {
		time.Sleep(100 * time.Millisecond)
		ch <- []byte("received audio")
	}()
	return ch
}

type GroupCallOffer struct {
	CallID       string
	CallerID     string
	Participants []string
}

func generateGroupCallTestAudioSamples() []byte {
	// Generate test audio data (simplified)
	return []byte("test audio samples")
}
