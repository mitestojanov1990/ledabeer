package network

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

type ConnectionPoolConfig struct {
	MaxConnections  int
	IdleTimeout     time.Duration
	CleanupInterval time.Duration
}

type ConnectionPool struct {
	host        host.Host
	config      *ConnectionPoolConfig
	connections map[peer.ID]*Connection
	streams     map[protocol.ID][]Stream
	stats       *PoolStats
	mutex       sync.RWMutex
	stopChan    chan struct{}
}

type Connection struct {
	peerID    peer.ID
	streams   map[protocol.ID]Stream
	lastUsed  time.Time
	createdAt time.Time
	mutex     sync.RWMutex
}

type PoolStats struct {
	ActiveConnections int
	ActiveStreams     int
	TotalConnections  int
	IdleConnections   int
}

type Stream interface {
	Write([]byte) (int, error)
	Read([]byte) (int, error)
	Close() error
	ID() string
}

type libp2pStream struct {
	stream network.Stream
}

func (s *libp2pStream) Write(data []byte) (int, error) {
	return s.stream.Write(data)
}

func (s *libp2pStream) Read(data []byte) (int, error) {
	return s.stream.Read(data)
}

func (s *libp2pStream) Close() error {
	return s.stream.Close()
}

func (s *libp2pStream) ID() string {
	return s.stream.ID()
}

func NewConnectionPool(h host.Host, config *ConnectionPoolConfig) *ConnectionPool {
	pool := &ConnectionPool{
		host:        h,
		config:      config,
		connections: make(map[peer.ID]*Connection),
		streams:     make(map[protocol.ID][]Stream),
		stats:       &PoolStats{},
		stopChan:    make(chan struct{}),
	}

	// Start cleanup goroutine
	go pool.cleanupLoop()

	return pool
}

func (p *ConnectionPool) GetStream(ctx context.Context, peerID peer.ID, protocolID protocol.ID) (Stream, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Check if we have an existing connection
	conn, exists := p.connections[peerID]
	if exists && conn.isActive() {
		// For this test, we'll always create a new stream
		// In a real implementation, we might reuse streams
		conn.updateLastUsed()
	}

	// Check connection limit
	if len(p.connections) >= p.config.MaxConnections {
		return nil, fmt.Errorf("connection limit reached: %d", p.config.MaxConnections)
	}

	// Create new connection if needed
	if !exists {
		var err error
		conn, err = p.createConnection(ctx, peerID)
		if err != nil {
			return nil, fmt.Errorf("failed to create connection: %w", err)
		}
		p.connections[peerID] = conn
	}

	// Create new stream
	stream, err := p.createStream(ctx, conn, protocolID)
	if err != nil {
		return nil, fmt.Errorf("failed to create stream: %w", err)
	}

	// Store stream in connection (allow multiple streams per protocol)
	if conn.streams == nil {
		conn.streams = make(map[protocol.ID]Stream)
	}
	// For simplicity, we'll store the latest stream for each protocol
	// In a real implementation, we'd track multiple streams
	conn.streams[protocolID] = stream
	p.streams[protocolID] = append(p.streams[protocolID], stream)

	// Update stats
	p.updateStats()

	return stream, nil
}

func (p *ConnectionPool) createConnection(ctx context.Context, peerID peer.ID) (*Connection, error) {
	// For testing, we'll create a connection without requiring a specific protocol
	// In a real implementation, this would establish a connection and keep it alive

	conn := &Connection{
		peerID:    peerID,
		streams:   make(map[protocol.ID]Stream),
		lastUsed:  time.Now(),
		createdAt: time.Now(),
	}

	return conn, nil
}

func (p *ConnectionPool) createStream(ctx context.Context, conn *Connection, protocolID protocol.ID) (Stream, error) {
	// Create new stream for the specific protocol
	stream, err := p.host.NewStream(ctx, conn.peerID, protocolID)
	if err != nil {
		return nil, err
	}

	return &libp2pStream{stream: stream}, nil
}

func (p *ConnectionPool) Close() {
	close(p.stopChan)

	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Close all connections
	for _, conn := range p.connections {
		conn.close()
	}
	p.connections = make(map[peer.ID]*Connection)
	p.streams = make(map[protocol.ID][]Stream)
}

func (p *ConnectionPool) GetStats() PoolStats {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return PoolStats{
		ActiveConnections: p.stats.ActiveConnections,
		ActiveStreams:     p.stats.ActiveStreams,
		TotalConnections:  p.stats.TotalConnections,
		IdleConnections:   p.stats.IdleConnections,
	}
}

func (p *ConnectionPool) cleanupLoop() {
	ticker := time.NewTicker(p.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.cleanup()
		case <-p.stopChan:
			return
		}
	}
}

func (p *ConnectionPool) cleanup() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	now := time.Now()
	toRemove := make([]peer.ID, 0)

	// Find idle connections
	for peerID, conn := range p.connections {
		if now.Sub(conn.lastUsed) > p.config.IdleTimeout {
			toRemove = append(toRemove, peerID)
		}
	}

	// Remove idle connections
	for _, peerID := range toRemove {
		conn := p.connections[peerID]
		conn.close()
		delete(p.connections, peerID)
	}

	p.updateStats()
}

func (p *ConnectionPool) updateStats() {
	activeConnections := 0
	activeStreams := 0
	idleConnections := 0

	now := time.Now()
	for _, conn := range p.connections {
		if now.Sub(conn.lastUsed) > p.config.IdleTimeout {
			idleConnections++
		} else {
			activeConnections++
		}
	}

	// Count all streams across all protocols
	for _, streams := range p.streams {
		activeStreams += len(streams)
	}

	p.stats.ActiveConnections = activeConnections
	p.stats.ActiveStreams = activeStreams
	p.stats.TotalConnections = len(p.connections)
	p.stats.IdleConnections = idleConnections
}

func (c *Connection) isActive() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// Check if connection is still valid
	// In a real implementation, this would check if the underlying connection is alive
	return len(c.streams) > 0
}

func (c *Connection) updateLastUsed() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.lastUsed = time.Now()
}

func (c *Connection) close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Close all streams
	for _, stream := range c.streams {
		stream.Close()
	}
	c.streams = make(map[protocol.ID]Stream)
}

// Helper function to connect two hosts
func Connect(ctx context.Context, host1, host2 host.Host) error {
	// Get host2's address
	addrs := host2.Addrs()
	if len(addrs) == 0 {
		return fmt.Errorf("host2 has no addresses")
	}

	// Connect host1 to host2
	peerInfo := peer.AddrInfo{
		ID:    host2.ID(),
		Addrs: addrs,
	}

	return host1.Connect(ctx, peerInfo)
}
