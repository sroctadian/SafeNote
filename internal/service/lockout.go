package service

import (
	"sync"
	"time"
)

// LockoutTracker enforces a temporary cooldown after too many failed
// unlock attempts on a given note, per spec section 4. State is kept
// in-memory only (never persisted) and is per-process, which is
// sufficient for a single-user offline desktop app.
type LockoutTracker struct {
	mu           sync.Mutex
	limit        int
	cooldown     time.Duration
	failures     map[string]int
	lockedUntil  map[string]time.Time
}

func NewLockoutTracker(limit int, cooldown time.Duration) *LockoutTracker {
	if limit <= 0 {
		limit = 5
	}
	if cooldown <= 0 {
		cooldown = 30 * time.Second
	}
	return &LockoutTracker{
		limit:       limit,
		cooldown:    cooldown,
		failures:    make(map[string]int),
		lockedUntil: make(map[string]time.Time),
	}
}

// Check returns true and the remaining wait if key is currently locked out.
func (l *LockoutTracker) Check(key string) (locked bool, remaining time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	until, ok := l.lockedUntil[key]
	if !ok {
		return false, 0
	}
	if time.Now().Before(until) {
		return true, time.Until(until)
	}
	// Cooldown expired: reset state for this key.
	delete(l.lockedUntil, key)
	delete(l.failures, key)
	return false, 0
}

// RecordFailure increments the failure count for key and, upon reaching
// the configured limit, starts a cooldown window.
func (l *LockoutTracker) RecordFailure(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.failures[key]++
	if l.failures[key] >= l.limit {
		l.lockedUntil[key] = time.Now().Add(l.cooldown)
	}
}

// RecordSuccess clears any failure state for key (e.g. after a correct PIN).
func (l *LockoutTracker) RecordSuccess(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.failures, key)
	delete(l.lockedUntil, key)
}
