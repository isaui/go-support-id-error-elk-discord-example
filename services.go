package main

import (
	"errors"
	"fmt"
	"time"
)

// DatabaseService simulates database operations
type DatabaseService struct{}

func NewDatabaseService() *DatabaseService {
	return &DatabaseService{}
}

// Connect simulates database connection that can fail
func (s *DatabaseService) Connect() error {
	// Simulate connection failure
	return errors.New("connection to database timed out after 30s")
}

// UserService simulates user-related operations
type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}

// ValidateEmail simulates email validation
func (s *UserService) ValidateEmail(email string) error {
	if email == "" || email == "not-an-email" {
		return errors.New("email format is invalid")
	}
	return nil
}

// RegisterUser simulates user registration
func (s *UserService) RegisterUser(email, username string) error {
	if err := s.ValidateEmail(email); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	// Simulate other registration logic
	return nil
}

// PaymentService simulates payment processing
type PaymentService struct{}

func NewPaymentService() *PaymentService {
	return &PaymentService{}
}

// ProcessPayment simulates payment processing that can fail
func (s *PaymentService) ProcessPayment(userID int, amount float64, cardLast4 string) error {
	// Simulate insufficient funds
	if amount > 100 {
		return errors.New("insufficient funds")
	}
	return nil
}

// ExternalAPIService simulates external API calls
type ExternalAPIService struct{}

func NewExternalAPIService() *ExternalAPIService {
	return &ExternalAPIService{}
}

// CallStripeAPI simulates calling Stripe API
func (s *ExternalAPIService) CallStripeAPI(endpoint string) error {
	// Simulate connection refused
	time.Sleep(100 * time.Millisecond) // Simulate network delay
	return errors.New("connection refused")
}

// AuthService simulates authentication
type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

// AuthenticateUser simulates user authentication
func (s *AuthService) AuthenticateUser(username, password string) error {
	// Simulate invalid credentials
	if username == "john.doe" && password == "wrong" {
		return errors.New("invalid credentials")
	}
	return nil
}

// DangerousService simulates operations that can panic
type DangerousService struct{}

func NewDangerousService() *DangerousService {
	return &DangerousService{}
}

// ProcessArray simulates array operation that can panic
func (s *DangerousService) ProcessArray(data []string) string {
	// This will panic if array is empty
	return data[0] // Intentional panic for demonstration
}

// UncaughtPanicOperation simulates uncaught panic (no recovery)
func (s *DangerousService) UncaughtPanicOperation() {
	// This will cause nil pointer dereference
	var ptr *string
	_ = *ptr // Panic: invalid memory address or nil pointer dereference
}
