package network

import (
	"bufio"
	"encoding/binary"
	"io"

	"github.com/libp2p/go-libp2p/core/network"
)

// StreamHandler handles incoming streams
type StreamHandler struct {
	handler func([]byte) error
}

// NewStreamHandler creates a new stream handler
func NewStreamHandler(handler func([]byte) error) *StreamHandler {
	return &StreamHandler{
		handler: handler,
	}
}

// Handle processes an incoming stream
func (h *StreamHandler) Handle(stream network.Stream) {
	defer stream.Close()

	// Read message
	data, err := ReadMessage(stream)
	if err != nil {
		return
	}

	// Process message
	h.handler(data)
}

// WriteMessage writes a length-prefixed message to a stream
func WriteMessage(stream network.Stream, data []byte) error {
	// Write length prefix (4 bytes)
	length := uint32(len(data))
	if err := binary.Write(stream, binary.BigEndian, length); err != nil {
		return err
	}

	// Write data
	_, err := stream.Write(data)
	return err
}

// ReadMessage reads a length-prefixed message from a stream
func ReadMessage(stream network.Stream) ([]byte, error) {
	reader := bufio.NewReader(stream)

	// Read length prefix (4 bytes)
	var length uint32
	if err := binary.Read(reader, binary.BigEndian, &length); err != nil {
		return nil, err
	}

	// Read data
	data := make([]byte, length)
	_, err := io.ReadFull(reader, data)
	if err != nil {
		return nil, err
	}

	return data, nil
}
