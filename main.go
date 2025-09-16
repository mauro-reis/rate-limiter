package main

import (
	"log"
	"time"

	"rate-limiter/config"
	"rate-limiter/limiter"
	"rate-limiter/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	redisStrategy, err := limiter.NewRedisStrategy(
		cfg.RedisHost,
		cfg.RedisPort,
		cfg.RedisPassword,
		cfg.RedisDB,
	)
	if err != nil {
		log.Fatalf("Failed to create Redis strategy: %v", err)
	}
	defer redisStrategy.Close()

	rateLimiter := limiter.NewRateLimiter(
		redisStrategy,
		cfg.IPMaxRequests,
		cfg.TokenMaxRequests,
		time.Duration(cfg.TimeWindowSeconds)*time.Second,
		time.Duration(cfg.BlockDurationSeconds)*time.Second,
	)

	router := gin.Default()

	router.Use(middleware.RateLimiterMiddleware(rateLimiter))

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello World!",
		})
	})

	log.Println("Starting server on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
