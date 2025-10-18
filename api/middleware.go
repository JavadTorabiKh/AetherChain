package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger middleware for logging requests
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		
		// Process request
		c.Next()
		
		// Calculate response time
		duration := time.Since(start)
		
		// Log request details
		clientIP := c.ClientIP()
		method := c.Request.Method
		path := c.Request.URL.Path
		statusCode := c.Writer.Status()
		
		// Log the request
		if statusCode >= 400 {
			// Log errors with more visibility
			println("❌ [API] %s | %3d | %13v | %15s | %-7s %s",
				time.Now().Format("2006/01/02 - 15:04:05"),
				statusCode,
				duration,
				clientIP,
				method,
				path,
			)
		} else {
			println("✅ [API] %s | %3d | %13v | %15s | %-7s %s",
				time.Now().Format("2006/01/02 - 15:04:05"),
				statusCode,
				duration,
				clientIP,
				method,
				path,
			)
		}
	}
}

// CORS middleware for Cross-Origin Resource Sharing
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RateLimiter middleware for basic rate limiting
func RateLimiter() gin.HandlerFunc {
	// Simple in-memory rate limiter
	// In production, use a more sophisticated solution like redis
	visitors := make(map[string]*rateInfo)
	
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		
		// Cleanup old entries (every 1000 requests)
		if len(visitors) > 1000 {
			cleanupRateLimit(visitors)
		}
		
		// Get or create rate info for this IP
		info, exists := visitors[clientIP]
		if !exists {
			info = &rateInfo{
				LastSeen: time.Now(),
				Count:    0,
			}
			visitors[clientIP] = info
		}
		
		// Reset counter if more than 1 minute has passed
		if time.Since(info.LastSeen) > time.Minute {
			info.Count = 0
			info.LastSeen = time.Now()
		}
		
		// Check rate limit (100 requests per minute)
		if info.Count >= 100 {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}
		
		// Increment counter
		info.Count++
		c.Next()
	}
}

// rateInfo stores rate limiting information for an IP
type rateInfo struct {
	LastSeen time.Time
	Count    int
}

// cleanupRateLimit removes old entries from the rate limit map
func cleanupRateLimit(visitors map[string]*rateInfo) {
	now := time.Now()
	for ip, info := range visitors {
		if now.Sub(info.LastSeen) > 5*time.Minute {
			delete(visitors, ip)
		}
	}
}

// AuthMiddleware for API authentication (placeholder)
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// In production, this would validate API keys or JWT tokens
		// For now, it's a placeholder that always allows access
		
		// Example: Check for API key in header
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			// For now, we'll allow requests without API keys
			// In production, you might want to require authentication
			// c.JSON(http.StatusUnauthorized, gin.H{"error": "API key required"})
			// c.Abort()
			// return
		}
		
		c.Next()
	}
}

// Recovery middleware for handling panics
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic
				println("🚨 [PANIC] recovered from panic: %v", err)
				
				// Return error response
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   "Internal server error",
				})
				
				c.Abort()
			}
		}()
		
		c.Next()
	}
}