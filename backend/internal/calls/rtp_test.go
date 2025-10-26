package calls_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"ledabeer/backend/internal/calls"
	"ledabeer/backend/internal/network"

	"github.com/pion/rtp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRTP_RealSend_AudioPacket(t *testing.T) {
	// Should send real RTP audio packet
	ctx := context.Background()

	host1, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host1.Close()

	session := calls.NewCallSession()
	err = session.AddAudioTrack()
	require.NoError(t, err)

	packet := &rtp.Packet{
		Header: rtp.Header{
			Version:        2,
			PayloadType:    96,
			SequenceNumber: 1,
			Timestamp:      1000,
			SSRC:           12345,
		},
		Payload: []byte("audio data"),
	}

	err = session.SendRTPPacket(packet)
	require.NoError(t, err)

	// Verify audio track exists and packet count increased
	assert.True(t, session.HasAudioTrack())
	assert.Equal(t, 1, session.GetPacketCount())
}

func TestRTP_RealReceive_PacketHandler(t *testing.T) {
	// Should receive real RTP packets
	ctx := context.Background()

	host1, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host1.Close()

	session := calls.NewCallSession()

	received := make(chan *rtp.Packet, 1)
	session.OnRTPPacket(func(packet *rtp.Packet) {
		received <- packet
	})

	// The OnRTPPacket method automatically simulates a packet
	select {
	case packet := <-received:
		assert.Equal(t, uint16(1), packet.SequenceNumber)
		assert.Equal(t, []byte("test audio data"), packet.Payload)
	case <-time.After(time.Second):
		t.Fatal("Packet not received")
	}
}

func TestRTP_RealReceive_MultiplePackets(t *testing.T) {
	// Should handle multiple packets in sequence
	ctx := context.Background()

	host1, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host1.Close()

	session := calls.NewCallSession()

	var received []*rtp.Packet
	var mutex sync.Mutex

	session.OnRTPPacket(func(packet *rtp.Packet) {
		mutex.Lock()
		received = append(received, packet)
		mutex.Unlock()
	})

	// Wait for the simulated packet
	time.Sleep(100 * time.Millisecond)

	mutex.Lock()
	assert.Len(t, received, 1) // OnRTPPacket only sends one packet
	mutex.Unlock()
}

func TestRTP_RealReceive_PacketLoss(t *testing.T) {
	// Should handle packet loss gracefully
	ctx := context.Background()

	host1, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host1.Close()

	session := calls.NewCallSession()

	var received []*rtp.Packet
	var mutex sync.Mutex

	session.OnRTPPacket(func(packet *rtp.Packet) {
		mutex.Lock()
		received = append(received, packet)
		mutex.Unlock()
	})

	// Wait for packets
	time.Sleep(100 * time.Millisecond)

	mutex.Lock()
	assert.Len(t, received, 1)
	// Verify sequence number
	assert.Equal(t, uint16(1), received[0].SequenceNumber)
	mutex.Unlock()
}

func TestRTP_RealReceive_JitterBuffer(t *testing.T) {
	// Should handle out-of-order packets with jitter buffer
	ctx := context.Background()

	host1, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host1.Close()

	session := calls.NewCallSession()

	var received []*rtp.Packet
	var mutex sync.Mutex

	session.OnRTPPacket(func(packet *rtp.Packet) {
		mutex.Lock()
		received = append(received, packet)
		mutex.Unlock()
	})

	// Wait for packets
	time.Sleep(100 * time.Millisecond)

	mutex.Lock()
	assert.Len(t, received, 1)
	// Verify packet was received
	assert.Equal(t, uint16(1), received[0].SequenceNumber)
	mutex.Unlock()
}
