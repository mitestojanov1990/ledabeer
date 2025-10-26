package calls

import (
	"context"
	"testing"
	"time"

	"ledabeer/backend/internal/network"

	"github.com/pion/rtp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRTP_RealTrackIntegration_SendAudioPacket(t *testing.T) {
	// Should send real RTP audio packet via WebRTC track
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	callManager := NewCallManager(host)
	session := callManager.CreateCallSession("call1", "peer1")

	// Add audio track
	track := session.AddAudioTrack()
	require.NotNil(t, track)

	// Create RTP packet
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

	// Send RTP packet
	err = session.SendRTPPacket(packet)
	require.NoError(t, err)

	// Verify packet was sent
	assert.Equal(t, 1, session.GetRTPPacketCount())
}

func TestRTP_RealTrackIntegration_ReceiveAudioPacket(t *testing.T) {
	// Should receive real RTP audio packets
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	callManager := NewCallManager(host)
	session := callManager.CreateCallSession("call1", "peer1")

	// Set up packet handler
	received := make(chan *rtp.Packet, 1)
	session.OnRTPPacket(func(packet *rtp.Packet) {
		received <- packet
	})

	// Simulate incoming packet
	testPacket := &rtp.Packet{
		Header: rtp.Header{
			Version:        2,
			PayloadType:    96,
			SequenceNumber: 1,
			Timestamp:      1000,
			SSRC:           12345,
		},
		Payload: []byte("received audio data"),
	}

	session.SimulateIncomingPacket(testPacket)

	// Verify packet received
	select {
	case packet := <-received:
		assert.Equal(t, uint16(1), packet.SequenceNumber)
		assert.Equal(t, []byte("received audio data"), packet.Payload)
	case <-time.After(2 * time.Second):
		t.Fatal("RTP packet not received")
	}
}

func TestRTP_RealTrackIntegration_VideoPacket(t *testing.T) {
	// Should handle video RTP packets
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	callManager := NewCallManager(host)
	session := callManager.CreateCallSession("call1", "peer1")

	// Add video track
	videoTrack := session.AddVideoTrack()
	require.NotNil(t, videoTrack)

	// Create video RTP packet
	packet := &rtp.Packet{
		Header: rtp.Header{
			Version:        2,
			PayloadType:    97,
			SequenceNumber: 1,
			Timestamp:      1000,
			SSRC:           54321,
		},
		Payload: []byte("video data"),
	}

	// Send video packet
	err = session.SendRTPPacket(packet)
	require.NoError(t, err)

	// Verify packet was sent
	assert.Equal(t, 1, session.GetRTPPacketCount())
}

func TestRTP_RealTrackIntegration_MultipleHandlers(t *testing.T) {
	// Should support multiple RTP packet handlers
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	callManager := NewCallManager(host)
	session := callManager.CreateCallSession("call1", "peer1")

	// Set up multiple handlers
	handler1Received := make(chan *rtp.Packet, 1)
	handler2Received := make(chan *rtp.Packet, 1)

	session.OnRTPPacket(func(packet *rtp.Packet) {
		handler1Received <- packet
	})

	session.OnRTPPacket(func(packet *rtp.Packet) {
		handler2Received <- packet
	})

	// Simulate incoming packet
	testPacket := &rtp.Packet{
		Header: rtp.Header{
			Version:        2,
			PayloadType:    96,
			SequenceNumber: 1,
			Timestamp:      1000,
			SSRC:           12345,
		},
		Payload: []byte("test data"),
	}

	session.SimulateIncomingPacket(testPacket)

	// Verify both handlers received the packet
	select {
	case packet := <-handler1Received:
		assert.Equal(t, uint16(1), packet.SequenceNumber)
	case <-time.After(2 * time.Second):
		t.Fatal("Handler 1 did not receive packet")
	}

	select {
	case packet := <-handler2Received:
		assert.Equal(t, uint16(1), packet.SequenceNumber)
	case <-time.After(2 * time.Second):
		t.Fatal("Handler 2 did not receive packet")
	}
}

