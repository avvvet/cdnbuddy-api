package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Client provides high-level messaging operations
type Client struct {
	nats       *NATSClient
	publisher  *Publisher
	subscriber *Subscriber
}

func NewClient(natsURL string) (*Client, error) {
	natsClient, err := NewNATSClient(natsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create NATS client: %w", err)
	}

	return &Client{
		nats:       natsClient,
		publisher:  NewPublisher(natsClient),
		subscriber: NewSubscriber(natsClient),
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
