package network

import (
	"sync"
	"time"
)

type CircuitBreakerConfig struct {
	FailureThreshold int
	RecoveryTimeout  time.Duration
	MaxRequests      int
}

type CircuitBreakerStats struct {
	Failures   int
	Successes  int
	Trips      int
	IsClosed   bool
	IsOpen     bool
	IsHalfOpen bool
}

type CircuitBreaker struct {
	config       *CircuitBreakerConfig
	state        CircuitBreakerState
	failures     int
	successes    int
	lastFailure  time.Time
	requestCount int
	trips        int
	mutex        sync.RWMutex
}

type CircuitBreakerState int

const (
	CircuitBreakerClosed CircuitBreakerState = iota
	CircuitBreakerOpen
	CircuitBreakerHalfOpen
)

func NewCircuitBreaker(config *CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		config: config,
		state:  CircuitBreakerClosed,
	}
}

func (cb *CircuitBreaker) AllowRequest() bool {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	// Check if we should transition from open to half-open
	if cb.state == CircuitBreakerOpen && time.Since(cb.lastFailure) > cb.config.RecoveryTimeout {
		cb.state = CircuitBreakerHalfOpen
		cb.requestCount = 0
	}

	switch cb.state {
	case CircuitBreakerClosed:
		return true
	case CircuitBreakerOpen:
		return false
	case CircuitBreakerHalfOpen:
		// Check if we've exceeded max requests
		if cb.requestCount >= cb.config.MaxRequests {
			return false
		}
		cb.requestCount++
		return true
	default:
		return false
	}
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.successes++
	cb.requestCount = 0
	cb.state = CircuitBreakerClosed
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failures++
	cb.lastFailure = time.Now()

	if cb.state == CircuitBreakerHalfOpen {
		cb.state = CircuitBreakerOpen
		cb.trips++
	} else if cb.failures >= cb.config.FailureThreshold && cb.state == CircuitBreakerClosed {
		cb.state = CircuitBreakerOpen
		cb.trips++
	}
}

func (cb *CircuitBreaker) IsClosed() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state == CircuitBreakerClosed
}

func (cb *CircuitBreaker) IsOpen() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state == CircuitBreakerOpen
}

func (cb *CircuitBreaker) IsHalfOpen() bool {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	// Check if we should transition from open to half-open
	if cb.state == CircuitBreakerOpen && time.Since(cb.lastFailure) > cb.config.RecoveryTimeout {
		cb.state = CircuitBreakerHalfOpen
		cb.requestCount = 0
	}

	return cb.state == CircuitBreakerHalfOpen
}

func (cb *CircuitBreaker) GetStats() CircuitBreakerStats {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	return CircuitBreakerStats{
		Failures:   cb.failures,
		Successes:  cb.successes,
		Trips:      cb.trips,
		IsClosed:   cb.state == CircuitBreakerClosed,
		IsOpen:     cb.state == CircuitBreakerOpen,
		IsHalfOpen: cb.state == CircuitBreakerHalfOpen,
	}
}

func (cb *CircuitBreaker) Reset() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failures = 0
	cb.successes = 0
	cb.requestCount = 0
	cb.trips = 0
	cb.state = CircuitBreakerClosed
}

func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}
