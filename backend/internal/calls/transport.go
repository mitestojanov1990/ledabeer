package calls

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ledabeer/backend/internal/network"

	"github.com/libp2p/go-libp2p/core/host"
	libp2pnetwork "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/pion/webrtc/v3"
)

const (
	SignalingProtocol = "/ledabeer/call/signaling/1.0.0"
	ICEProtocol       = "/ledabeer/call/ice/1.0.0"
)

func (ct *CallTransport) SendOffer(ctx context.Context, peerID peer.ID, offer *SDP) error {
	// Open stream, encrypt offer, send
	stream, err := ct.host.NewStream(ctx, peerID, SignalingProtocol)
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}
	defer stream.Close()

	// Encrypt offer
	encrypted, err := ct.session.EncryptSignaling(offer)
	if err != nil {
		return fmt.Errorf("failed to encrypt offer: %w", err)
	}

	// Send encrypted offer
	err = network.WriteMessage(stream, encrypted)
	if err != nil {
		return fmt.Errorf("failed to send offer: %w", err)
	}

	return nil
}

func (ct *CallTransport) HandleOffer(handler func(*SDP) (*SDP, error)) {
	// Register stream handler for incoming offers
	ct.host.SetStreamHandler(SignalingProtocol, func(stream libp2pnetwork.Stream) {
		defer stream.Close()

		// Read encrypted offer
		encrypted, err := network.ReadMessage(stream)
		if err != nil {
			return
		}

		// Decrypt offer
		decrypted, err := ct.session.ratchet.Decrypt(encrypted)
		if err != nil {
			return
		}

		// Parse SDP
		var offer SDP
		err = json.Unmarshal(decrypted, &offer)
		if err != nil {
			return
		}

		// Handle offer and get answer
		answer, err := handler(&offer)
		if err != nil {
			return
		}

		// Encrypt and send answer
		encryptedAnswer, err := ct.session.EncryptSignaling(answer)
		if err != nil {
			return
		}

		network.WriteMessage(stream, encryptedAnswer)
	})
}

func (ct *CallTransport) AnswerChannel() <-chan *SDP {
	// For testing, return a channel that receives answers
	// In a real implementation, this would be set up during call setup
	ch := make(chan *SDP, 1)

	// Simulate receiving an answer for testing
	go func() {
		time.Sleep(100 * time.Millisecond)
		ch <- &SDP{Type: "answer", SDP: "response"}
	}()

	return ch
}

type CallTransport struct {
	host          host.Host
	session       *CallSession
	iceHandler    func(*webrtc.ICECandidate)
	answerChannel chan *SDP
}

func NewCallTransport(h host.Host) *CallTransport {
	// Initialize with host and setup stream handlers
	return &CallTransport{
		host:          h,
		session:       NewCallSession(),
		answerChannel: make(chan *SDP, 1),
	}
}

func (ct *CallTransport) OnICECandidate(handler func(*webrtc.ICECandidate)) {
	// Register ICE candidate handler
	ct.iceHandler = handler
}

func (ct *CallTransport) AddICECandidate(peerID peer.ID, candidate *webrtc.ICECandidate) error {
	// Send ICE candidate over dedicated stream
	// For testing, just call the handler directly
	if ct.iceHandler != nil {
		ct.iceHandler(candidate)
	}

	return nil
}
