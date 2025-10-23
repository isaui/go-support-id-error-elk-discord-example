package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	errorid "github.com/isaui/go-support-id-error"
)

// Handlers holds all service dependencies
type Handlers struct {
	dbService       *DatabaseService
	userService     *UserService
	paymentService  *PaymentService
	apiService      *ExternalAPIService
	authService     *AuthService
	dangerService   *DangerousService
}

// NewHandlers creates a new handlers instance
func NewHandlers() *Handlers {
	return &Handlers{
		dbService:      NewDatabaseService(),
		userService:    NewUserService(),
		paymentService: NewPaymentService(),
		apiService:     NewExternalAPIService(),
		authService:    NewAuthService(),
		dangerService:  NewDangerousService(),
	}
}

// HealthCheck handles health check endpoint
func (h *Handlers) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"service": "go-support-id-example",
	})
}

// HandleDatabaseError demonstrates database error handling
func (h *Handlers) HandleDatabaseError(c *gin.Context) {
	// Call database service
	err := h.dbService.Connect()
	
	// Real-world error handling pattern
	if err != nil {
		// Wrap error with context and metadata
		wrappedErr := errorid.WrapWithDetails(
			err,
			"failed to connect to PostgreSQL",
			map[string]interface{}{
				"database": "postgres",
				"host":     "db.example.com",
				"port":     5432,
				"timeout":  "30s",
			},
		)

		// Use errorid.WriteError helper from library
		errorid.WriteError(c.Writer, wrappedErr)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "database connected"})
}

// HandleValidationError demonstrates validation error handling
func (h *Handlers) HandleValidationError(c *gin.Context) {
	// Simulate registration attempt
	email := "not-an-email"
	username := "john.doe"

	err := h.userService.RegisterUser(email, username)
	
	if err != nil {
		wrappedErr := errorid.WrapWithDetails(
			err,
			"user registration validation failed",
			map[string]interface{}{
				"field":          "email",
				"provided_value": email,
				"expected":       "valid email format",
				"username":       username,
			},
		)

		// Use errorid.WriteError helper from library
		errorid.WriteError(c.Writer, wrappedErr)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user registered"})
}

// HandleNetworkError demonstrates network error handling
func (h *Handlers) HandleNetworkError(c *gin.Context) {
	// Call external API
	err := h.apiService.CallStripeAPI("/v1/charges")
	
	if err != nil {
		wrappedErr := errorid.WrapWithDetails(
			err,
			"failed to call payment gateway API",
			map[string]interface{}{
				"api":      "stripe",
				"endpoint": "https://api.stripe.com/v1/charges",
				"method":   "POST",
				"timeout":  "10s",
			},
		)

		// Use errorid.WriteError helper from library
		errorid.WriteError(c.Writer, wrappedErr)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API call successful"})
}

// HandleAuthError demonstrates authentication error handling
func (h *Handlers) HandleAuthError(c *gin.Context) {
	// Attempt authentication
	username := "john.doe"
	password := "wrong"

	err := h.authService.AuthenticateUser(username, password)
	
	if err != nil {
		wrappedErr := errorid.WrapWithDetails(
			err,
			"user authentication failed",
			map[string]interface{}{
				"username":   username,
				"ip_address": c.ClientIP(),
				"user_agent": c.Request.UserAgent(),
				"attempts":   3,
			},
		)

		// Use errorid.WriteError helper from library
		errorid.WriteError(c.Writer, wrappedErr)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "authentication successful"})
}

// HandlePaymentError demonstrates payment error handling
func (h *Handlers) HandlePaymentError(c *gin.Context) {
	// Process payment
	userID := 12345
	amount := 150.00
	cardLast4 := "4242"

	err := h.paymentService.ProcessPayment(userID, amount, cardLast4)
	
	if err != nil {
		wrappedErr := errorid.WrapWithDetails(
			err,
			"payment processing failed",
			map[string]interface{}{
				"user_id":     userID,
				"amount":      amount,
				"currency":    "USD",
				"card_last4":  cardLast4,
				"merchant_id": "merchant_abc123",
			},
		)

		// Use errorid.WriteError helper from library
		errorid.WriteError(c.Writer, wrappedErr)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "payment successful"})
}

// HandlePanicError demonstrates panic that will be caught by Gin's recovery
func (h *Handlers) HandlePanicError(c *gin.Context) {
	// This will panic and be caught by Gin's recovery middleware
	emptySlice := []string{}
	result := h.dangerService.ProcessArray(emptySlice)
	
	c.JSON(http.StatusOK, gin.H{"result": result})
}

// HandleUncaughtPanic demonstrates UNCAUGHT panic (no middleware protection)
// This endpoint intentionally bypasses error handling to show what happens
// when panic is not caught by recovery middleware
func (h *Handlers) HandleUncaughtPanic(c *gin.Context) {
	// WARNING: This will crash the application if not caught by middleware!
	// Demonstrates the importance of recovery middleware
	h.dangerService.UncaughtPanicOperation()
	
	// This line will never be reached
	c.JSON(http.StatusOK, gin.H{"message": "success"})
}
