package network

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCircuitBreaker_RealBreaker_OpenOnFailure(t *testing.T) {
	// Should open circuit breaker after consecutive failures
	ctx := context.Background()

	host, err := NewHost(ctx, &Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	// Create circuit breaker
	breaker := NewCircuitBreaker(&CircuitBreakerConfig{
		FailureThreshold: 3,
		RecoveryTimeout:  time.Millisecond * 100,
		MaxRequests:      5,
	})

	// Simulate failures to trigger circuit breaker
	for i := 0; i < 3; i++ {
		breaker.RecordFailure()
	}

	// Circuit should be open
	assert.True(t, breaker.IsOpen())
	assert.False(t, breaker.AllowRequest())
}

func TestCircuitBreaker_RealBreaker_Recovery(t *testing.T) {
	// Should recover from open state after timeout
	ctx := context.Background()

	host, err := NewHost(ctx, &Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	breaker := NewCircuitBreaker(&CircuitBreakerConfig{
		FailureThreshold: 2,
		RecoveryTimeout:  time.Millisecond * 50,
		MaxRequests:      5,
	})

	// Trigger circuit breaker
	breaker.RecordFailure()
	breaker.RecordFailure()
	assert.True(t, breaker.IsOpen())

	// Wait for recovery timeout
	time.Sleep(100 * time.Millisecond)

	// Circuit should be in half-open state
	assert.True(t, breaker.IsHalfOpen())
	assert.True(t, breaker.AllowRequest())
}

func TestCircuitBreaker_RealBreaker_SuccessfulRecovery(t *testing.T) {
	// Should close circuit after successful request in half-open state
	ctx := context.Background()

	host, err := NewHost(ctx, &Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	breaker := NewCircuitBreaker(&CircuitBreakerConfig{
		FailureThreshold: 2,
		RecoveryTimeout:  time.Millisecond * 50,
		MaxRequests:      5,
	})

	// Trigger circuit breaker
	breaker.RecordFailure()
	breaker.RecordFailure()
	assert.True(t, breaker.IsOpen())

	// Wait for recovery
	time.Sleep(100 * time.Millisecond)
	assert.True(t, breaker.IsHalfOpen())

	// Record success
	breaker.RecordSuccess()

	// Circuit should be closed
	assert.True(t, breaker.IsClosed())
	assert.True(t, breaker.AllowRequest())
}

func TestCircuitBreaker_RealBreaker_FailureInHalfOpen(t *testing.T) {
	// Should open circuit again if failure occurs in half-open state
	ctx := context.Background()

	host, err := NewHost(ctx, &Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	breaker := NewCircuitBreaker(&CircuitBreakerConfig{
		FailureThreshold: 2,
		RecoveryTimeout:  time.Millisecond * 50,
		MaxRequests:      5,
	})

	// Trigger circuit breaker
	breaker.RecordFailure()
	breaker.RecordFailure()
	assert.True(t, breaker.IsOpen())

	// Wait for recovery
	time.Sleep(100 * time.Millisecond)
	assert.True(t, breaker.IsHalfOpen())

	// Record failure in half-open state
	breaker.RecordFailure()

	// Circuit should be open again
	assert.True(t, breaker.IsOpen())
	assert.False(t, breaker.AllowRequest())
}

func TestCircuitBreaker_RealBreaker_RequestLimiting(t *testing.T) {
	// Should limit requests in half-open state
	ctx := context.Background()

	host, err := NewHost(ctx, &Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	breaker := NewCircuitBreaker(&CircuitBreakerConfig{
		FailureThreshold: 2,
		RecoveryTimeout:  time.Millisecond * 50,
		MaxRequests:      2, // Limit to 2 requests
	})

	// Trigger circuit breaker
	breaker.RecordFailure()
	breaker.RecordFailure()
	assert.True(t, breaker.IsOpen())

	// Wait for recovery
	time.Sleep(100 * time.Millisecond)
	assert.True(t, breaker.IsHalfOpen())

	// First two requests should be allowed
	assert.True(t, breaker.AllowRequest())
	assert.True(t, breaker.AllowRequest())

	// Third request should be blocked
	assert.False(t, breaker.AllowRequest())
}

func TestCircuitBreaker_RealBreaker_ConcurrentAccess(t *testing.T) {
	// Should handle concurrent access safely
	ctx := context.Background()

	host, err := NewHost(ctx, &Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	breaker := NewCircuitBreaker(&CircuitBreakerConfig{
		FailureThreshold: 3,
		RecoveryTimeout:  time.Millisecond * 100,
		MaxRequests:      10,
	})

	// Concurrent operations
	numGoroutines := 10
	results := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			results <- breaker.AllowRequest()
		}()
	}

	// Collect results
	allowedCount := 0
	for i := 0; i < numGoroutines; i++ {
		select {
		case allowed := <-results:
			if allowed {
				allowedCount++
			}
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for results")
		}
	}

	// Should have some requests allowed (circuit starts closed)
	assert.Greater(t, allowedCount, 0)
}

func TestCircuitBreaker_RealBreaker_StatsTracking(t *testing.T) {
	// Should track circuit breaker statistics
	ctx := context.Background()

	host, err := NewHost(ctx, &Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	breaker := NewCircuitBreaker(&CircuitBreakerConfig{
		FailureThreshold: 2,
		RecoveryTimeout:  time.Millisecond * 100,
		MaxRequests:      5,
	})

	// Record some events
	breaker.RecordFailure()
	breaker.RecordFailure()
	breaker.RecordSuccess()

	// Check stats
	stats := breaker.GetStats()
	assert.Equal(t, 2, stats.Failures)
	assert.Equal(t, 1, stats.Successes)
	assert.Equal(t, 1, stats.Trips) // Circuit opened once
	assert.True(t, stats.IsClosed)
}

func TestCircuitBreaker_RealBreaker_StateTransitions(t *testing.T) {
	// Should properly transition between states
	ctx := context.Background()

	host, err := NewHost(ctx, &Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	breaker := NewCircuitBreaker(&CircuitBreakerConfig{
		FailureThreshold: 2,
		RecoveryTimeout:  time.Millisecond * 50,
		MaxRequests:      5,
	})

	// Start closed
	assert.True(t, breaker.IsClosed())
	assert.True(t, breaker.AllowRequest())

	// Record failures to open
	breaker.RecordFailure()
	assert.True(t, breaker.IsClosed())
	breaker.RecordFailure()
	assert.True(t, breaker.IsOpen())
	assert.False(t, breaker.AllowRequest())

	// Wait for recovery
	time.Sleep(100 * time.Millisecond)
	assert.True(t, breaker.IsHalfOpen())
	assert.True(t, breaker.AllowRequest())

	// Record success to close
	breaker.RecordSuccess()
	assert.True(t, breaker.IsClosed())
	assert.True(t, breaker.AllowRequest())
}
