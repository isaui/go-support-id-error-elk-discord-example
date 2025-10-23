package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	errorid "github.com/isaui/go-support-id-error"
)

// GinRecoveryMiddleware wraps errorid.RecoveryMiddleware for use with Gin
// This demonstrates how to use the library's built-in RecoveryMiddleware
func GinRecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a wrapper that adapts Gin to http.Handler
		handler := errorid.RecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Continue with Gin's processing
			c.Next()
		}))

		// Execute the wrapped handler
		handler.ServeHTTP(c.Writer, c.Request)
	}
}
