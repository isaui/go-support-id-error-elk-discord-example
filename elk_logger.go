package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	errorid "github.com/isaui/go-support-id-error"
)

// ELKLogger sends logs to ELK (Elasticsearch/Logstash) cluster
type ELKLogger struct {
	elkURL     string
	httpClient *http.Client
}

// NewELKLogger creates a new ELK logger instance
func NewELKLogger(elkURL string) *ELKLogger {
	return &ELKLogger{
		elkURL: elkURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// Error implements the errorid.Logger interface
func (l *ELKLogger) Error(errorID string, err error, context string, details map[string]interface{}) {
	// Log to stderr
	fmt.Fprintf(os.Stderr, "[ERROR-ID] ID=%s | Context=%s | Error=%v\n", errorID, context, err)
	
	// Send to ELK with structured data
	go l.sendStructuredError(errorID, err, context, details)
}

// Info implements the errorid.Logger interface
func (l *ELKLogger) Info(msg string) {
	// Log to stdout
	fmt.Fprintln(os.Stdout, msg)
}

// sendStructuredError sends structured error data to ELK
func (l *ELKLogger) sendStructuredError(id string, err error, context string, details map[string]interface{}) {
	if l.elkURL == "" {
		return
	}

	// Prepare fully structured log entry
	logEntry := map[string]interface{}{
		"@timestamp":  time.Now().UTC().Format(time.RFC3339),
		"error_id":    id,
		"error_type":  "tracked",
		"context":     context,
		"error":       err.Error(),
		"service":     "go-support-id-example",
		"level":       "error",
		"environment": os.Getenv("ENVIRONMENT"),
	}

	// Add all details as separate fields for better filtering
	if len(details) > 0 {
		for key, value := range details {
			logEntry[key] = value
		}
	}

	jsonData, err := json.Marshal(logEntry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal structured error: %v\n", err)
		return
	}

	// Send to ELK
	req, err := http.NewRequest("POST", l.elkURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create ELK request: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	if username := os.Getenv("ELK_USERNAME"); username != "" {
		password := os.Getenv("ELK_PASSWORD")
		req.SetBasicAuth(username, password)
	}

	resp, err := l.httpClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send to ELK: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		fmt.Fprintf(os.Stderr, "ELK returned error status: %d\n", resp.StatusCode)
	}
}

// Ensure ELKLogger implements errorid.Logger interface
var _ errorid.Logger = (*ELKLogger)(nil)
