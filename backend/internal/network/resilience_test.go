package network

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNetworkResilience_RealRetry_ExponentialBackoff(t *testing.T) {
	// Should retry failed operations with exponential backoff
	ctx := context.Background()

	host, err := NewHost(ctx, &Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	// Create resilient client
	client := NewResilientClient(host, &ResilienceConfig{
		MaxRetries:      3,
		InitialDelay:    time.Millisecond * 10,
		MaxDelay:        time.Millisecond * 100,
		BackoffFactor:   2.0,
		RetryableErrors: []string{"connection failed", "timeout"},
	})

	// Track retry attempts
	attempts := 0
	operation := func() error {
		attempts++
		if attempts < 3 {
			return fmt.Errorf("connection failed")
		}
		return nil
	}

	// Execute with retry
	start := time.Now()
	err = client.ExecuteWithRetry(ctx, operation)
	require.NoError(t, err)

	// Verify retry attempts
	assert.Equal(t, 3, attempts)

	// Verify exponential backoff timing
	elapsed := time.Since(start)
	assert.Greater(t, elapsed, time.Millisecond*30) // At least 10ms + 20ms + 40ms
	assert.Less(t, elapsed, time.Millisecond*200)   // But not too long
}

func TestNetworkResilience_RealRetry_MaxRetries(t *testing.T) {
	// Should stop retrying after max retries
	ctx := context.Background()

	host, err := NewHost(ctx, &Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	client := NewResilientClient(host, &ResilienceConfig{
		MaxRetries:      2,
		InitialDelay:    time.Millisecond * 10,
		MaxDelay:        time.Millisecond * 50,
		BackoffFactor:   2.0,
		RetryableErrors: []string{"connection failed"},
	})

	attempts := 0
	operation := func() error {
		attempts++
		return fmt.Errorf("connection failed")
	}

	// Should fail after max retries
	err = client.ExecuteWithRetry(ctx, operation)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "max retries exceeded")
	assert.Equal(t, 3, attempts) // Initial + 2 retries
}

func TestNetworkResilience_RealRetry_NonRetryableError(t *testing.T) {
	// Should not retry non-retryable errors
	ctx := context.Background()

	host, err := NewHost(ctx, &Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	client := NewResilientClient(host, &ResilienceConfig{
		MaxRetries:      3,
		InitialDelay:    time.Millisecond * 10,
		MaxDelay:        time.Millisecond * 50,
		BackoffFactor:   2.0,
		RetryableErrors: []string{"connection failed"},
	})

	attempts := 0
	operation := func() error {
		attempts++
		return fmt.Errorf("invalid request") // Non-retryable error
	}

	// Should fail immediately without retry
	err = client.ExecuteWithRetry(ctx, operation)
	assert.Error(t, err)
	assert.Equal(t, 1, attempts) // Only initial attempt
}

func TestNetworkResilience_RealRetry_ContextCancellation(t *testing.T) {
	// Should respect context cancellation
	ctx, cancel := context.WithCancel(context.Background())

	host, err := NewHost(ctx, &Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	client := NewResilientClient(host, &ResilienceConfig{
		MaxRetries:      5,
		InitialDelay:    time.Millisecond * 50,
		MaxDelay:        time.Millisecond * 200,
		BackoffFactor:   2.0,
		RetryableErrors: []string{"connection failed"},
	})

	attempts := 0
	operation := func() error {
		attempts++
		return fmt.Errorf("connection failed")
	}

	// Cancel context after first attempt
	go func() {
		time.Sleep(time.Millisecond * 25)
		cancel()
	}()

	// Should fail due to context cancellation
	err = client.ExecuteWithRetry(ctx, operation)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
	assert.Equal(t, 1, attempts) // Only initial attempt before cancellation
}

func TestNetworkResilience_RealRetry_CircuitBreaker(t *testing.T) {
	// Should open circuit breaker after consecutive failures
	ctx := context.Background()

	host, err := NewHost(ctx, &Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	client := NewResilientClient(host, &ResilienceConfig{
		MaxRetries:      2,
		InitialDelay:    time.Millisecond * 10,
		MaxDelay:        time.Millisecond * 50,
		BackoffFactor:   2.0,
		RetryableErrors: []string{"connection failed"},
		CircuitBreaker: &CircuitBreakerConfig{
			FailureThreshold: 3,
			RecoveryTimeout:  time.Millisecond * 100,
		},
	})

	// Fail multiple operations to trigger circuit breaker
	for i := 0; i < 4; i++ {
		operation := func() error {
			return fmt.Errorf("connection failed")
		}
		err = client.ExecuteWithRetry(ctx, operation)
		assert.Error(t, err)
	}

	// Circuit should be open now
	operation := func() error {
		return fmt.Errorf("connection failed")
	}
	err = client.ExecuteWithRetry(ctx, operation)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circuit breaker open")

	// Wait for circuit to recover
	time.Sleep(150 * time.Millisecond)

	// Should work again after recovery
	operation = func() error {
		return nil
	}
	err = client.ExecuteWithRetry(ctx, operation)
	assert.NoError(t, err)
}

func TestNetworkResilience_RealRetry_ConcurrentOperations(t *testing.T) {
	// Should handle concurrent retry operations safely
	ctx := context.Background()

	host, err := NewHost(ctx, &Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	client := NewResilientClient(host, &ResilienceConfig{
		MaxRetries:      2,
		InitialDelay:    time.Millisecond * 10,
		MaxDelay:        time.Millisecond * 50,
		BackoffFactor:   2.0,
		RetryableErrors: []string{"connection failed"},
	})

	// Concurrent operations
	numGoroutines := 5
	results := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			operation := func() error {
				if id%2 == 0 {
					return nil // Success
				}
				return fmt.Errorf("connection failed")
			}
			results <- client.ExecuteWithRetry(ctx, operation)
		}(i)
	}

	// Collect results
	successCount := 0
	errorCount := 0
	for i := 0; i < numGoroutines; i++ {
		select {
		case err := <-results:
			if err != nil {
				errorCount++
			} else {
				successCount++
			}
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for results")
		}
	}

	// Should have some successes and some failures
	assert.Greater(t, successCount, 0)
	assert.Greater(t, errorCount, 0)
}

func TestNetworkResilience_RealRetry_StatsTracking(t *testing.T) {
	// Should track retry statistics
	ctx := context.Background()

	host, err := NewHost(ctx, &Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	client := NewResilientClient(host, &ResilienceConfig{
		MaxRetries:      2,
		InitialDelay:    time.Millisecond * 10,
		MaxDelay:        time.Millisecond * 50,
		BackoffFactor:   2.0,
		RetryableErrors: []string{"connection failed"},
		CircuitBreaker: &CircuitBreakerConfig{
			FailureThreshold: 1,
			RecoveryTimeout:  time.Millisecond * 100,
		},
	})

	// Execute operations with different outcomes
	operation1 := func() error {
		return nil // Success
	}
	err = client.ExecuteWithRetry(ctx, operation1)
	require.NoError(t, err)

	operation2 := func() error {
		return fmt.Errorf("connection failed")
	}
	err = client.ExecuteWithRetry(ctx, operation2)
	assert.Error(t, err)

	// Check stats
	stats := client.GetStats()
	assert.Equal(t, 1, stats.SuccessfulOperations)
	assert.Equal(t, 1, stats.FailedOperations)
	assert.Equal(t, 2, stats.TotalRetries)
	assert.Equal(t, 1, stats.CircuitBreakerTrips)
}
