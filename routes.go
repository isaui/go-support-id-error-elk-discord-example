package main

import (
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all application routes
func SetupRoutes(router *gin.Engine, handlers *Handlers) {
	// Health check endpoint (no middleware)
	router.GET("/health", handlers.HealthCheck)

	// API group
	api := router.Group("/api")
	{
		// Error simulation endpoints
		errorGroup := api.Group("/error")
		{
			// Database error - demonstrates service call with error handling
			errorGroup.GET("/database", handlers.HandleDatabaseError)
			
			// Validation error - demonstrates input validation
			errorGroup.GET("/validation", handlers.HandleValidationError)
			
			// Network error - demonstrates external API failure
			errorGroup.GET("/network", handlers.HandleNetworkError)
			
			// Authentication error - demonstrates auth failure
			errorGroup.GET("/auth", handlers.HandleAuthError)
			
			// Payment error - demonstrates payment processing failure
			errorGroup.GET("/payment", handlers.HandlePaymentError)
			
			// Panic error - demonstrates panic caught by errorid.RecoveryMiddleware
			errorGroup.GET("/panic", handlers.HandlePanicError)
			
			// Uncaught panic - also caught by errorid.RecoveryMiddleware (global)
			// Returns JSON: {"error_id": "ERR-xxx", "message": "...", "timestamp": ...}
			errorGroup.GET("/uncaught-panic", handlers.HandleUncaughtPanic)
		}
	}
}
