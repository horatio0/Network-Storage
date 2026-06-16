package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger(logger *log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logger.Printf("%s %s %s %s", c.Request.Method, c.Request.URL.Path, time.Since(start), c.ClientIP())
	}
}
