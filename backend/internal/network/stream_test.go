package network_test

import (
	"context"
	"testing"
	"time"

	"ledabeer/backend/internal/network"

	"github.com/stretchr/testify/assert"
)

func TestStream_SendReceiveMessage(t *testing.T) {
	// Setup two hosts
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	host1, _ := network.NewHost(ctx, &network.Config{ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"}})
	host2, _ := network.NewHost(ctx, &network.Config{ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"}})
	defer host1.Close()
	defer host2.Close()

	// Setup stream handler on host2
	received := make(chan []byte, 1)
	handler := network.NewStreamHandler(func(data []byte) error {
		received <- data
		return nil
	})
	host2.SetStreamHandler("/chat/1.0.0", handler.Handle)

	// Connect hosts
	addrs := host2.Addrs()
	peerInfo := host2.Peerstore().PeerInfo(host2.ID())
	err := host1.Connect(ctx, peerInfo)
	assert.NoError(t, err)
	_ = addrs // Use addrs to avoid unused variable

	// Send from host1
	stream, err := host1.NewStream(ctx, host2.ID(), "/chat/1.0.0")
	assert.NoError(t, err)

	testData := []byte("test message")
	err = network.WriteMessage(stream, testData)
	assert.NoError(t, err)

	// Verify receipt
	select {
	case msg := <-received:
		assert.Equal(t, testData, msg)
	case <-time.After(2 * time.Second):
		t.Fatal("Message not received")
	}
}

func TestStream_MultipleMessages(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	host1, _ := network.NewHost(ctx, &network.Config{ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"}})
	host2, _ := network.NewHost(ctx, &network.Config{ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"}})
	defer host1.Close()
	defer host2.Close()

	// Setup stream handler
	receivedMsgs := make(chan []byte, 10)
	handler := network.NewStreamHandler(func(data []byte) error {
		receivedMsgs <- data
		return nil
	})
	host2.SetStreamHandler("/chat/1.0.0", handler.Handle)

	// Connect
	peerInfo := host2.Peerstore().PeerInfo(host2.ID())
	err := host1.Connect(ctx, peerInfo)
	assert.NoError(t, err)

	// Send multiple messages
	messages := []string{"msg1", "msg2", "msg3"}
	for _, msg := range messages {
		stream, err := host1.NewStream(ctx, host2.ID(), "/chat/1.0.0")
		assert.NoError(t, err)

		err = network.WriteMessage(stream, []byte(msg))
		assert.NoError(t, err)
		stream.Close()
	}

	// Verify all received
	for i, expected := range messages {
		select {
		case msg := <-receivedMsgs:
			assert.Equal(t, expected, string(msg), "Message %d mismatch", i)
		case <-time.After(2 * time.Second):
			t.Fatalf("Message %d not received", i)
		}
	}
}
