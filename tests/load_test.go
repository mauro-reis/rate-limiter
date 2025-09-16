package tests

import (
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"rate-limiter/limiter"
	"rate-limiter/middleware"

	"github.com/gin-gonic/gin"
)

func BenchmarkRateLimiter(b *testing.B) {
	redisStrategy, err := limiter.NewRedisStrategy(
		"localhost", "6379", "", 0,
	)

	if err != nil {
		b.Fatalf("Failed to create Redis strategy: %v", err)
	}
	defer redisStrategy.Close()

	rateLimiter := limiter.NewRateLimiter(
		redisStrategy,
		1000,
		2000,
		1*time.Second,
		5*time.Second,
	)

	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(middleware.RateLimiterMiddleware(rateLimiter))
	router.GET("/benchmark", func(c *gin.Context) {
		c.String(200, "OK")
	})

	server := &http.Server{
		Addr:    ":18080",
		Handler: router,
	}

	go server.ListenAndServe()

	defer server.Close()

	time.Sleep(100 * time.Millisecond)

	b.Run("SingleIP", func(b *testing.B) {
		client := &http.Client{
			Timeout: 5 * time.Second,
		}

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			resp, err := client.Get("http://localhost:18080/benchmark")
			if err != nil {
				b.Fatalf("Request failed: %v", err)
			}
			resp.Body.Close()
		}
	})

	b.Run("MultipleIPs", func(b *testing.B) {
		client := &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				MaxIdleConnsPerHost: 100,
			},
		}

		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				resp, err := client.Get("http://localhost:18080/benchmark")
				if err != nil {
					b.Fatalf("Request failed: %v", err)
				}
				resp.Body.Close()
			}
		})
	})

	b.Run("TokenVsIP", func(b *testing.B) {

		rateLimiter = limiter.NewRateLimiter(
			redisStrategy,
			10,
			100,
			1*time.Second,
			5*time.Second,
		)

		router = gin.New()

		router.Use(middleware.RateLimiterMiddleware(rateLimiter))

		router.GET("/token-vs-ip", func(c *gin.Context) {
			c.String(200, "OK")
		})

		tokenServer := &http.Server{
			Addr:    ":18081",
			Handler: router,
		}

		go tokenServer.ListenAndServe()

		defer tokenServer.Close()

		// Apenas pra driblar, tive que dar um sleep, pois estava dando problema executando o código antes de subir o server.
		time.Sleep(150 * time.Millisecond)

		numClients := 20

		cases := []struct {
			name     string
			useToken bool
		}{
			{"IP", false},
			{"Token", true},
			{"Mixed", false},
		}

		for _, tc := range cases {
			b.Run(tc.name, func(b *testing.B) {

				wg := sync.WaitGroup{}

				wg.Add(numClients)

				requestsPerClient := b.N / numClients

				if requestsPerClient < 1 {
					requestsPerClient = 1
				}

				for i := 0; i < numClients; i++ {
					go func(clientID int) {
						defer wg.Done()

						client := &http.Client{
							Timeout: 5 * time.Second,
						}

						useToken := tc.useToken

						if tc.name == "Mixed" {
							useToken = clientID%2 == 0
						}

						for j := 0; j < requestsPerClient; j++ {

							req, _ := http.NewRequest("GET", "http://localhost:18081/token-vs-ip", nil)

							if useToken {
								token := fmt.Sprintf("token-%d", clientID)

								req.Header.Set("API_KEY", token)
							}

							resp, err := client.Do(req)

							if err != nil {
								continue
							}

							resp.Body.Close()
						}
					}(i)
				}

				wg.Wait()
			})
		}
	})
}
