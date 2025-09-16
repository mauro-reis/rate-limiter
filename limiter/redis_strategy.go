package limiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStrategy struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisStrategy(host, port, password string, db int) (*RedisStrategy, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       db,
	})

	ctx := context.Background()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisStrategy{
		client: client,
		ctx:    ctx,
	}, nil
}

func (rs *RedisStrategy) Check(key string, limit int, window time.Duration) (bool, int, error) {
	now := time.Now().UnixNano() / int64(time.Millisecond)
	windowStart := now - int64(window.Milliseconds())

	script := `
        local key = KEYS[1]
        local now = tonumber(ARGV[1])
        local windowStart = tonumber(ARGV[2])
        local limit = tonumber(ARGV[3])
        local windowSize = tonumber(ARGV[4])
        
        redis.call('ZREMRANGEBYSCORE', key, 0, windowStart)
        
        local count = redis.call('ZCARD', key)
        
        local allowed = count < limit
        
        if allowed then
            redis.call('ZADD', key, now, now..':'..math.random())
            redis.call('EXPIRE', key, windowSize)
            count = count + 1
        end
        
        return {allowed and 1 or 0, limit - count}
    `

	result, err := rs.client.Eval(rs.ctx, script, []string{key},
		now, windowStart, limit, int(window.Seconds())).Result()
	if err != nil {
		return false, 0, fmt.Errorf("failed to execute Redis script: %w", err)
	}

	results, ok := result.([]interface{})
	if !ok || len(results) != 2 {
		return false, 0, fmt.Errorf("unexpected result from Redis")
	}

	allowed := results[0].(int64) == 1
	remaining := int(results[1].(int64))

	return allowed, remaining, nil
}

func (rs *RedisStrategy) Block(key string, duration time.Duration) error {
	blockKey := fmt.Sprintf("block:%s", key)
	err := rs.client.Set(rs.ctx, blockKey, 1, duration).Err()
	if err != nil {
		return fmt.Errorf("failed to block key: %w", err)
	}
	return nil
}

func (rs *RedisStrategy) IsBlocked(key string) (bool, error) {
	blockKey := fmt.Sprintf("block:%s", key)
	result, err := rs.client.Exists(rs.ctx, blockKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check if key is blocked: %w", err)
	}
	return result > 0, nil
}

func (rs *RedisStrategy) Close() error {
	return rs.client.Close()
}
