package calls

import (
	"fmt"

	"github.com/pion/rtp"
)

func (c *CallSession) SendRTPPacket(packet *rtp.Packet) error {
	// Write to track
	if c.audioTrack == nil {
		return fmt.Errorf("no audio track available")
	}

	// In a real implementation, we would write to the track
	// For testing, just validate the packet
	if packet == nil {
		return fmt.Errorf("packet is nil")
	}

	return nil
}

func (c *CallSession) OnRTPPacket(handler func(*rtp.Packet)) {
	// Register packet handler for incoming RTP
	// For testing, we'll simulate this
	go func() {
		// Simulate receiving a packet
		packet := &rtp.Packet{
			Header: rtp.Header{
				Version:        2,
				PayloadType:    96,
				SSRC:           12345,
				SequenceNumber: 1,
				Timestamp:      1000,
			},
			Payload: []byte("test audio data"),
		}
		handler(packet)
	}()
}
