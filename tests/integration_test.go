package tests

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"rate-limiter/limiter"
	"rate-limiter/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRateLimiterIntegration(t *testing.T) {
	memoryStrategy := limiter.NewMemoryStrategy()

	rateLimiter := limiter.NewRateLimiter(
		memoryStrategy,
		3,
		5,
		1*time.Second,
		5*time.Second,
	)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RateLimiterMiddleware(rateLimiter))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	t.Run("IP Rate Limiting", func(t *testing.T) {

		// Três requisições no mesmo IP tem que permitir
		for i := 0; i < 3; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "192.168.1.1:1234"
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		}

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Quarta requisição deve ser criticada
		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.Contains(t, w.Body.String(), "maximum number of requests")
	})

	t.Run("Token Rate Limiting", func(t *testing.T) {
		token := "test-token-123"

		// Cinco requisições no mesmo IP tem que permitir
		for i := 0; i < 5; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("API_KEY", token)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		}

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("API_KEY", token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Sexta requisição deve ser criticada
		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.Contains(t, w.Body.String(), "maximum number of requests")
	})

	t.Run("Token Priority Over IP", func(t *testing.T) {
		memoryStrategy = limiter.NewMemoryStrategy()
		rateLimiter = limiter.NewRateLimiter(
			memoryStrategy,
			2,
			10,
			1*time.Second,
			5*time.Second,
		)

		router = gin.New()
		router.Use(middleware.RateLimiterMiddleware(rateLimiter))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		ip := "192.168.1.2:1234"
		token := "high-limit-token"

		for i := 0; i < 2; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = ip
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, fmt.Sprintf("Request %d with IP should be allowed", i+1))
		}

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = ip
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusTooManyRequests, w.Code, "IP should be rate limited after 2 requests")

		for i := 0; i < 5; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = ip
			req.Header.Set("API_KEY", token)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, fmt.Sprintf("Request %d with token should be allowed", i+1))
		}
	})
}
