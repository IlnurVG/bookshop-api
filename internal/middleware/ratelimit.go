package middleware

import (
	"net/http"
	"strings"

	"github.com/bookshop/api/internal/pkg/ratelimit"
	"github.com/bookshop/api/pkg/logger"
	"github.com/labstack/echo/v4"
)

// IPRateLimiter limits requests based on client IP address
type IPRateLimiter struct {
	limiter *ratelimit.MultiRateLimiter
	logger  logger.Logger
}

// NewIPRateLimiter creates a new IP-based rate limiter
// limit - maximum number of requests per second from a single IP
func NewIPRateLimiter(limit int, logger logger.Logger) *IPRateLimiter {
	return &IPRateLimiter{
		limiter: ratelimit.NewMultiRateLimiter(limit, 0), // 0 - uses default interval (1 sec)
		logger:  logger,
	}
}

// Middleware creates middleware for request rate limiting
func (rl *IPRateLimiter) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get client IP address
			ip := getClientIP(c.Request())

			// Process request with rate limiting
			err := rl.limiter.Process(ip, func() {
				// This function will be executed
				// only if the limit is not exceeded
			})

			if err != nil {
				rl.logger.Error("Rate limit exceeded", "ip", ip, "path", c.Request().URL.Path)
				return c.JSON(http.StatusTooManyRequests, map[string]string{
					"error": "Too many requests. Please try again later.",
				})
			}

			// Continue request processing
			return next(c)
		}
	}
}

// Stop stops the rate limiter
func (rl *IPRateLimiter) Stop() {
	rl.limiter.Stop()
}

// PathRateLimiter limits requests for specific API paths
type PathRateLimiter struct {
	pathLimiters map[string]*ratelimit.RateLimiter
	defaultLimit int
	logger       logger.Logger
}

// NewPathRateLimiter creates a new path-based rate limiter
// defaultLimit - default requests per second limit
func NewPathRateLimiter(defaultLimit int, logger logger.Logger) *PathRateLimiter {
	return &PathRateLimiter{
		pathLimiters: make(map[string]*ratelimit.RateLimiter),
		defaultLimit: defaultLimit,
		logger:       logger,
	}
}

// SetPathLimit sets the rate limit for a specific path
func (rl *PathRateLimiter) SetPathLimit(path string, limit int) {
	rl.pathLimiters[path] = ratelimit.NewRateLimiter(limit)
}

// Middleware creates middleware for path-based rate limiting
func (rl *PathRateLimiter) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path

			// Get limiter for the path, or create a new one with default limit
			var limiter *ratelimit.RateLimiter
			for pattern, pathLimiter := range rl.pathLimiters {
				if pathMatches(path, pattern) {
					limiter = pathLimiter
					break
				}
			}

			if limiter == nil {
				limiter = ratelimit.NewRateLimiter(rl.defaultLimit)
				rl.pathLimiters[path] = limiter
			}

			// Process request with rate limiting
			err := limiter.Process(func() {
				// This function will be executed
				// only if the limit is not exceeded
			})

			if err != nil {
				rl.logger.Error("Path rate limit exceeded", "path", path)
				return c.JSON(http.StatusTooManyRequests, map[string]string{
					"error": "Too many requests to this resource. Please try again later.",
				})
			}

			// Continue request processing
			return next(c)
		}
	}
}

// Stop stops all rate limiters
func (rl *PathRateLimiter) Stop() {
	for _, limiter := range rl.pathLimiters {
		limiter.Stop()
	}
}

// Helper function to get client IP address
func getClientIP(r *http.Request) string {
	// Check headers that may contain the real client IP
	// behind a proxy or load balancer
	for _, header := range []string{"X-Forwarded-For", "X-Real-IP"} {
		if ip := r.Header.Get(header); ip != "" {
			// X-Forwarded-For may contain a list of IPs separated by commas
			if header == "X-Forwarded-For" {
				ips := strings.Split(ip, ",")
				if len(ips) > 0 {
					return strings.TrimSpace(ips[0])
				}
			}
			return ip
		}
	}

	// If headers don't contain IP, use RemoteAddr
	ip := r.RemoteAddr
	// Remove port from IP address if present
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// Helper function to check if a path matches a pattern
func pathMatches(path, pattern string) bool {
	// Simple implementation, can be replaced with regex
	// or more complex routing logic
	if pattern == path {
		return true
	}

	// Support wildcard at the end of the pattern, like "/api/v1/*"
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(path, pattern[:len(pattern)-1])
	}

	return false
}
