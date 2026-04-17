package phi

import (
	"sync"
	"time"
)

type PlusCodeLimiter struct {
	mu       sync.Mutex
	limit    int
	counters map[string]*dayCounter
}

type dayCounter struct {
	date  string
	count int
}

func NewPlusCodeLimiter(limit int) *PlusCodeLimiter {
	return &PlusCodeLimiter{limit: limit, counters: make(map[string]*dayCounter)}
}

func (l *PlusCodeLimiter) Allow(userID string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	today := time.Now().UTC().Format("2006-01-02")
	dc, ok := l.counters[userID]
	if !ok || dc.date != today {
		l.counters[userID] = &dayCounter{date: today, count: 1}
		return true
	}
	if dc.count >= l.limit { return false }
	dc.count++
	return true
}
