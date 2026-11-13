package auth

import (
	"sync"
	"time"
)

func defaultNow() int64 { return time.Now().UnixNano() }

// Limiter is a per-key fixed-window counter. One process, one map: the
// service runs a single replica in v1.
type Limiter struct {
	mu      sync.Mutex
	limit   int
	window  int64 // nanoseconds
	entries map[string]entry
	now     func() int64
}

type entry struct {
	start int64
	count int
}

// NewLimiter allows limit attempts per key per window.
func NewLimiter(limit int, windowNanos int64) *Limiter {
	return &Limiter{
		limit:   limit,
		window:  windowNanos,
		entries: make(map[string]entry),
		now:     defaultNow,
	}
}

// Allow records one attempt for key and reports whether it is within budget.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	if len(l.entries) > 10000 {
		// ponytail: full sweep under the lock; with one replica and a
		// per-minute window this is cheap. Shard by key if it ever grows.
		for k, e := range l.entries {
			if now-e.start >= l.window {
				delete(l.entries, k)
			}
		}
	}

	e, ok := l.entries[key]
	if !ok || now-e.start >= l.window {
		l.entries[key] = entry{start: now, count: 1}
		return true
	}
	if e.count >= l.limit {
		return false
	}
	e.count++
	l.entries[key] = e
	return true
}
