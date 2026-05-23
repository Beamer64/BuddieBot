package helper

import (
	"sync"
	"time"
)

// RateLimiter enforces a per-key minimum interval between actions.
// Safe for concurrent use.
type RateLimiter struct {
	mu       sync.Mutex
	lastSeen map[string]time.Time
	cooldown time.Duration
}

func NewRateLimiter(cooldown time.Duration) *RateLimiter {
	return &RateLimiter{
		lastSeen: make(map[string]time.Time),
		cooldown: cooldown,
	}
}

// Allow reports whether an action for key may proceed right now. When
// rate-limited, the second return value is how long the caller should
// wait before trying again.
func (r *RateLimiter) Allow(key string) (bool, time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if last, ok := r.lastSeen[key]; ok {
		if elapsed := now.Sub(last); elapsed < r.cooldown {
			return false, r.cooldown - elapsed
		}
	}
	r.lastSeen[key] = now
	return true, 0
}
