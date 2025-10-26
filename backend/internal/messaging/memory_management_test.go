package messaging

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"

	"ledabeer/backend/internal/network"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryManagement_RealCleanup_LongRunningSessions(t *testing.T) {
	// Should clean up memory for long-running sessions
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	// Create message handler with memory management
	msgHandler := NewMessageHandlerWithMemoryManagement(host, &MemoryConfig{
		MaxSessions:     10,
		SessionTimeout:  time.Millisecond * 100,
		CleanupInterval: time.Millisecond * 50,
		MaxMemoryMB:     1, // 1MB limit
	})

	// Create multiple sessions
	sessionIDs := make([]string, 15)
	for i := 0; i < 15; i++ {
		sessionID, err := msgHandler.CreateSession(ctx, fmt.Sprintf("peer%d", i))
		require.NoError(t, err)
		sessionIDs[i] = sessionID
	}

	// Wait for cleanup
	time.Sleep(200 * time.Millisecond)

	// Verify only max sessions remain
	stats := msgHandler.GetMemoryStats()
	assert.LessOrEqual(t, stats.ActiveSessions, 10)
	assert.Greater(t, stats.CleanupCount, 0)
}

func TestMemoryManagement_RealCleanup_SessionTimeout(t *testing.T) {
	// Should clean up expired sessions
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	// Create message handler with short timeout
	msgHandler := NewMessageHandlerWithMemoryManagement(host, &MemoryConfig{
		MaxSessions:     100,
		SessionTimeout:  time.Millisecond * 50, // Very short timeout
		CleanupInterval: time.Millisecond * 25, // Frequent cleanup
		MaxMemoryMB:     10,
	})

	// Create session
	_, err = msgHandler.CreateSession(ctx, "test-peer")
	require.NoError(t, err)

	// Wait for timeout
	time.Sleep(100 * time.Millisecond)

	// Verify session was cleaned up
	stats := msgHandler.GetMemoryStats()
	assert.Equal(t, 0, stats.ActiveSessions)
	assert.Greater(t, stats.CleanupCount, 0)
}

func TestMemoryManagement_RealCleanup_MemoryLimit(t *testing.T) {
	// Should clean up when memory limit is reached
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	// Create message handler with small memory limit
	msgHandler := NewMessageHandlerWithMemoryManagement(host, &MemoryConfig{
		MaxSessions:     100,
		SessionTimeout:  time.Minute,
		CleanupInterval: time.Millisecond * 50,
		MaxMemoryMB:     1, // 1MB limit
	})

	// Create sessions with large data
	for i := 0; i < 5; i++ {
		sessionID, err := msgHandler.CreateSession(ctx, fmt.Sprintf("peer%d", i))
		require.NoError(t, err)

		// Add large data to session
		largeData := make([]byte, 300*1024) // 300KB per session
		err = msgHandler.AddSessionData(ctx, sessionID, largeData)
		require.NoError(t, err)
	}

	// Wait for memory cleanup
	time.Sleep(200 * time.Millisecond)

	// Verify memory usage is within limit
	stats := msgHandler.GetMemoryStats()
	assert.LessOrEqual(t, stats.MemoryUsageMB, 1.0)
	assert.Greater(t, stats.CleanupCount, 0)
}

func TestMemoryManagement_RealCleanup_GarbageCollection(t *testing.T) {
	// Should trigger garbage collection when needed
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	// Create message handler
	msgHandler := NewMessageHandlerWithMemoryManagement(host, &MemoryConfig{
		MaxSessions:     10,
		SessionTimeout:  time.Minute,
		CleanupInterval: time.Millisecond * 100,
		MaxMemoryMB:     1,
	})

	// Force garbage collection
	runtime.GC()

	// Get initial memory stats
	initialStats := msgHandler.GetMemoryStats()

	// Create and destroy many sessions
	for i := 0; i < 20; i++ {
		sessionID, err := msgHandler.CreateSession(ctx, fmt.Sprintf("peer%d", i))
		require.NoError(t, err)

		// Add some data
		data := make([]byte, 50*1024) // 50KB
		err = msgHandler.AddSessionData(ctx, sessionID, data)
		require.NoError(t, err)

		// Close session
		err = msgHandler.CloseSession(ctx, sessionID)
		require.NoError(t, err)
	}

	// Force garbage collection
	runtime.GC()

	// Verify memory was cleaned up
	finalStats := msgHandler.GetMemoryStats()
	assert.LessOrEqual(t, finalStats.MemoryUsageMB, initialStats.MemoryUsageMB+0.1) // Should be similar
}

func TestMemoryManagement_RealCleanup_ConcurrentSessions(t *testing.T) {
	// Should handle concurrent session creation and cleanup
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	// Create message handler
	msgHandler := NewMessageHandlerWithMemoryManagement(host, &MemoryConfig{
		MaxSessions:     5,
		SessionTimeout:  time.Millisecond * 100,
		CleanupInterval: time.Millisecond * 50,
		MaxMemoryMB:     10,
	})

	// Concurrent session creation
	numGoroutines := 10
	results := make(chan string, numGoroutines)
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			sessionID, err := msgHandler.CreateSession(ctx, fmt.Sprintf("peer%d", id))
			if err != nil {
				errors <- err
				return
			}
			results <- sessionID
		}(i)
	}

	// Collect results
	successCount := 0
	for i := 0; i < numGoroutines; i++ {
		select {
		case sessionID := <-results:
			assert.NotEmpty(t, sessionID)
			successCount++
		case err := <-errors:
			t.Logf("Session creation failed: %v", err)
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for session creation")
		}
	}

	// Wait for cleanup
	time.Sleep(200 * time.Millisecond)

	// Verify cleanup occurred
	stats := msgHandler.GetMemoryStats()
	assert.LessOrEqual(t, stats.ActiveSessions, 5)
	assert.Greater(t, stats.CleanupCount, 0)
}

func TestMemoryManagement_RealCleanup_ResourceTracking(t *testing.T) {
	// Should track resource usage accurately
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	// Create message handler
	msgHandler := NewMessageHandlerWithMemoryManagement(host, &MemoryConfig{
		MaxSessions:     10,
		SessionTimeout:  time.Minute,
		CleanupInterval: time.Second,
		MaxMemoryMB:     10,
	})

	// Create session
	sessionID, err := msgHandler.CreateSession(ctx, "test-peer")
	require.NoError(t, err)

	// Add data of known size
	dataSize := 100 * 1024 // 100KB
	data := make([]byte, dataSize)
	err = msgHandler.AddSessionData(ctx, sessionID, data)
	require.NoError(t, err)

	// Verify resource tracking
	stats := msgHandler.GetMemoryStats()
	assert.Equal(t, 1, stats.ActiveSessions)
	assert.GreaterOrEqual(t, stats.MemoryUsageMB, 0.09) // At least 90KB (accounting for precision)
	assert.LessOrEqual(t, stats.MemoryUsageMB, 0.2)     // But not more than 200KB
}
