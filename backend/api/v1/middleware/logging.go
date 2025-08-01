package middleware

import (
	"time"
	"fantasy-esports-backend/pkg/logger"
	"github.com/gin-gonic/gin"
)

func RequestLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logger.Info(
			"Request:",
			"method", param.Method,
			"path", param.Path,
			"status", param.StatusCode,
			"latency", param.Latency,
			"ip", param.ClientIP,
			"user_agent", param.Request.UserAgent(),
			"timestamp", param.TimeStamp.Format(time.RFC3339),
		)
		return ""
	})
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			logger.Error("Request error:", err.Err)
			
			c.JSON(500, gin.H{
				"success": false,
				"error":   "Internal server error",
				"code":    "INTERNAL_ERROR",
			})
		}
	}
}