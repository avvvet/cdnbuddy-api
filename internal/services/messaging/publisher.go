package messaging

import (
	"context"
	"time"

	"github.com/avvvet/cdnbuddy-api/internal/domain"
	"github.com/sirupsen/logrus"
)

type Publisher struct {
	client *NATSClient
}

func NewPublisher(client *NATSClient) *Publisher {
	return &Publisher{client: client}
}

// CDN Service Events
func (p *Publisher) PublishCDNServiceCreated(service *domain.CDNService) error {
	event := CDNServiceEvent{
		Type:      EventCDNServiceCreated,
		ServiceID: service.ID,
		UserID:    service.UserID,
		Provider:  string(service.Provider),
		Name:      service.Name,
		Status:    service.Status,
		Timestamp: time.Now(),
	}

	return p.client.Publish(SubjectCDNService, event)
}

func (p *Publisher) PublishCDNServiceUpdated(service *domain.CDNService) error {
	event := CDNServiceEvent{
		Type:      EventCDNServiceUpdated,
		ServiceID: service.ID,
		UserID:    service.UserID,
		Provider:  string(service.Provider),
		Name:      service.Name,
		Status:    service.Status,
		Timestamp: time.Now(),
	}

	return p.client.Publish(SubjectCDNService, event)
}

func (p *Publisher) PublishCDNServiceDeleted(serviceID, userID string) error {
	event := CDNServiceEvent{
		Type:      EventCDNServiceDeleted,
		ServiceID: serviceID,
		UserID:    userID,
		Timestamp: time.Now(),
	}

	return p.client.Publish(SubjectCDNService, event)
}

// Domain Events
func (p *Publisher) PublishDomainAdded(domain *domain.Domain) error {
	event := DomainEvent{
		Type:         EventDomainAdded,
		DomainID:     domain.ID,
		CDNServiceID: domain.CDNServiceID,
		Name:         domain.Name,
		Status:       domain.Status,
		Regions:      domain.Regions,
		Timestamp:    time.Now(),
	}

	return p.client.Publish(SubjectDomain, event)
}

func (p *Publisher) PublishDomainRemoved(domain *domain.Domain) error {
	event := DomainEvent{
		Type:         EventDomainRemoved,
		DomainID:     domain.ID,
		CDNServiceID: domain.CDNServiceID,
		Name:         domain.Name,
		Status:       domain.Status,
		Regions:      domain.Regions,
		Timestamp:    time.Now(),
	}

	return p.client.Publish(SubjectDomain, event)
}

func (p *Publisher) PublishDomainStatusChanged(domain *domain.Domain, oldStatus string) error {
	event := DomainEvent{
		Type:         EventDomainStatusChanged,
		DomainID:     domain.ID,
		CDNServiceID: domain.CDNServiceID,
		Name:         domain.Name,
		Status:       domain.Status,
		OldStatus:    oldStatus,
		Regions:      domain.Regions,
		Timestamp:    time.Now(),
	}

	return p.client.Publish(SubjectDomain, event)
}

// Cache Events
func (p *Publisher) PublishCachePurged(serviceID, userID string, paths []string) error {
	event := CacheEvent{
		Type:      EventCachePurged,
		ServiceID: serviceID,
		UserID:    userID,
		Paths:     paths,
		Timestamp: time.Now(),
	}

	return p.client.Publish(SubjectCache, event)
}

func (p *Publisher) PublishCacheRulesUpdated(serviceID, userID string, rules interface{}) error {
	event := CacheEvent{
		Type:      EventCacheRulesUpdated,
		ServiceID: serviceID,
		UserID:    userID,
		Rules:     rules,
		Timestamp: time.Now(),
	}

	return p.client.Publish(SubjectCache, event)
}

// Metrics Events
func (p *Publisher) PublishMetricsUpdated(metrics *domain.Metrics) error {
	event := MetricsEvent{
		Type:            EventMetricsUpdated,
		ServiceID:       metrics.CDNServiceID,
		CacheHitRatio:   metrics.CacheHitRatio,
		AvgResponseTime: metrics.AvgResponseTime,
		TotalRequests:   metrics.TotalRequests,
		Timestamp:       time.Now(),
	}

	return p.client.Publish(SubjectMetrics, event)
}

