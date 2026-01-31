package models

import (
	"fmt"
	"time"
)

// NATS Request from backend
type IntentRequest struct {
	SessionID           string                `json:"session_id"`
	UserMessage         string                `json:"user_message"`
	ConversationHistory []ConversationMessage `json:"conversation_history"`
	AvailableActions    []ActionSchema        `json:"available_actions"`
}

type ConversationMessage struct {
	Role      string    `json:"role"` // "user" or "assistant"
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

type ActionSchema struct {
	Action     string   `json:"action"`
	Parameters []string `json:"parameters"`
}

// NATS Response to backend
type IntentResponse struct {
	SessionID    string             `json:"session_id"`
	Action       *string            `json:"action"`
	Status       string             `json:"status"` // "NEEDS_INFO", "READY", "ERROR"
	Parameters   map[string]*string `json:"parameters"`
	UserMessage  string             `json:"user_message"`
	ErrorCode    *string            `json:"error_code,omitempty"`
	ErrorMessage *string            `json:"error_message,omitempty"`
}

// ExecutionPlan represents a pending execution plan for the user
type ExecutionPlan struct {
	ID                string             `json:"id"`
	Title             string             `json:"title"`
	Description       string             `json:"description"`
	Steps             []string           `json:"steps"`
	EstimatedDuration string             `json:"estimated_duration"`
	Action            string             `json:"action"`
	Parameters        map[string]*string `json:"parameters"`
	IntentResponse    *IntentResponse    `json:"-"` // Store original intent (not sent to frontend)
	CreatedAt         time.Time          `json:"created_at"`
	ExpiresAt         time.Time          `json:"expires_at"`
}

// BuildExecutionPlan creates an execution plan from IntentResponse
func BuildExecutionPlan(intent *IntentResponse) ExecutionPlan {
	plan := ExecutionPlan{
		ID:                generatePlanID(),
		CreatedAt:         time.Now(),
		ExpiresAt:         time.Now().Add(5 * time.Minute),
		Parameters:        intent.Parameters,
		IntentResponse:    intent,
		EstimatedDuration: "30 seconds",
	}

	// Set action
	if intent.Action != nil {
		plan.Action = *intent.Action
	}

	// Build user-friendly steps based on action
	if intent.Action == nil {
		return plan
	}

	switch *intent.Action {
	case "SETUP_CDN":
		domain := ""
		origin := ""
		if d := intent.Parameters["domain"]; d != nil {
			domain = *d
		}
		if o := intent.Parameters["origin"]; o != nil {
			origin = *o
		}

		plan.Title = fmt.Sprintf("Setup CDN for %s", domain)
		plan.Description = "Create and configure CDN service"
		plan.Steps = []string{
			fmt.Sprintf("Create CDN service for %s", domain),
			fmt.Sprintf("Configure origin: %s", origin),
			"Enable SSL certificate",
			"Configure caching rules",
		}

	case "PURGE_CACHE":
		domain := ""
		if d := intent.Parameters["domain"]; d != nil {
			domain = *d
		}
		plan.Title = fmt.Sprintf("Purge cache for %s", domain)
		plan.Description = "Clear CDN cache"
		plan.Steps = []string{
			fmt.Sprintf("Clear cache for %s", domain),
			"Propagate changes across CDN nodes",
		}

	default:
		plan.Title = "Execute action"
		plan.Description = "Process your request"
		plan.Steps = []string{"Execute requested action"}
	}

	return plan
}

// generatePlanID creates a unique plan ID
func generatePlanID() string {
	return fmt.Sprintf("plan_%d", time.Now().UnixNano())
}
