package middleware

import (
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.Mutex

	rate    rate.Limit
	burst   int
	message string
	logName string
}

func NewRateLimiter(
	requestsPerMinute int,
	message string,
	logName string,
) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     rate.Limit(requestsPerMinute) / 60,
		burst:    requestsPerMinute,
		message:  message,
		logName:  logName,
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)

		rl.mu.Lock()

		limiter, exists := rl.limiters[ip]

		if !exists {
			limiter = rate.NewLimiter(rl.rate, rl.burst)
			rl.limiters[ip] = limiter
		}

		allowed := limiter.Allow()

		rl.mu.Unlock()

		if !allowed {
			log.Printf(
				"[%s RATE LIMIT EXCEEDED] ip=%s endpoint=%s timestamp=%s",
				rl.logName,
				ip,
				r.URL.Path,
				time.Now().UTC().Format(time.RFC3339),
			)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)

			w.Write([]byte(`{"error":"` + rl.message + `"}`))

			return
		}

		next.ServeHTTP(w, r)
	})
}

func getClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return strings.Split(forwarded, ",")[0]
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}