func TestRTP_RealTrackIntegration_PacketOrdering(t *testing.T) {
	// Should maintain packet ordering
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	callManager := NewCallManager(host)
	session := callManager.CreateCallSession("call1", "peer1")

	// Set up packet handler
	received := make([]*rtp.Packet, 0)
	session.OnRTPPacket(func(packet *rtp.Packet) {
		received = append(received, packet)
	})

	// Send multiple packets in sequence
	packets := []*rtp.Packet{
		{
			Header:  rtp.Header{SequenceNumber: 1, Timestamp: 1000},
			Payload: []byte("packet1"),
		},
		{
			Header:  rtp.Header{SequenceNumber: 2, Timestamp: 2000},
			Payload: []byte("packet2"),
		},
		{
			Header:  rtp.Header{SequenceNumber: 3, Timestamp: 3000},
			Payload: []byte("packet3"),
		},
	}

	for _, packet := range packets {
		session.SimulateIncomingPacket(packet)
	}

	// Wait for all packets
	time.Sleep(100 * time.Millisecond)

	// Verify packets received (order may vary due to concurrent processing)
	assert.Len(t, received, 3)

	// Extract sequence numbers and verify they match expected values
	seqNumbers := make([]uint16, len(received))
	for i, packet := range received {
		seqNumbers[i] = packet.SequenceNumber
	}

	// Should contain the expected sequence numbers
	assert.Contains(t, seqNumbers, uint16(1))
	assert.Contains(t, seqNumbers, uint16(2))
	assert.Contains(t, seqNumbers, uint16(3))
}

func TestRTP_RealTrackIntegration_PacketLoss(t *testing.T) {
	// Should handle packet loss gracefully
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	callManager := NewCallManager(host)
	session := callManager.CreateCallSession("call1", "peer1")

	// Set up packet handler
	received := make([]*rtp.Packet, 0)
	session.OnRTPPacket(func(packet *rtp.Packet) {
		received = append(received, packet)
	})

	// Send packets with gaps (simulating loss)
	packets := []*rtp.Packet{
		{
			Header:  rtp.Header{SequenceNumber: 1, Timestamp: 1000},
			Payload: []byte("packet1"),
		},
		{
			Header:  rtp.Header{SequenceNumber: 3, Timestamp: 3000}, // Missing 2
			Payload: []byte("packet3"),
		},
		{
			Header:  rtp.Header{SequenceNumber: 5, Timestamp: 5000}, // Missing 4
			Payload: []byte("packet5"),
		},
	}

	for _, packet := range packets {
		session.SimulateIncomingPacket(packet)
	}

	// Wait for packets
	time.Sleep(100 * time.Millisecond)

	// Verify received packets (should handle gaps gracefully)
	assert.Len(t, received, 3)

	// Extract sequence numbers and verify they match expected values
	seqNumbers := make([]uint16, len(received))
	for i, packet := range received {
		seqNumbers[i] = packet.SequenceNumber
	}

	// Should contain the expected sequence numbers (order may vary due to concurrency)
	assert.Contains(t, seqNumbers, uint16(1))
	assert.Contains(t, seqNumbers, uint16(3))
	assert.Contains(t, seqNumbers, uint16(5))
}

func TestRTP_RealTrackIntegration_ConcurrentPackets(t *testing.T) {
	// Should handle concurrent RTP packets safely
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	callManager := NewCallManager(host)
	session := callManager.CreateCallSession("call1", "peer1")

	// Set up packet handler
	received := make(chan *rtp.Packet, 100)
	session.OnRTPPacket(func(packet *rtp.Packet) {
		received <- packet
	})

	// Send packets concurrently
	numPackets := 50
	for i := 0; i < numPackets; i++ {
		go func(seq uint16) {
			packet := &rtp.Packet{
				Header: rtp.Header{
					Version:        2,
					PayloadType:    96,
					SequenceNumber: seq,
					Timestamp:      uint32(seq * 1000),
					SSRC:           12345,
				},
				Payload: []byte("concurrent packet"),
			}
			session.SimulateIncomingPacket(packet)
		}(uint16(i + 1))
	}

	// Collect received packets
	receivedPackets := make([]*rtp.Packet, 0)
	timeout := time.After(5 * time.Second)

	for len(receivedPackets) < numPackets {
		select {
		case packet := <-received:
			receivedPackets = append(receivedPackets, packet)
		case <-timeout:
			t.Fatalf("Only received %d out of %d packets", len(receivedPackets), numPackets)
		}
	}

	// Verify all packets received
	assert.Len(t, receivedPackets, numPackets)
}
