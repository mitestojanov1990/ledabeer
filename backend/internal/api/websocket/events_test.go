package websocket_test

import (
	"testing"
	"time"

	"ledabeer/backend/internal/api/websocket"

	"github.com/stretchr/testify/assert"
)

func TestWebSocket_MessageEvent(t *testing.T) {
	// Should send message events
	event := &websocket.MessageEvent{
		EventType: "message",
		From:      "peer1",
		Content:   []byte("test"),
		Timestamp: time.Now().Unix(),
	}

	assert.Equal(t, "message", event.Type())
	assert.Equal(t, "peer1", event.From)
	assert.Equal(t, []byte("test"), event.Content)
}

func TestWebSocket_CallEvent(t *testing.T) {
	// Should send call events
	event := &websocket.CallEvent{
		EventType: "call_incoming",
		CallID:    "call123",
		From:      "peer1",
	}

	assert.Equal(t, "call_incoming", event.Type())
	assert.Equal(t, "call123", event.CallID)
	assert.Equal(t, "peer1", event.From)
}

func TestWebSocket_PresenceEvent(t *testing.T) {
	// Should send presence updates
	event := &websocket.PresenceEvent{
		EventType: "peer_online",
		PeerID:    "peer2",
		Status:    "online",
	}

	assert.Equal(t, "peer_online", event.Type())
	assert.Equal(t, "peer2", event.PeerID)
	assert.Equal(t, "online", event.Status)
}

func TestWebSocket_MediaEvent(t *testing.T) {
	// Should send media events
	event := &websocket.MediaEvent{
		EventType: "media_received",
		From:      "peer1",
		MediaCID:  "QmTest...",
		MimeType:  "image/jpeg",
	}

	assert.Equal(t, "media_received", event.Type())
	assert.Equal(t, "peer1", event.From)
	assert.Equal(t, "QmTest...", event.MediaCID)
	assert.Equal(t, "image/jpeg", event.MimeType)
}
