package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

// ErrorBot periodically hits error endpoints to generate errors
type ErrorBot struct {
	baseURL    string
	interval   time.Duration
	httpClient *http.Client
	stopChan   chan struct{}
}

// NewErrorBot creates a new error bot instance
func NewErrorBot(baseURL string, interval time.Duration) *ErrorBot {
	return &ErrorBot{
		baseURL:  baseURL,
		interval: interval,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		stopChan: make(chan struct{}),
	}
}

// Start begins the bot's periodic error generation
func (b *ErrorBot) Start() {
	fmt.Printf("🤖 Error Bot started - hitting endpoints every %v\n", b.interval)
	
	ticker := time.NewTicker(b.interval)
	defer ticker.Stop()

	// Hit immediately on start
	b.hitRandomEndpoint()

	for {
		select {
		case <-ticker.C:
			b.hitRandomEndpoint()
		case <-b.stopChan:
			fmt.Println("🤖 Error Bot stopped")
			return
		}
	}
}

// Stop stops the bot
func (b *ErrorBot) Stop() {
	close(b.stopChan)
}

// hitRandomEndpoint selects and hits a random error endpoint
func (b *ErrorBot) hitRandomEndpoint() {
	endpoints := []string{
		"/api/error/database",
		"/api/error/validation",
		"/api/error/network",
		"/api/error/auth",
		"/api/error/payment",
		"/api/error/panic",
		// Note: Uncomment below to test uncaught panic (will crash server!)
		// "/api/error/uncaught-panic",
	}

	// Select random endpoint
	endpoint := endpoints[rand.Intn(len(endpoints))]
	url := b.baseURL + endpoint

	fmt.Printf("🎯 Bot hitting: %s\n", endpoint)

	resp, err := b.httpClient.Get(url)
	if err != nil {
		fmt.Printf("❌ Bot request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	
	if resp.StatusCode >= 400 {
		fmt.Printf("✅ Bot triggered error (Status: %d)\n", resp.StatusCode)
		fmt.Printf("   Response: %s\n", string(body))
	} else {
		fmt.Printf("⚠️  Bot expected error but got: %d\n", resp.StatusCode)
	}
}

// StartErrorBot starts the error bot as a background goroutine
func StartErrorBot(baseURL string, interval time.Duration) *ErrorBot {
	bot := NewErrorBot(baseURL, interval)
	
	go bot.Start()
	
	return bot
}
