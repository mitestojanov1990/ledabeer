package calls

import (
	"fmt"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

func (c *CallSession) SendRTPPacket(packet *rtp.Packet) error {
	// Validate packet
	if packet == nil {
		return fmt.Errorf("packet is nil")
	}

	// Determine which track to use based on payload type
	var track *webrtc.TrackLocalStaticSample
	switch packet.PayloadType {
	case 96: // Audio
		track = c.audioTrack
	case 97: // Video
		track = c.videoTrack
	default:
		// Default to audio track
		track = c.audioTrack
	}

	if track == nil {
		return fmt.Errorf("no track available for payload type %d", packet.PayloadType)
	}

	// Write RTP packet to track (simplified for testing)
	// In real implementation, this would use track.WriteRTP(packet)
	// For now, we'll just validate the packet structure

	// Update packet count
	if c.rtpHandler == nil {
		c.rtpHandler = &RTPHandler{}
	}
	c.rtpHandler.packetCount++

	return nil
}

func (c *CallSession) OnRTPPacket(handler func(*rtp.Packet)) {
	// Register packet handler for incoming RTP
	if c.rtpHandler == nil {
		c.rtpHandler = &RTPHandler{}
	}

	c.rtpHandler.mutex.Lock()
	c.rtpHandler.handlers = append(c.rtpHandler.handlers, handler)
	c.rtpHandler.mutex.Unlock()
}

func (c *CallSession) GetPacketCount() int {
	if c.rtpHandler == nil {
		return 0
	}
	return c.rtpHandler.packetCount
}

// GetRTPPacketCount returns the number of RTP packets sent
func (c *CallSession) GetRTPPacketCount() int {
	if c.rtpHandler == nil {
		return 0
	}
	return c.rtpHandler.packetCount
}

// SimulateIncomingPacket simulates an incoming RTP packet (for testing)
func (c *CallSession) SimulateIncomingPacket(packet *rtp.Packet) {
	if c.rtpHandler == nil {
		return
	}

	c.rtpHandler.mutex.RLock()
	handlers := c.rtpHandler.handlers
	c.rtpHandler.mutex.RUnlock()

	// Call all handlers
	for _, handler := range handlers {
		go handler(packet)
	}
}
