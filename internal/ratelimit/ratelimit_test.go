package ratelimit

import "testing"

func TestTokenBucketAllow(t *testing.T) {
	b := NewTokenBucket(2, 1)
	if !b.Allow() {
		t.Fatalf("expected first allow to be true")
	}
	if !b.Allow() {
		t.Fatalf("expected second allow to be true")
	}
	if b.Allow() {
		t.Fatalf("expected third allow to be false (capacity reached)")
	}
}

func TestRateLimiterPerKey(t *testing.T) {
	rl := NewRateLimiter(1, 1)
	if !rl.Allow("a") {
		t.Fatalf("expected allow for key a")
	}
	if rl.Allow("a") {
		t.Fatalf("expected second allow for key a to be false")
	}
	if !rl.Allow("b") {
		t.Fatalf("expected allow for key b to be true")
	}
}
