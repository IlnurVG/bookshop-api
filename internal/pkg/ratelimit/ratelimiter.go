package ratelimit

import (
	"errors"
	"sync"
	"time"
)

// Errors
var (
	// ErrRateLimitExceeded is returned when the rate limit is exceeded
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
)

// RateLimiter limits the number of requests in a specified time interval
type RateLimiter struct {
	limit    int           // Maximum number of requests per interval
	interval time.Duration // Time interval for rate limiting (e.g., 1 second)
	mu       sync.Mutex    // Mutex to protect the counter
	count    int           // Current request count
	timer    *time.Timer   // Timer for resetting the counter
	stopChan chan struct{} // Channel for stopping the goroutine
}

// NewRateLimiter creates a new rate limiter with the specified requests per second limit
func NewRateLimiter(limit int) *RateLimiter {
	return NewRateLimiterWithInterval(limit, time.Second)
}

// NewRateLimiterWithInterval creates a new rate limiter with the specified limit and interval
func NewRateLimiterWithInterval(limit int, interval time.Duration) *RateLimiter {
	rl := &RateLimiter{
		limit:    limit,
		interval: interval,
		count:    0,
		timer:    time.NewTimer(interval),
		stopChan: make(chan struct{}),
	}

	// Start a goroutine to reset the counter
	go func() {
		for {
			select {
			case <-rl.timer.C:
				rl.mu.Lock()
				rl.count = 0 // Reset the counter
				rl.mu.Unlock()
				rl.timer.Reset(rl.interval)
			case <-rl.stopChan:
				if !rl.timer.Stop() {
					select {
					case <-rl.timer.C: // Drain the channel if timer already fired
					default:
					}
				}
				return
			}
		}
	}()

	return rl
}

// Process executes the function if the rate limit is not exceeded
func (rl *RateLimiter) Process(f func()) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Check if the limit is exceeded
	if rl.count >= rl.limit {
		return ErrRateLimitExceeded
	}

	// Increment the counter
	rl.count++

	// Execute the function
	f()
	return nil
}

// Stop stops the rate limiter and releases resources
func (rl *RateLimiter) Stop() {
	close(rl.stopChan)
}

// MultiRateLimiter manages multiple rate limiters by different keys
type MultiRateLimiter struct {
	limiters map[string]*RateLimiter
	limit    int
	interval time.Duration
	mu       sync.Mutex
}

// NewMultiRateLimiter creates a new composite rate limiter for
// managing rate limits by different keys (e.g., IP addresses)
func NewMultiRateLimiter(limit int, interval time.Duration) *MultiRateLimiter {
	return &MultiRateLimiter{
		limiters: make(map[string]*RateLimiter),
		limit:    limit,
		interval: interval,
		mu:       sync.Mutex{},
	}
}

// Process executes the function for the specified key if the rate limit is not exceeded
func (mrl *MultiRateLimiter) Process(key string, f func()) error {
	mrl.mu.Lock()
	limiter, ok := mrl.limiters[key]
	if !ok {
		limiter = NewRateLimiterWithInterval(mrl.limit, mrl.interval)
		mrl.limiters[key] = limiter
	}
	mrl.mu.Unlock()

	return limiter.Process(f)
}

// Stop stops all rate limiters and releases resources
func (mrl *MultiRateLimiter) Stop() {
	mrl.mu.Lock()
	defer mrl.mu.Unlock()

	for _, limiter := range mrl.limiters {
		limiter.Stop()
	}
	mrl.limiters = make(map[string]*RateLimiter)
}

// Cleanup removes unused rate limiters
// This function can be called periodically to free memory
func (mrl *MultiRateLimiter) Cleanup(maxAge time.Duration) {
	// Implement a mechanism to clean up old unused limiters
}
