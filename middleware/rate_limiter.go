package middleware

import (
	"fmt"
	"net/http"
	"time"

	"rate-limiter/limiter"

	"github.com/gin-gonic/gin"
)

func RateLimiterMiddleware(limiter *limiter.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("API_KEY")
		clientIP := c.ClientIP()

		var allowed bool
		var remaining int
		var err error

		if token != "" {
			allowed, remaining, err = limiter.CheckToken(token)
		} else {
			allowed, remaining, err = limiter.CheckIP(clientIP)
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Internal server error while checking rate limit",
			})
			c.Abort()
			return
		}

		c.Header("X-Ratelimit-Remaining", fmt.Sprintf("%d", remaining))

		if !allowed {
			blockDuration := time.Duration(0)
			if token != "" {
				_, blockDuration, _ = getBlockInfoForKey("token:"+token, limiter)
			} else {
				_, blockDuration, _ = getBlockInfoForKey("ip:"+clientIP, limiter)
			}

			c.Header("Retry-After", string(int(blockDuration.Seconds())))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"message": "you have reached the maximum number of requests or actions allowed within a certain time frame",
			})

			c.Abort()

			return
		}

		c.Next()
	}
}

func getBlockInfoForKey(key string, limiter *limiter.RateLimiter) (bool, time.Duration, error) {
	return true, 5 * time.Minute, nil
}
