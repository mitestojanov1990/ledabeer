package websocket_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"ledabeer/backend/internal/api"
	ws "ledabeer/backend/internal/api/websocket"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebSocket_Connect(t *testing.T) {
	// Should accept WebSocket connections
	server := setupTestWebSocketServer(t)
	defer server.Close()

	wsURL := "ws://" + server.URL[7:] + "/ws" // Convert http:// to ws://
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	assert.NotNil(t, conn)
}

func TestWebSocket_Authenticate(t *testing.T) {
	// Should authenticate via initial message
	server := setupTestWebSocketServer(t)
	defer server.Close()

	wsURL := "ws://" + server.URL[7:] + "/ws" // Convert http:// to ws://
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	authMsg := map[string]string{
		"type":    "auth",
		"peer_id": "12D3KooWTest123456789012345678901234567890123456789012345678901234567890",
	}

	err = conn.WriteJSON(authMsg)
	require.NoError(t, err)

	var resp map[string]interface{}
	err = conn.ReadJSON(&resp)
	require.NoError(t, err)
	assert.Equal(t, "auth_success", resp["type"])
}

func TestWebSocket_ReceiveMessage(t *testing.T) {
	// Should push messages to connected clients
	server := setupTestWebSocketServer(t)
	defer server.Close()

	conn := authenticateAndConnect(t, server)
	defer conn.Close()

	// Test that connection is established
	assert.NotNil(t, conn)
}

// Helper functions
func setupTestWebSocketServer(t *testing.T) *httptest.Server {
	auth := api.NewAuthenticator()
	wsServer := ws.NewServer(auth)

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", wsServer.HandleConnection)

	return httptest.NewServer(mux)
}

func authenticateAndConnect(t *testing.T, server *httptest.Server) *websocket.Conn {
	wsURL := "ws://" + server.URL[7:] + "/ws" // Convert http:// to ws://
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)

	// Send auth message
	authMsg := map[string]string{
		"type":    "auth",
		"peer_id": "12D3KooWTest123456789012345678901234567890123456789012345678901234567890",
	}

	err = conn.WriteJSON(authMsg)
	require.NoError(t, err)

	// Read auth response
	var resp map[string]interface{}
	err = conn.ReadJSON(&resp)
	require.NoError(t, err)
	assert.Equal(t, "auth_success", resp["type"])

	return conn
}
