package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kunal/life-log/backend/configs"
	"github.com/kunal/life-log/backend/internal/auth"
)

type Middleware struct {
	authSvc interface {
		ValidateAccessToken(string) (*auth.Claims, error)
	}
	cfg *configs.Config
}

func New(authSvc interface{ ValidateAccessToken(string) (*auth.Claims, error) }) *Middleware {
	return &Middleware{authSvc: authSvc}
}

func (m *Middleware) authenticate(c *gin.Context) *auth.Claims {
	token, ok := extractBearerToken(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return nil
	}
	claims, err := m.authSvc.ValidateAccessToken(token)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return nil
	}
	return claims
}

func (m *Middleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := m.authenticate(c)
		if c.IsAborted() {
			return
		}
		c.Set("userID", claims.Subject)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func (m *Middleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := m.authenticate(c)
		if c.IsAborted() {
			return
		}
		if claims.Role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}
		c.Set("userID", claims.Subject)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func (m *Middleware) RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := m.authenticate(c)
		if c.IsAborted() {
			return
		}
		for _, role := range roles {
			if claims.Role == role {
				c.Set("userID", claims.Subject)
				c.Set("email", claims.Email)
				c.Set("role", claims.Role)
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
	}
}

func (m *Middleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := extractBearerToken(c)
		if !ok {
			c.Next()
			return
		}
		claims, err := m.authSvc.ValidateAccessToken(token)
		if err != nil {
			c.Next()
			return
		}
		c.Set("userID", claims.Subject)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Next()
	}
}

type rateLimiterEntry struct {
	count   int
	resetAt time.Time
}

type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*rateLimiterEntry
	rate     int
	burst    int
	window   time.Duration
}

func NewRateLimiter(cfg *configs.RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		visitors: make(map[string]*rateLimiterEntry),
		rate:     cfg.RequestsPerMin,
		burst:    cfg.BurstSize,
		window:   time.Minute,
	}
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := clientIP(c)
		rl.mu.Lock()
		entry, exists := rl.visitors[ip]
		now := time.Now()
		if !exists || now.After(entry.resetAt) {
			rl.visitors[ip] = &rateLimiterEntry{
				count:   1,
				resetAt: now.Add(rl.window),
			}
			rl.mu.Unlock()
			c.Next()
			return
		}
		entry.count++
		if entry.count > rl.rate {
			entry.count--
			rl.mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": entry.resetAt.Sub(now).Seconds(),
			})
			return
		}
		rl.mu.Unlock()
		c.Next()
	}
}

func clientIP(c *gin.Context) string {
	if fwd := c.GetHeader("X-Forwarded-For"); fwd != "" {
		parts := strings.Split(fwd, ",")
		return strings.TrimSpace(parts[0])
	}
	if realIP := c.GetHeader("X-Real-IP"); realIP != "" {
		return realIP
	}
	return c.ClientIP()
}

func extractBearerToken(c *gin.Context) (string, bool) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", false
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return "", false
	}
	return parts[1], true
}

func SecureHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("requestID", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

func CSRFMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodPost ||
			c.Request.Method == http.MethodPut ||
			c.Request.Method == http.MethodPatch ||
			c.Request.Method == http.MethodDelete {
			token := c.GetHeader("X-CSRF-Token")
			if token == "" {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "CSRF token required"})
				return
			}
		}
		c.Next()
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Request-ID, X-CSRF-Token")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "43200")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
