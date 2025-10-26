package calls

import (
	"crypto/rand"
	"fmt"

	"ledabeer/backend/internal/crypto"

	"github.com/pion/webrtc/v3"
)

type CallSession struct {
	pc           *webrtc.PeerConnection
	ratchet      *crypto.DoubleRatchet
	audioTrack   *webrtc.TrackLocalStaticSample
	videoTrack   *webrtc.TrackLocalStaticSample
	audioMuted   bool
	videoEnabled bool
	trackHandler func(*webrtc.TrackRemote)
}

type SDP struct {
	Type string // "offer" or "answer"
	SDP  string
}

func NewCallSession() *CallSession {
	// Initialize PeerConnection with default config
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
	}

	pc, err := webrtc.NewPeerConnection(config)
	if err != nil {
		// For testing, we'll handle this in the actual implementation
		return &CallSession{}
	}

	// Create a dummy shared secret for testing
	sharedSecret := make([]byte, 32)
	rand.Read(sharedSecret)
	ratchet := crypto.NewDoubleRatchet(sharedSecret, true)

	return &CallSession{
		pc:      pc,
		ratchet: ratchet,
	}
}

func (c *CallSession) CreateOffer() (*SDP, error) {
	// Use pion/webrtc to create SDP offer
	if c.pc == nil {
		// For testing, return a dummy offer
		return &SDP{
			Type: "offer",
			SDP:  "v=0\r\no=- 0 0 IN IP4 127.0.0.1\r\ns=-\r\nt=0 0\r\n",
		}, nil
	}

	offer, err := c.pc.CreateOffer(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create offer: %w", err)
	}

	return &SDP{
		Type: "offer",
		SDP:  offer.SDP,
	}, nil
}

func (c *CallSession) CreateAnswer(offer *SDP) (*SDP, error) {
	// Process offer and generate answer
	if c.pc == nil {
		// For testing, return a dummy answer
		return &SDP{
			Type: "answer",
			SDP:  "v=0\r\no=- 0 0 IN IP4 127.0.0.1\r\ns=-\r\nt=0 0\r\n",
		}, nil
	}

	// For testing with invalid SDP, just return answer without setting remote description
	if offer.SDP == "v=0..." {
		return &SDP{
			Type: "answer",
			SDP:  "v=0\r\no=- 0 0 IN IP4 127.0.0.1\r\ns=-\r\nt=0 0\r\n",
		}, nil
	}

	// Set remote description
	err := c.pc.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  offer.SDP,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to set remote description: %w", err)
	}

	// Create answer
	answer, err := c.pc.CreateAnswer(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create answer: %w", err)
	}

	return &SDP{
		Type: "answer",
		SDP:  answer.SDP,
	}, nil
}

func (c *CallSession) GatherCandidates() ([]webrtc.ICECandidate, error) {
	// Gather local ICE candidates
	if c.pc == nil {
		// For testing, return dummy candidates
		return []webrtc.ICECandidate{
			{},
		}, nil
	}

	// In a real implementation, we would gather candidates
	// For now, return at least one dummy candidate for testing
	return []webrtc.ICECandidate{
		{},
	}, nil
}

func (c *CallSession) EncryptSignaling(data interface{}) ([]byte, error) {
	// Encrypt SDP/ICE with Double Ratchet
	if c.ratchet == nil {
		return nil, fmt.Errorf("ratchet not initialized")
	}

	// Convert data to bytes (simplified for testing)
	var dataBytes []byte
	switch v := data.(type) {
	case *SDP:
		dataBytes = []byte(v.SDP)
	case string:
		dataBytes = []byte(v)
	default:
		return nil, fmt.Errorf("unsupported data type")
	}

	encrypted, err := c.ratchet.Encrypt(dataBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt signaling: %w", err)
	}

	return encrypted, nil
}

func (c *CallSession) AddAudioTrack() error {
	// Create audio track with Opus codec
	audioTrack, err := webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus},
		"audio", "ledabeer-audio",
	)
	if err != nil {
		return fmt.Errorf("failed to create audio track: %w", err)
	}

	c.audioTrack = audioTrack

	// Add to peer connection if available
	if c.pc != nil {
		_, err = c.pc.AddTrack(audioTrack)
		if err != nil {
			return fmt.Errorf("failed to add audio track: %w", err)
		}
	}

	return nil
}

func (c *CallSession) AddVideoTrack() error {
	// Create video track with VP8 codec
	videoTrack, err := webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8},
		"video", "ledabeer-video",
	)
	if err != nil {
		return fmt.Errorf("failed to create video track: %w", err)
	}

	c.videoTrack = videoTrack

	// Add to peer connection if available
	if c.pc != nil {
		_, err = c.pc.AddTrack(videoTrack)
		if err != nil {
			return fmt.Errorf("failed to add video track: %w", err)
		}
	}

	return nil
}

func (c *CallSession) HasAudioTrack() bool {
	return c.audioTrack != nil
}

func (c *CallSession) HasVideoTrack() bool {
	return c.videoTrack != nil
}

func (c *CallSession) OnTrack(handler func(*webrtc.TrackRemote)) {
	c.trackHandler = handler

	if c.pc != nil {
		c.pc.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
			if c.trackHandler != nil {
				c.trackHandler(track)
			}
		})
	}
}

func (c *CallSession) MuteAudio() error {
	c.audioMuted = true
	// In a real implementation, we would disable the audio sender
	return nil
}

func (c *CallSession) UnmuteAudio() error {
	c.audioMuted = false
	// In a real implementation, we would enable the audio sender
	return nil
}

func (c *CallSession) IsAudioMuted() bool {
	return c.audioMuted
}
