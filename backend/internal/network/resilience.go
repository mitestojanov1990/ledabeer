package network

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
)

type ResilienceConfig struct {
	MaxRetries      int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	BackoffFactor   float64
	RetryableErrors []string
	CircuitBreaker  *CircuitBreakerConfig
}

type ResilienceStats struct {
	SuccessfulOperations int
	FailedOperations     int
	TotalRetries         int
	CircuitBreakerTrips  int
}

type ResilientClient struct {
	host           host.Host
	config         *ResilienceConfig
	circuitBreaker *CircuitBreaker
	stats          *ResilienceStats
	mutex          sync.RWMutex
}

func NewResilientClient(h host.Host, config *ResilienceConfig) *ResilientClient {
	client := &ResilientClient{
		host:   h,
		config: config,
		stats:  &ResilienceStats{},
	}

	if config.CircuitBreaker != nil {
		client.circuitBreaker = NewCircuitBreaker(config.CircuitBreaker)
	}

	return client
}

func (c *ResilientClient) ExecuteWithRetry(ctx context.Context, operation func() error) error {
	// Check circuit breaker
	if c.circuitBreaker != nil {
		if !c.circuitBreaker.AllowRequest() {
			return fmt.Errorf("circuit breaker open")
		}
	}

	// Execute operation with retry logic
	var lastErr error
	delay := c.config.InitialDelay

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Execute operation
		err := operation()
		if err == nil {
			// Success - update circuit breaker and stats
			if c.circuitBreaker != nil {
				c.circuitBreaker.RecordSuccess()
			}
			c.mutex.Lock()
			c.stats.SuccessfulOperations++
			c.mutex.Unlock()
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !c.isRetryableError(err) {
			c.mutex.Lock()
			c.stats.FailedOperations++
			c.mutex.Unlock()
			return err
		}

		// Record failure in circuit breaker
		if c.circuitBreaker != nil {
			wasClosed := c.circuitBreaker.IsClosed()
			c.circuitBreaker.RecordFailure()
			// Track circuit breaker trips
			if wasClosed && c.circuitBreaker.IsOpen() {
				c.mutex.Lock()
				c.stats.CircuitBreakerTrips++
				c.mutex.Unlock()
			}
		}

		// If this was the last attempt, don't wait
		if attempt == c.config.MaxRetries {
			break
		}

		// Wait before retry
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}

		// Update delay for next retry (exponential backoff)
		delay = time.Duration(float64(delay) * c.config.BackoffFactor)
		if delay > c.config.MaxDelay {
			delay = c.config.MaxDelay
		}

		// Update retry stats
		c.mutex.Lock()
		c.stats.TotalRetries++
		c.mutex.Unlock()
	}

	// All retries exhausted
	c.mutex.Lock()
	c.stats.FailedOperations++
	c.mutex.Unlock()

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

func (c *ResilientClient) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	for _, retryableErr := range c.config.RetryableErrors {
		if strings.Contains(errStr, retryableErr) {
			return true
		}
	}

	return false
}

func (c *ResilientClient) GetStats() ResilienceStats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return ResilienceStats{
		SuccessfulOperations: c.stats.SuccessfulOperations,
		FailedOperations:     c.stats.FailedOperations,
		TotalRetries:         c.stats.TotalRetries,
		CircuitBreakerTrips:  c.stats.CircuitBreakerTrips,
	}
}

// Circuit Breaker Implementation

// Helper function to calculate exponential backoff delay
func calculateBackoffDelay(attempt int, initialDelay time.Duration, maxDelay time.Duration, factor float64) time.Duration {
	delay := time.Duration(float64(initialDelay) * math.Pow(factor, float64(attempt)))
	if delay > maxDelay {
		delay = maxDelay
	}
	return delay
}

// Helper function to add jitter to prevent thundering herd
func addJitter(delay time.Duration) time.Duration {
	// Add random jitter between 0.5x and 1.5x the delay
	jitter := time.Duration(float64(delay) * 0.5)
	return delay + time.Duration(float64(jitter)*(0.5+0.5*math.Sin(float64(time.Now().UnixNano()))))
}
