package config

import (
	"log"
	"os"
	"strconv"

	"github.com/lpernett/godotenv"
)

type Config struct {
	IPMaxRequests        int
	TokenMaxRequests     int
	TimeWindowSeconds    int
	BlockDurationSeconds int
	RedisHost            string
	RedisPort            string
	RedisPassword        string
	RedisDB              int
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	ipMaxReq, _ := strconv.Atoi(getEnv("RATE_LIMITER_IP_MAX_REQUESTS", "10"))
	tokenMaxReq, _ := strconv.Atoi(getEnv("RATE_LIMITER_TOKEN_MAX_REQUESTS", "100"))
	timeWindowSec, _ := strconv.Atoi(getEnv("RATE_LIMITER_TIME_WINDOW_SECONDS", "1"))
	blockDurationSec, _ := strconv.Atoi(getEnv("BLOCK_DURATION_SECONDS", "300"))
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

	return &Config{
		IPMaxRequests:        ipMaxReq,
		TokenMaxRequests:     tokenMaxReq,
		TimeWindowSeconds:    timeWindowSec,
		BlockDurationSeconds: blockDurationSec,
		RedisHost:            getEnv("REDIS_HOST", "localhost"),
		RedisPort:            getEnv("REDIS_PORT", "6379"),
		RedisPassword:        getEnv("REDIS_PASSWORD", ""),
		RedisDB:              redisDB,
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
