package ratelimit

import (
	"sync"
	"time"
)

type TokenBucket struct {
	capacity   int
	tokens     int
	refillRate int
	lastRefill time.Time
	mu         sync.Mutex
}

func NewTokenBucket(capacity, refillRate int) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)

	tokensToAdd := int(elapsed.Seconds()) * tb.refillRate
	if tokensToAdd > 0 {
		tb.tokens += tokensToAdd
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		tb.lastRefill = now
	}
}

type RateLimiter struct {
	capacity   int
	refillRate int
	buckets    map[string]*TokenBucket
	mu         sync.Mutex
}

func NewRateLimiter(capacity, refillRate int) *RateLimiter {
	return &RateLimiter{
		capacity:   capacity,
		refillRate: refillRate,
		buckets:    make(map[string]*TokenBucket),
	}
}

func (rl *RateLimiter) Allow(key string) bool {
	bucket := rl.getBucket(key)
	return bucket.Allow()
}

func (rl *RateLimiter) getBucket(key string) *TokenBucket {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if bucket, ok := rl.buckets[key]; ok {
		return bucket
	}

	bucket := NewTokenBucket(rl.capacity, rl.refillRate)
	rl.buckets[key] = bucket
	return bucket
}
