package limiter

import (
	"fmt"
	"time"
)

type LimiterStrategyInterface interface {
	Check(key string, limit int, window time.Duration) (bool, int, error)

	Block(key string, duration time.Duration) error

	IsBlocked(key string) (bool, error)

	Close() error
}

type RateLimiter struct {
	strategy         LimiterStrategyInterface
	ipMaxRequests    int
	tokenMaxRequests int
	timeWindow       time.Duration
	blockDuration    time.Duration
}

func NewRateLimiter(
	strategy LimiterStrategyInterface,
	ipMaxRequests int,
	tokenMaxRequests int,
	timeWindow time.Duration,
	blockDuration time.Duration,
) *RateLimiter {
	return &RateLimiter{
		strategy:         strategy,
		ipMaxRequests:    ipMaxRequests,
		tokenMaxRequests: tokenMaxRequests,
		timeWindow:       timeWindow,
		blockDuration:    blockDuration,
	}
}

func (rl *RateLimiter) CheckIP(ip string) (bool, int, error) {
	key := fmt.Sprintf("ip:%s", ip)

	blocked, err := rl.strategy.IsBlocked(key)
	if err != nil {
		return false, 0, err
	}
	if blocked {
		return false, 0, nil
	}

	allowed, remaining, err := rl.strategy.Check(key, rl.ipMaxRequests, rl.timeWindow)
	if err != nil {
		return false, 0, err
	}

	if !allowed {
		if err := rl.strategy.Block(key, rl.blockDuration); err != nil {
			return false, 0, err
		}
	}

	return allowed, remaining, nil
}

func (rl *RateLimiter) CheckToken(token string) (bool, int, error) {
	key := fmt.Sprintf("token:%s", token)

	blocked, err := rl.strategy.IsBlocked(key)
	if err != nil {
		return false, 0, err
	}
	if blocked {
		return false, 0, nil
	}

	allowed, remaining, err := rl.strategy.Check(key, rl.tokenMaxRequests, rl.timeWindow)
	if err != nil {
		return false, 0, err
	}

	if !allowed {
		if err := rl.strategy.Block(key, rl.blockDuration); err != nil {
			return false, 0, err
		}
	}

	return allowed, remaining, nil
}

func (rl *RateLimiter) Close() error {
	return rl.strategy.Close()
}
