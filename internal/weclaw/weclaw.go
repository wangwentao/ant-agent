package weclaw

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// StreamEvent represents a single event for stream-json output
type StreamEvent struct {
	Type      string `json:"type"`
	SessionID string `json:"session_id"`
	Result    string `json:"result"`
	IsError   bool   `json:"is_error"`
}

// Config holds WeClaw-specific configuration
type Config struct {
	OutputFormat    string
	ResumeSessionID string
	SystemPrompt    string
	SessionID       string
}

// GenerateSessionID generates a unique session ID if none is provided
func GenerateSessionID(resumeSessionID string) string {
	if resumeSessionID != "" {
		return resumeSessionID
	}
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}

// CreateSessionEvent creates a session event
func CreateSessionEvent(sessionID string) string {
	event := StreamEvent{
		Type:      "session",
		SessionID: sessionID,
	}
	jsonEvent, _ := json.Marshal(event)
	return string(jsonEvent)
}

// CreateResultEvent creates a result event
func CreateResultEvent(sessionID, result string, isError bool) string {
	event := StreamEvent{
		Type:      "result",
		SessionID: sessionID,
		Result:    result,
		IsError:   isError,
	}
	jsonEvent, _ := json.Marshal(event)
	return string(jsonEvent)
}

// OutputResultEvent outputs a result event
func OutputResultEvent(sessionID, result string, isError bool) {
	event := CreateResultEvent(sessionID, result, isError)
	fmt.Println(event)
	os.Stdout.Sync()
}

// ShouldUseStreamJSON checks if output format is stream-json
func ShouldUseStreamJSON(outputFormat string) bool {
	return outputFormat == "stream-json"
}
