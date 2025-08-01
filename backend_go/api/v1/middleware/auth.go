package middleware

import (
	"net/http"
	"strings"
	"fantasy-esports-backend/utils"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Authorization header required",
				"code":    "UNAUTHORIZED",
			})
			c.Abort()
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		claims, err := utils.ValidateToken(tokenString, jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid token",
				"code":    "INVALID_TOKEN",
			})
			c.Abort()
			return
		}

		if claims.TokenType != "access" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid token type",
				"code":    "INVALID_TOKEN_TYPE",
			})
			c.Abort()
			return
		}

		// Set user context
		c.Set("user_id", claims.UserID)
		c.Set("mobile", claims.Mobile)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func AdminAuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for login endpoint
		if c.Request.URL.Path == "/api/v1/admin/login" {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Authorization header required",
				"code":    "UNAUTHORIZED",
			})
			c.Abort()
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		claims, err := utils.ValidateAdminToken(tokenString, jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid admin token",
				"code":    "INVALID_TOKEN",
			})
			c.Abort()
			return
		}

		// Set admin context
		c.Set("admin_id", claims.AdminID)
		c.Set("username", claims.Username)
		c.Set("admin_role", claims.Role)
		c.Next()
	}
}

func AdminWebSocketMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Token required for WebSocket connection",
				"code":    "UNAUTHORIZED",
			})
			c.Abort()
			return
		}

		claims, err := utils.ValidateAdminToken(token, jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid admin token",
				"code":    "INVALID_TOKEN",
			})
			c.Abort()
			return
		}

		// Set admin context for WebSocket
		c.Set("admin_id", claims.AdminID)
		c.Set("username", claims.Username)
		c.Set("admin_role", claims.Role)
		c.Next()
	}
}