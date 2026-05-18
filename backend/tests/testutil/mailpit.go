package testutil

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"time"
)

const (
	mailpitAPI = "http://localhost:8025/api/v1"
)

// MailpitMessage represents a message in Mailpit
type MailpitMessage struct {
	ID       string `json:"ID"`
	From     string `json:"From"`
	To       string `json:"To"`
	Subject  string `json:"Subject"`
	Body     string `json:"Body"`
	HTMLBody string `json:"HTMLBody"`
	Date     string `json:"Date"`
	Snippet  string `json:"Snippet"`
}

// MailpitRecipient represents a recipient in Mailpit API v1
type MailpitRecipient struct {
	Name    string `json:"Name"`
	Address string `json:"Address"`
}

// MailpitMessageV1 represents a message in Mailpit API v1
type MailpitMessageV1 struct {
	ID       string             `json:"ID"`
	From     MailpitRecipient   `json:"From"`
	To       []MailpitRecipient `json:"To"`
	Subject  string             `json:"Subject"`
	Body     string             `json:"Body"`
	HTMLBody string             `json:"HTMLBody"`
	Date     string             `json:"Date"`
	Snippet  string             `json:"Snippet"`
}

// ClearMessages deletes all messages from Mailpit
func ClearMessages() error {
	req, err := http.NewRequest("DELETE", mailpitAPI+"/deleteall", nil)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to clear messages: status %d", resp.StatusCode)
	}
	return nil
}

// GetMessages returns all messages from Mailpit
func GetMessages() ([]MailpitMessage, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(mailpitAPI + "/messages")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get messages: status %d", resp.StatusCode)
	}

	var result struct {
		Messages []MailpitMessageV1 `json:"messages"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Convert V1 format to our format
	messages := make([]MailpitMessage, len(result.Messages))
	for i, m := range result.Messages {
		toAddr := ""
		if len(m.To) > 0 {
			toAddr = m.To[0].Address
		}
		messages[i] = MailpitMessage{
			ID:       m.ID,
			From:     m.From.Address,
			To:       toAddr,
			Subject:  m.Subject,
			Body:     m.Body,
			HTMLBody: m.HTMLBody,
			Date:     m.Date,
			Snippet:  m.Snippet,
		}
	}
	return messages, nil
}

// GetMessage retrieves a single message by ID
func GetMessage(id string) (*MailpitMessage, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(mailpitAPI + "/message/" + id)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get message: status %d", resp.StatusCode)
	}

	var msg MailpitMessage
	if err := json.NewDecoder(resp.Body).Decode(&msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// WaitForMessage waits for a message to arrive for a specific recipient
func WaitForMessage(to string, timeout time.Duration) (*MailpitMessage, error) {
	deadline := time.Now().Add(timeout)
	pollInterval := 500 * time.Millisecond

	for time.Now().Before(deadline) {
		messages, err := GetMessages()
		if err != nil {
			time.Sleep(pollInterval)
			continue
		}

		for _, msg := range messages {
			// msg.To is an array of email objects, check if any match
			if msg.To == to {
				return &msg, nil
			}
		}
		time.Sleep(pollInterval)
	}
	return nil, fmt.Errorf("timeout waiting for message to %s", to)
}

// ExtractVerificationToken extracts token from email verification message
func ExtractVerificationToken(msg *MailpitMessage) string {
	// Try snippet first (usually contains the full link/text)
	token := extractTokenFromString(msg.Snippet)
	if token != "" {
		return token
	}
	// Try plain text body
	token = extractTokenFromString(msg.Body)
	if token != "" {
		return token
	}
	// Try HTML body
	return extractTokenFromString(msg.HTMLBody)
}

// ExtractPasswordResetToken extracts token from password reset message
func ExtractPasswordResetToken(msg *MailpitMessage) string {
	// Try plain text body first
	token := extractTokenFromString(msg.Body)
	if token != "" {
		return token
	}
	// Try HTML body
	return extractTokenFromString(msg.HTMLBody)
}

// extractTokenFromString uses regex to find verification/reset tokens in text
func extractTokenFromString(text string) string {
	// Pattern for verification tokens (UUID format)
	uuidPattern := regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}`)
	if matches := uuidPattern.FindString(text); matches != "" {
		return matches
	}

	// Pattern for short numeric tokens (6-8 digits)
	numPattern := regexp.MustCompile(`\b\d{6,8}\b`)
	if matches := numPattern.FindString(text); matches != "" {
		return matches
	}

	// Pattern for URL-safe tokens (32+ chars)
	urlTokenPattern := regexp.MustCompile(`[A-Za-z0-9]{32,}`)
	if matches := urlTokenPattern.FindString(text); matches != "" {
		return matches
	}

	return ""
}
