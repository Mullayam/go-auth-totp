package ratelimit

import (
	"sync"
	"time"
)

// Limiter defines the interface for checking if an action is allowed.
type Limiter interface {
	Allow(key string) bool
}

// InMemoryLimiter is a simple thread-safe rate limiter.
// In production, use Redis.
type InMemoryLimiter struct {
	mu       sync.Mutex
	counters map[string]*bucket
	rate     time.Duration // Time to refill one token
	capacity int           // Max burst
}

type bucket struct {
	tokens     float64
	lastUpdate time.Time
}

// NewInMemoryLimiter creates a limiter that allows 'capacity' events, refilling at 1 per 'rate'.
// Example: 3 attempts, refill 1 every 30s.
func NewInMemoryLimiter(rate time.Duration, capacity int) *InMemoryLimiter {
	return &InMemoryLimiter{
		counters: make(map[string]*bucket),
		rate:     rate,
		capacity: capacity,
	}
}

// Allow checks if the action is allowed for the given key.
func (l *InMemoryLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	b, exists := l.counters[key]
	if !exists {
		// New user/key locally
		l.counters[key] = &bucket{
			tokens:     float64(l.capacity - 1), // Allow this one
			lastUpdate: time.Now(),
		}
		return true
	}

	now := time.Now()
	elapsed := now.Sub(b.lastUpdate)

	// Refill tokens
	added := float64(elapsed) / float64(l.rate)
	b.tokens += added
	if b.tokens > float64(l.capacity) {
		b.tokens = float64(l.capacity)
	}
	b.lastUpdate = now

	if b.tokens >= 1 {
		b.tokens -= 1
		return true
	}

	return false
}
