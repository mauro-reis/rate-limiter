package limiter

import (
	"sync"
	"time"
)

type MemoryStrategy struct {
	mu           sync.Mutex
	requests     map[string][]int64
	blockedUntil map[string]int64
}

func NewMemoryStrategy() *MemoryStrategy {
	return &MemoryStrategy{
		requests:     make(map[string][]int64),
		blockedUntil: make(map[string]int64),
	}
}

func (ms *MemoryStrategy) Check(key string, limit int, window time.Duration) (bool, int, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	now := time.Now().UnixNano()
	windowStart := now - window.Nanoseconds()

	if timestamps, ok := ms.requests[key]; ok {
		var validTimestamps []int64
		for _, ts := range timestamps {
			if ts > windowStart {
				validTimestamps = append(validTimestamps, ts)
			}
		}
		ms.requests[key] = validTimestamps
	} else {
		ms.requests[key] = []int64{}
	}

	count := len(ms.requests[key])
	if count >= limit {
		return false, 0, nil
	}

	ms.requests[key] = append(ms.requests[key], now)

	return true, limit - count - 1, nil
}

func (ms *MemoryStrategy) Block(key string, duration time.Duration) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.blockedUntil[key] = time.Now().Add(duration).UnixNano()
	return nil
}

func (ms *MemoryStrategy) IsBlocked(key string) (bool, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if expiresAt, ok := ms.blockedUntil[key]; ok {
		if expiresAt > time.Now().UnixNano() {
			return true, nil
		}

		delete(ms.blockedUntil, key)
	}

	return false, nil
}

func (ms *MemoryStrategy) Close() error {
	return nil
}
