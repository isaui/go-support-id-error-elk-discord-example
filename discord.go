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

// DiscordWebhook handles sending notifications to Discord
type DiscordWebhook struct {
	webhookURL string
	httpClient *http.Client
}

// NewDiscordWebhook creates a new Discord webhook handler
func NewDiscordWebhook(webhookURL string) *DiscordWebhook {
	return &DiscordWebhook{
		webhookURL: webhookURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// DiscordMessage represents a Discord webhook message
type DiscordMessage struct {
	Content string   `json:"content,omitempty"`
	Embeds  []Embed  `json:"embeds,omitempty"`
}

// Embed represents a Discord embed
type Embed struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Color       int     `json:"color"`
	Fields      []Field `json:"fields,omitempty"`
	Timestamp   string  `json:"timestamp"`
}

// Field represents a Discord embed field
type Field struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

// SendErrorNotification sends error notification to Discord
func (d *DiscordWebhook) SendErrorNotification(err *errorid.ErrorWithID) {
	if d.webhookURL == "" {
		fmt.Println("Discord webhook URL not configured, skipping notification")
		return
	}

	// Build Discord embed with proper limits
	description := fmt.Sprintf("**Error:** %v\n**Context:** %s", err.Original, err.Context)
	if len(description) > 2048 {
		description = description[:2048]
	}

	embed := Embed{
		Title:       truncateString(fmt.Sprintf("Error: %s", err.ID), 256),
		Description: description,
		Color:       15158332, // Red color
		Fields:      []Field{},
	}

	// Add details (metadata) if available
	if err.Details != nil && len(err.Details) > 0 {
		detailsValue := formatDetails(err.Details)
		if detailsValue != "" && detailsValue != "None" {
			embed.Fields = append(embed.Fields, Field{
				Name:   "Details",
				Value:  truncateString(detailsValue, 1024),
				Inline: false,
			})
		}
	}

	// Add environment info
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		embed.Fields = append(embed.Fields, Field{
			Name:   "Environment",
			Value:  env,
			Inline: true,
		})
	}

	// Add stack trace if available (separate from details in v1.1.0+)
	if err.StackTrace != "" {
		stackTrace := err.StackTrace
		if len(stackTrace) > 900 {
			stackTrace = stackTrace[:900] + "..."
		}
		embed.Fields = append(embed.Fields, Field{
			Name:   "Stack Trace",
			Value:  "```\n" + stackTrace + "\n```",
			Inline: false,
		})
	}

	message := DiscordMessage{
		Embeds: []Embed{embed},
	}

	// Send to Discord
	go d.sendToDiscord(message)
}

// sendToDiscord sends message to Discord webhook
func (d *DiscordWebhook) sendToDiscord(message DiscordMessage) {
	jsonData, err := json.Marshal(message)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal Discord message: %v\n", err)
		return
	}

	req, err := http.NewRequest("POST", d.webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create Discord request: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send to Discord: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		fmt.Fprintf(os.Stderr, "Discord webhook returned error status: %d\n", resp.StatusCode)
	} else {
		fmt.Printf("Error notification sent to Discord: %s\n", message.Embeds[0].Title)
	}
}

// formatDetails converts Details map to readable string for Discord
func formatDetails(details map[string]interface{}) string {
	if len(details) == 0 {
		return "None"
	}

	var result string
	for key, value := range details {
		result += fmt.Sprintf("• **%s**: %v\n", key, value)
	}
	return result
}

// truncateString truncates string to max length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
