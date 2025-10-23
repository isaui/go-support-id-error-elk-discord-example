package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	errorid "github.com/isaui/go-support-id-error"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	godotenv.Load()

	// Initialize integrations
	discordWebhook := NewDiscordWebhook(os.Getenv("DISCORD_WEBHOOK_URL"))
	elkLogger := NewELKLogger(os.Getenv("ELK_URL"))

	// Configure error-id library
	configureErrorTracking(discordWebhook, elkLogger)

	// Setup server
	router := setupServer()

	// Start error bot
	bot := startErrorBot()
	defer bot.Stop()

	// Graceful shutdown
	setupGracefulShutdown(bot)

	// Start server
	printStartupInfo()
	router.Run(":" + getPort())
}

func setupServer() *gin.Engine {
	// Create router WITHOUT default middleware
	router := gin.New()
	
	// Use Gin's logger
	router.Use(gin.Logger())
	
	// Use errorid.RecoveryMiddleware via adapter (library's middleware!)
	router.Use(GinRecoveryMiddleware())

	// Initialize handlers with service dependencies
	handlers := NewHandlers()

	// Setup all routes
	SetupRoutes(router, handlers)

	return router
}

// configureErrorTracking sets up error-id library with integrations
func configureErrorTracking(discord *DiscordWebhook, elk *ELKLogger) {
	errorid.Configure(errorid.Config{
		OnError: func(err *errorid.ErrorWithID) {
			// Send to Discord
			discord.SendErrorNotification(err)
		},
		AsyncCallback:     true, // Non-blocking
		Logger:            elk,
		IncludeStackTrace: true,
		Environment:       getEnvironment(),
	})
}

// startErrorBot starts the error bot goroutine
func startErrorBot() *ErrorBot {
	botInterval := 30 * time.Second
	if interval := os.Getenv("BOT_INTERVAL"); interval != "" {
		if d, err := time.ParseDuration(interval); err == nil {
			botInterval = d
		}
	}

	baseURL := fmt.Sprintf("http://localhost:%s", getPort())
	return StartErrorBot(baseURL, botInterval)
}

// setupGracefulShutdown configures graceful shutdown handlers
func setupGracefulShutdown(bot *ErrorBot) {
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		fmt.Println("\nShutting down server...")
		bot.Stop()
		os.Exit(0)
	}()
}

// printStartupInfo prints server startup information
func printStartupInfo() {
	port := getPort()
	separator := "============================================================"
	
	fmt.Printf("\n%s\n", separator)
	fmt.Printf("Server starting on port %s\n", port)
	fmt.Printf("ELK URL: %s\n", os.Getenv("ELK_URL"))
	fmt.Printf("Discord Webhook: %s\n", maskWebhookURL(os.Getenv("DISCORD_WEBHOOK_URL")))
	fmt.Printf("Error Bot interval: %s\n", os.Getenv("BOT_INTERVAL"))
	fmt.Printf("Environment: %s\n", getEnvironment())
	fmt.Printf("%s\n\n", separator)
	
	fmt.Println("Available endpoints:")
	fmt.Println("   GET  /health")
	fmt.Println("   GET  /api/error/database")
	fmt.Println("   GET  /api/error/validation")
	fmt.Println("   GET  /api/error/network")
	fmt.Println("   GET  /api/error/auth")
	fmt.Println("   GET  /api/error/payment")
	fmt.Println("   GET  /api/error/panic")
	fmt.Println("   GET  /api/error/uncaught-panic")
	fmt.Printf("\n%s\n\n", separator)
}

// Helper functions

func getEnvironment() string {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		return "development"
	}
	return env
}

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		return "8080"
	}
	return port
}

func maskWebhookURL(url string) string {
	if url == "" {
		return "not configured"
	}
	if len(url) > 50 {
		return url[:50] + "..."
	}
	return url
}
