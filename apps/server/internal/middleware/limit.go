package middleware

import (
	"sync"
	"time"
)

type window struct {
	count int
	until time.Time
}
type Limiter struct {
	mu              sync.Mutex
	peers           map[string]window
	Now             func() time.Time
	Duration        time.Duration
	Limit, Capacity int
}

func NewLimiter() *Limiter {
	return &Limiter{peers: map[string]window{}, Now: time.Now, Duration: time.Minute, Limit: 20, Capacity: 4096}
}

// Allow counts failed attempts only. Authorized sessions can still use the peer.
func (l *Limiter) Allow(peer string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.Now()
	v, ok := l.peers[peer]
	if !ok && len(l.peers) >= l.Capacity {
		for k, w := range l.peers {
			if !now.Before(w.until) {
				delete(l.peers, k)
			}
		}
		if len(l.peers) >= l.Capacity {
			return false
		}
	}
	if !now.Before(v.until) {
		v = window{until: now.Add(l.Duration)}
	}
	if v.count >= l.Limit {
		return false
	}
	v.count++
	l.peers[peer] = v
	return true
}

func (l *Limiter) Exhausted(peer string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	w := l.peers[peer]
	return l.Now().Before(w.until) && w.count >= l.Limit
}
