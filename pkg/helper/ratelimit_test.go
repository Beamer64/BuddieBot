package helper

import (
	"sync"
	"testing"
	"time"
)

func TestRateLimiterAllowsFirst(t *testing.T) {
	rl := NewRateLimiter(50 * time.Millisecond)
	if ok, _ := rl.Allow("u1"); !ok {
		t.Fatal("first call should be allowed")
	}
}

func TestRateLimiterBlocksWithinCooldown(t *testing.T) {
	rl := NewRateLimiter(50 * time.Millisecond)
	_, _ = rl.Allow("u1")
	ok, retry := rl.Allow("u1")
	if ok {
		t.Fatal("second call within cooldown should be blocked")
	}
	if retry <= 0 || retry > 50*time.Millisecond {
		t.Fatalf("expected retry in (0, 50ms], got %v", retry)
	}
}

func TestRateLimiterAllowsAfterCooldown(t *testing.T) {
	rl := NewRateLimiter(20 * time.Millisecond)
	_, _ = rl.Allow("u1")
	time.Sleep(25 * time.Millisecond)
	if ok, _ := rl.Allow("u1"); !ok {
		t.Fatal("call after cooldown should be allowed")
	}
}

func TestRateLimiterPerKey(t *testing.T) {
	rl := NewRateLimiter(50 * time.Millisecond)
	_, _ = rl.Allow("u1")
	if ok, _ := rl.Allow("u2"); !ok {
		t.Fatal("different key should not be affected by u1's cooldown")
	}
}

func TestRateLimiterConcurrent(t *testing.T) {
	// Hammer Allow from many goroutines for the same key. Exactly one
	// of them should win the first slot; the race detector (go test
	// -race) catches map access without the mutex.
	rl := NewRateLimiter(1 * time.Hour)
	const n = 100
	var wins int64
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			if ok, _ := rl.Allow("shared"); ok {
				mu.Lock()
				wins++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if wins != 1 {
		t.Fatalf("expected exactly 1 allowed call, got %d", wins)
	}
}
