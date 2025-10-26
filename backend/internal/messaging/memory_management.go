package messaging

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
)

type MemoryConfig struct {
	MaxSessions     int
	SessionTimeout  time.Duration
	CleanupInterval time.Duration
	MaxMemoryMB     int
}

type MemoryStats struct {
	ActiveSessions int
	MemoryUsageMB  float64
	CleanupCount   int
	MaxSessions    int
	MaxMemoryMB    int
	GCTriggered    int
}

type SessionData struct {
	ID         string
	PeerID     string
	Data       []byte
	CreatedAt  time.Time
	LastAccess time.Time
}

type MessageHandlerWithMemoryManagement struct {
	host     host.Host
	config   *MemoryConfig
	sessions map[string]*SessionData
	stats    *MemoryStats
	mutex    sync.RWMutex
	stopChan chan struct{}
}

func NewMessageHandlerWithMemoryManagement(h host.Host, config *MemoryConfig) *MessageHandlerWithMemoryManagement {
	handler := &MessageHandlerWithMemoryManagement{
		host:     h,
		config:   config,
		sessions: make(map[string]*SessionData),
		stats:    &MemoryStats{MaxSessions: config.MaxSessions, MaxMemoryMB: config.MaxMemoryMB},
		stopChan: make(chan struct{}),
	}

	// Start cleanup goroutine
	go handler.cleanupLoop()

	return handler
}

func (m *MessageHandlerWithMemoryManagement) CreateSession(ctx context.Context, peerID string) (string, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check session limit
	if len(m.sessions) >= m.config.MaxSessions {
		// Clean up oldest sessions
		m.cleanupOldestSessions(1)
	}

	// Create new session
	sessionID := generateSessionID()
	session := &SessionData{
		ID:         sessionID,
		PeerID:     peerID,
		Data:       make([]byte, 0),
		CreatedAt:  time.Now(),
		LastAccess: time.Now(),
	}

	m.sessions[sessionID] = session
	m.stats.ActiveSessions = len(m.sessions)

	return sessionID, nil
}

func (m *MessageHandlerWithMemoryManagement) AddSessionData(ctx context.Context, sessionID string, data []byte) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// Check memory limit
	currentMemoryMB := m.calculateMemoryUsage()
	if currentMemoryMB+float64(len(data))/(1024*1024) > float64(m.config.MaxMemoryMB) {
		// Trigger cleanup
		m.cleanupMemory()
	}

	// Add data to session
	session.Data = append(session.Data, data...)
	session.LastAccess = time.Now()

	// Update memory stats
	m.stats.MemoryUsageMB = m.calculateMemoryUsage()

	return nil
}

func (m *MessageHandlerWithMemoryManagement) CloseSession(ctx context.Context, sessionID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// Clear session data
	session.Data = nil
	delete(m.sessions, sessionID)
	m.stats.ActiveSessions = len(m.sessions)
	m.stats.MemoryUsageMB = m.calculateMemoryUsage()

	return nil
}

func (m *MessageHandlerWithMemoryManagement) GetMemoryStats() MemoryStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return MemoryStats{
		ActiveSessions: m.stats.ActiveSessions,
		MemoryUsageMB:  m.stats.MemoryUsageMB,
		CleanupCount:   m.stats.CleanupCount,
		MaxSessions:    m.stats.MaxSessions,
		MaxMemoryMB:    m.stats.MaxMemoryMB,
		GCTriggered:    m.stats.GCTriggered,
	}
}

func (m *MessageHandlerWithMemoryManagement) cleanupLoop() {
	ticker := time.NewTicker(m.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cleanup()
		case <-m.stopChan:
			return
		}
	}
}

func (m *MessageHandlerWithMemoryManagement) cleanup() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()
	toRemove := make([]string, 0)

	// Find expired sessions
	for sessionID, session := range m.sessions {
		if now.Sub(session.LastAccess) > m.config.SessionTimeout {
			toRemove = append(toRemove, sessionID)
		}
	}

	// Remove expired sessions
	for _, sessionID := range toRemove {
		delete(m.sessions, sessionID)
		m.stats.CleanupCount++
	}

	// Update stats
	m.stats.ActiveSessions = len(m.sessions)
	m.stats.MemoryUsageMB = m.calculateMemoryUsage()

	// Trigger GC if memory usage is high
	if m.stats.MemoryUsageMB > float64(m.config.MaxMemoryMB)*0.8 {
		runtime.GC()
		m.stats.GCTriggered++
	}
}

func (m *MessageHandlerWithMemoryManagement) cleanupOldestSessions(count int) {
	// Find oldest sessions
	sessions := make([]*SessionData, 0, len(m.sessions))
	for _, session := range m.sessions {
		sessions = append(sessions, session)
	}

	// Sort by last access time (oldest first)
	for i := 0; i < len(sessions)-1; i++ {
		for j := i + 1; j < len(sessions); j++ {
			if sessions[i].LastAccess.After(sessions[j].LastAccess) {
				sessions[i], sessions[j] = sessions[j], sessions[i]
			}
		}
	}

	// Remove oldest sessions
	for i := 0; i < count && i < len(sessions); i++ {
		delete(m.sessions, sessions[i].ID)
		m.stats.CleanupCount++
	}
}

func (m *MessageHandlerWithMemoryManagement) cleanupMemory() {
	// Remove sessions until memory usage is acceptable
	for m.calculateMemoryUsage() > float64(m.config.MaxMemoryMB)*0.8 && len(m.sessions) > 0 {
		m.cleanupOldestSessions(1)
	}
}

func (m *MessageHandlerWithMemoryManagement) calculateMemoryUsage() float64 {
	totalBytes := 0
	for _, session := range m.sessions {
		totalBytes += len(session.Data)
	}
	return float64(totalBytes) / (1024 * 1024) // Convert to MB
}

func (m *MessageHandlerWithMemoryManagement) Close() {
	close(m.stopChan)

	// Clear all sessions
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for sessionID := range m.sessions {
		delete(m.sessions, sessionID)
	}
	m.stats.ActiveSessions = 0
	m.stats.MemoryUsageMB = 0
}

// Helper function to generate session ID
func generateSessionID() string {
	// Simple implementation for testing
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}
