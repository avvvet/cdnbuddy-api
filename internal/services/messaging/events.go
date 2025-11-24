package messaging

import "time"

// NATS Subjects
const (
	SubjectCDNService = "cdnbuddy.cdn.service"
	SubjectDomain     = "cdnbuddy.cdn.domain"
	SubjectCache      = "cdnbuddy.cdn.cache"
	SubjectMetrics    = "cdnbuddy.cdn.metrics"
	SubjectOperation  = "cdnbuddy.operation"
	SubjectChat       = "cdnbuddy.chat"

	SubjectExecutionPlan  = "cdnbuddy.execution_plan"
	SubjectStatusRequest  = "cdnbuddy.status.request"
	SubjectStatusResponse = "cdnbuddy.status.response"

	SubjectChatResponse = "cdnbuddy.chat.response" // For AI responses
	SubjectNotification = "cdnbuddy.notification"  // For notifications

)

// Event Types
const (
	// CDN Service Events
	EventCDNServiceCreated = "cdn_service.created"
	EventCDNServiceUpdated = "cdn_service.updated"
	EventCDNServiceDeleted = "cdn_service.deleted"

	// Domain Events
	EventDomainAdded         = "domain.added"
	EventDomainRemoved       = "domain.removed"
	EventDomainStatusChanged = "domain.status_changed"

	// Cache Events
	EventCachePurged       = "cache.purged"
	EventCacheRulesUpdated = "cache.rules_updated"

	// Metrics Events
	EventMetricsUpdated = "metrics.updated"

	// Operation Events
	EventOperationStarted   = "operation.started"
	EventOperationProgress  = "operation.progress"
	EventOperationCompleted = "operation.completed"
	EventOperationFailed    = "operation.failed"

	// Chat Events
	EventChatMessage = "chat.message"
	EventAIResponse  = "chat.ai_response"

	// Execution Plan Events
	EventExecutionPlan = "execution_plan.created"
)

// CDN Service Events
type CDNServiceEvent struct {
	Type      string    `json:"type"`
	ServiceID string    `json:"service_id"`
	UserID    string    `json:"user_id"`
	Provider  string    `json:"provider"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Config    string    `json:"config,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// Domain Events
type DomainEvent struct {
	Type         string    `json:"type"`
	DomainID     string    `json:"domain_id"`
	CDNServiceID string    `json:"cdn_service_id"`
	Name         string    `json:"name"`
	Status       string    `json:"status"`
	OldStatus    string    `json:"old_status,omitempty"`
	Regions      int       `json:"regions"`
	Timestamp    time.Time `json:"timestamp"`
}

// Cache Events
type CacheEvent struct {
	Type      string      `json:"type"`
	ServiceID string      `json:"service_id"`
	UserID    string      `json:"user_id"`
	Paths     []string    `json:"paths,omitempty"`
	Rules     interface{} `json:"rules,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// Metrics Events
type MetricsEvent struct {
	Type            string    `json:"type"`
	ServiceID       string    `json:"service_id"`
	CacheHitRatio   float64   `json:"cache_hit_ratio"`
	AvgResponseTime int       `json:"avg_response_time"`
	TotalRequests   int64     `json:"total_requests"`
	Timestamp       time.Time `json:"timestamp"`
}

// Operation Events
type OperationEvent struct {
	Type        string                 `json:"type"`
	OperationID string                 `json:"operation_id"`
	ServiceID   string                 `json:"service_id"`
	UserID      string                 `json:"user_id"`
	OpType      string                 `json:"op_type"`
	Status      string                 `json:"status"`
	Progress    string                 `json:"progress,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Params      map[string]interface{} `json:"params,omitempty"`
	Result      map[string]interface{} `json:"result,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// Chat Events
type ChatEvent struct {
	Type      string    `json:"type"`
	UserID    string    `json:"user_id"`
	SessionID string    `json:"session_id"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// Execution Plan Events
type ExecutionPlanEvent struct {
	Type      string      `json:"type"`
	UserID    string      `json:"user_id"`
	SessionID string      `json:"session_id"`
	Plan      interface{} `json:"plan"`
	Timestamp time.Time   `json:"timestamp"`
}

// StatusRequestEvent is received from Socket Server
type StatusRequestEvent struct {
	UserID    string    `json:"user_id"`
	SessionID string    `json:"session_id"`
	Timestamp time.Time `json:"timestamp"`
}

// StatusResponseEvent is sent back to Socket Server
type StatusResponseEvent struct {
	UserID    string          `json:"user_id"`
	SessionID string          `json:"session_id"`
	Services  []ServiceStatus `json:"services"`
	Timestamp time.Time       `json:"timestamp"`
}

// ServiceStatus represents a CDN service status
type ServiceStatus struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	TestURL  string `json:"test_url"`
	Provider string `json:"provider"`
}
