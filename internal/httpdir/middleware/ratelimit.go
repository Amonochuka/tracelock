package middleware

import (
	"net/http"
	"sync"
	"time"
)

type client struct {
	tokens   float64
	lastSeen time.Time
}

type RateLimiter struct {
	mu        sync.Mutex
	clients   map[string]*client
	rate      float64 // tokens per second
	maxTokens float64
}

func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	r1 := &RateLimiter{
		clients:   make(map[string]*client),
		rate:      float64(requestsPerMinute) / 60.0,
		maxTokens: float64(requestsPerMinute),
	}

	// clean up old clients
	go r1.cleanup()
	return r1
}

// check if IP has tokensa remaining
func (r1 *RateLimiter) allow(ip string) bool {
	r1.mu.Lock()
	defer r1.mu.Unlock()

	c, exists := r1.clients[ip]
	if !exists {
		r1.clients[ip] = &client{tokens: r1.maxTokens - 1, lastSeen: time.Now()}
		return true
	}

	// refill tokens based on time passed
	elapsed := time.Since(c.lastSeen).Seconds()
	c.tokens += elapsed * r1.rate
	if c.tokens > r1.maxTokens {
		c.tokens = r1.maxTokens
	}
	c.lastSeen = time.Now()

	if c.tokens < 1 {
		return false
	}

	c.tokens--
	return true
}

// cleanup removes clients not seen in 3 minutes
func (r1 *RateLimiter) cleanup() {
	for {
		time.Sleep(time.Minute)
		r1.mu.Lock()
		for ip, c := range r1.clients {
			if time.Since(c.lastSeen) > 3*time.Minute {
				delete(r1.clients, ip)
			}
		}
		r1.mu.Unlock()
	}
}

// middleware returns an http.handler that rate limits by IP
func (r1 *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ip := r.RemoteAddr
		ip := getIP(r) // for render purposes

		if !r1.allow(ip) {
			http.Error(w, `{"error":"too many requests"}`, http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// One thing to note — r.RemoteAddr on Render will return the proxy IP,
// not the real client IP. Fix that by checking the X-Forwarded-For header
func getIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}
	return ip
}