// Operation Events (for execution plans)
func (p *Publisher) PublishOperationStarted(operation *domain.CDNOperation) error {
	event := OperationEvent{
		Type:        EventOperationStarted,
		OperationID: operation.ID,
		ServiceID:   getServiceIDFromOperation(operation),
		UserID:      getUserIDFromOperation(operation),
		OpType:      operation.Type,
		Status:      operation.Status,
		Params:      operation.Params,
		Timestamp:   time.Now(),
	}

	return p.client.Publish(SubjectOperation, event)
}

func (p *Publisher) PublishOperationProgress(operation *domain.CDNOperation, progress string) error {
	event := OperationEvent{
		Type:        EventOperationProgress,
		OperationID: operation.ID,
		ServiceID:   getServiceIDFromOperation(operation),
		UserID:      getUserIDFromOperation(operation),
		OpType:      operation.Type,
		Status:      operation.Status,
		Progress:    progress,
		Params:      operation.Params,
		Timestamp:   time.Now(),
	}

	return p.client.Publish(SubjectOperation, event)
}

func (p *Publisher) PublishOperationCompleted(operation *domain.CDNOperation) error {
	event := OperationEvent{
		Type:        EventOperationCompleted,
		OperationID: operation.ID,
		ServiceID:   getServiceIDFromOperation(operation),
		UserID:      getUserIDFromOperation(operation),
		OpType:      operation.Type,
		Status:      operation.Status,
		Params:      operation.Params,
		Result:      operation.Result,
		Timestamp:   time.Now(),
	}

	return p.client.Publish(SubjectOperation, event)
}

func (p *Publisher) PublishOperationFailed(operation *domain.CDNOperation, errorMsg string) error {
	event := OperationEvent{
		Type:        EventOperationFailed,
		OperationID: operation.ID,
		ServiceID:   getServiceIDFromOperation(operation),
		UserID:      getUserIDFromOperation(operation),
		OpType:      operation.Type,
		Status:      "failed",
		Error:       errorMsg,
		Params:      operation.Params,
		Timestamp:   time.Now(),
	}

	return p.client.Publish(SubjectOperation, event)
}

// Chat Events (for socket service integration)
func (p *Publisher) PublishChatMessage(userID, sessionID, message string) error {
	event := ChatEvent{
		Type:      EventChatMessage,
		UserID:    userID,
		SessionID: sessionID,
		Message:   message,
		Timestamp: time.Now(),
	}

	return p.client.Publish(SubjectChat, event)
}

func (p *Publisher) PublishAIResponse(userID, sessionID, response string) error {
	event := ChatEvent{
		Type:      EventAIResponse,
		UserID:    userID,
		SessionID: sessionID,
		Message:   response,
		Timestamp: time.Now(),
	}

	return p.client.Publish(SubjectChatResponse, event)
}

// Remove manual marshaling, let client.Publish handle it
func (p *Publisher) PublishExecutionPlan(ctx context.Context, event ExecutionPlanEvent) error {
	subject := "cdnbuddy.execution.plan"
	logrus.WithFields(logrus.Fields{
		"subject": subject,
		"plan_id": event.Plan.ID,
		"user_id": event.UserID,
	}).Info("📤 Publishing execution plan")

	return p.client.Publish(subject, event) // Pass event, not data
}

// PublishStatusResponse sends CDN status back to Socket Server
func (p *Publisher) PublishStatusResponse(userID, sessionID string, services []ServiceStatus) error {
	event := StatusResponseEvent{
		UserID:    userID,
		SessionID: sessionID,
		Services:  services,
		Timestamp: time.Now(),
	}

	return p.client.Publish(SubjectStatusResponse, event)
}

// Helper functions to extract IDs from operation params
func getServiceIDFromOperation(op *domain.CDNOperation) string {
	if serviceID, ok := op.Params["service_id"].(string); ok {
		return serviceID
	}
	return ""
}

func getUserIDFromOperation(op *domain.CDNOperation) string {
	if userID, ok := op.Params["user_id"].(string); ok {
		return userID
	}
	return ""
}
