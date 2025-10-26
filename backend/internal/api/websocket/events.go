package websocket

type Event interface {
	Type() string
}

type MessageEvent struct {
	EventType string `json:"type"`
	From      string `json:"from"`
	Content   []byte `json:"content"`
	Timestamp int64  `json:"timestamp"`
}

func (e *MessageEvent) Type() string {
	return e.EventType
}

type CallEvent struct {
	EventType string `json:"type"`
	CallID    string `json:"call_id"`
	From      string `json:"from"`
	SDP       string `json:"sdp,omitempty"`
}

func (e *CallEvent) Type() string {
	return e.EventType
}

type PresenceEvent struct {
	EventType string `json:"type"`
	PeerID    string `json:"peer_id"`
	Status    string `json:"status"`
}

func (e *PresenceEvent) Type() string {
	return e.EventType
}

type MediaEvent struct {
	EventType string `json:"type"`
	From      string `json:"from"`
	MediaCID  string `json:"media_cid"`
	MimeType  string `json:"mime_type"`
}

func (e *MediaEvent) Type() string {
	return e.EventType
}
