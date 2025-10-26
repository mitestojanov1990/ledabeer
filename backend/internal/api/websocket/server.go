package websocket

import (
	"encoding/json"
	"net/http"
	"sync"

	"ledabeer/backend/internal/auth"

	"github.com/gorilla/websocket"
)

type Server struct {
	upgrader websocket.Upgrader
	clients  map[string]*Client
	mutex    sync.RWMutex
	auth     *auth.Authenticator
}

type Client struct {
	conn   *websocket.Conn
	peerID string
	send   chan []byte
}

func NewServer(auth *auth.Authenticator) *Server {
	return &Server{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		clients: make(map[string]*Client),
		auth:    auth,
	}
}

func (s *Server) HandleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &Client{
		conn: conn,
		send: make(chan []byte, 256),
	}

	go client.readPump(s)
	go client.writePump()
}

func (s *Server) BroadcastMessage(peerID string, data []byte) {
	s.mutex.RLock()
	client, exists := s.clients[peerID]
	s.mutex.RUnlock()

	if exists {
		// Create a message event
		msg := map[string]interface{}{
			"type":    "message",
			"content": string(data),
		}

		msgBytes, _ := json.Marshal(msg)

		select {
		case client.send <- msgBytes:
		default:
			close(client.send)
			s.mutex.Lock()
			delete(s.clients, peerID)
			s.mutex.Unlock()
		}
	}
}

func (c *Client) readPump(server *Server) {
	defer func() {
		server.mutex.Lock()
		delete(server.clients, c.peerID)
		server.mutex.Unlock()
		c.conn.Close()
	}()

	for {
		var msg map[string]interface{}
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			break
		}

		// Handle authentication
		if msgType, ok := msg["type"].(string); ok && msgType == "auth" {
			if peerID, ok := msg["peer_id"].(string); ok {
				c.peerID = peerID
				server.mutex.Lock()
				server.clients[peerID] = c
				server.mutex.Unlock()

				// Send auth success
				authResp := map[string]string{
					"type": "auth_success",
				}
				c.conn.WriteJSON(authResp)
			}
		}
	}
}

func (c *Client) writePump() {
	defer c.conn.Close()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Send message as JSON
			var msg map[string]interface{}
			json.Unmarshal(message, &msg)
			c.conn.WriteJSON(msg)
		}
	}
}
