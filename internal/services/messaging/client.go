package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/avvvet/cdnbuddy-api/internal/models"
)

// Client provides high-level messaging operations
type Client struct {
	nats       *NATSClient
	publisher  *Publisher
	subscriber *Subscriber

	sessionHistory map[string][]models.ConversationMessage // Add this
	mu             sync.RWMutex                            // Add mutex for concurrent access
}

func NewClient(natsURL string) (*Client, error) {
	natsClient, err := NewNATSClient(natsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create NATS client: %w", err)
	}

	return &Client{
		nats:           natsClient,
		publisher:      NewPublisher(natsClient),
		subscriber:     NewSubscriber(natsClient),
		sessionHistory: make(map[string][]models.ConversationMessage),
	}, nil
}

func (c *Client) Close() {
	c.nats.Close()
}

func (c *Client) Publisher() *Publisher {
	return c.publisher
}

func (c *Client) Subscriber() *Subscriber {
	return c.subscriber
}

// Request CDN status from socket service
func (c *Client) RequestCDNStatus(ctx context.Context, userID, sessionID string) (*StatusResponse, error) {
	request := StatusRequest{
		UserID:    userID,
		SessionID: sessionID,
		Timestamp: time.Now(),
	}

	msg, err := c.nats.Request(SubjectStatusRequest, request, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to request status: %w", err)
	}

	var response StatusResponse
	if err := json.Unmarshal(msg.Data, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// RequestIntentAnalysis sends a chat message to the intent service for analysis
func (c *Client) RequestIntentAnalysis(ctx context.Context, sessionID, userMessage string) (*models.IntentResponse, error) {
	// Get existing conversation history for this session
	conversationHistory := c.getSessionHistory(sessionID)

	// Append the current user message to session
	c.AppendToSession(sessionID, "user", userMessage)

	// Prepare request with complete conversation history
	request := models.IntentRequest{
		SessionID:           sessionID,
		UserMessage:         userMessage,
		ConversationHistory: conversationHistory,
		AvailableActions:    []models.ActionSchema{}, // Empty for now
	}

	// Send request to intent service
	msg, err := c.nats.Request("intent.analyze", request, 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to request intent analysis: %w", err)
	}

	// Parse and return response
	var response models.IntentResponse
	if err := json.Unmarshal(msg.Data, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal intent response: %w", err)
	}

	return &response, nil
}

// Send execution plan to socket service
func (c *Client) SendExecutionPlan(ctx context.Context, userID, sessionID string, plan interface{}) error {
	return c.publisher.PublishExecutionPlan(userID, sessionID, plan)
}

// Send AI response to socket service
func (c *Client) SendAIResponse(ctx context.Context, userID, sessionID, response string) error {
	return c.publisher.PublishAIResponse(userID, sessionID, response)
}

// Health check
func (c *Client) IsHealthy() bool {
	return c.nats.IsConnected()
}

// Get connection stats
func (c *Client) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"connected":   c.nats.IsConnected(),
		"server_info": c.nats.conn.ConnectedServerName(),
		"url":         c.nats.conn.ConnectedUrl(),
	}
}

// Session management methods

// getSessionHistory retrieves conversation history for a session
func (c *Client) getSessionHistory(sessionID string) []models.ConversationMessage {
	c.mu.RLock()
	defer c.mu.RUnlock()

	history, exists := c.sessionHistory[sessionID]
	if !exists {
		return []models.ConversationMessage{}
	}

	// Return a copy to avoid external modifications
	historyCopy := make([]models.ConversationMessage, len(history))
	copy(historyCopy, history)
	return historyCopy
}

// appendToSession adds a message to the session history
func (c *Client) AppendToSession(sessionID, role, message string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Initialize session if it doesn't exist
	if _, exists := c.sessionHistory[sessionID]; !exists {
		c.sessionHistory[sessionID] = []models.ConversationMessage{}
	}

	// Append new message
	c.sessionHistory[sessionID] = append(c.sessionHistory[sessionID], models.ConversationMessage{
		Role:      role,
		Message:   message,
		Timestamp: time.Now(),
	})
}

// clearSession removes a session from history (called when conversation is complete)
func (c *Client) clearSession(sessionID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.sessionHistory, sessionID)
}

// getSessionCount returns the number of active sessions (useful for monitoring)
func (c *Client) getSessionCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.sessionHistory)
}

// getAllSessionIDs returns all active session IDs (useful for cleanup/debugging)
func (c *Client) getAllSessionIDs() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	sessionIDs := make([]string, 0, len(c.sessionHistory))
	for sessionID := range c.sessionHistory {
		sessionIDs = append(sessionIDs, sessionID)
	}
	return sessionIDs
}
