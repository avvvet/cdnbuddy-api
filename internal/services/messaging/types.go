package messaging

import "time"

// Request/Response types for RPC-style communication

// Status Request/Response
type StatusRequest struct {
	UserID    string    `json:"user_id"`
	SessionID string    `json:"session_id"`
	Timestamp time.Time `json:"timestamp"`
}

type StatusResponse struct {
	Provider string   `json:"provider"`
	Domains  []Domain `json:"domains"`
	Metrics  Metrics  `json:"metrics"`
	Error    string   `json:"error,omitempty"`
}

type Domain struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Regions int    `json:"regions"`
}

type Metrics struct {
	CacheHitRatio   string `json:"cache_hit_ratio"`
	AvgResponseTime string `json:"avg_response_time"`
	TotalRequests   string `json:"total_requests"`
}

// Execution Plan types
type ExecutionPlan struct {
	ID    string     `json:"id"`
	Title string     `json:"title"`
	Steps []PlanStep `json:"steps"`
}

type PlanStep struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Details string `json:"details,omitempty"`
}

// Chat AI Request/Response
type ChatRequest struct {
	UserID    string    `json:"user_id"`
	SessionID string    `json:"session_id"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

type ChatResponse struct {
	Content string         `json:"content"`
	Plan    *ExecutionPlan `json:"plan,omitempty"`
	Error   string         `json:"error,omitempty"`
}

// CDN Operation Request types
type CDNOperationRequest struct {
	Type      string                 `json:"type"`
	ServiceID string                 `json:"service_id"`
	UserID    string                 `json:"user_id"`
	Params    map[string]interface{} `json:"params"`
	Timestamp time.Time              `json:"timestamp"`
}

type CDNOperationResponse struct {
	OperationID string                 `json:"operation_id"`
	Status      string                 `json:"status"`
	Result      map[string]interface{} `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// CDN Management Request types
type CreateServiceRequest struct {
	UserID   string      `json:"user_id"`
	Provider string      `json:"provider"`
	Name     string      `json:"name"`
	Config   interface{} `json:"config"`
}

type AddDomainRequest struct {
	ServiceID string `json:"service_id"`
	UserID    string `json:"user_id"`
	Domain    string `json:"domain"`
}

type PurgeCacheRequest struct {
	ServiceID string   `json:"service_id"`
	UserID    string   `json:"user_id"`
	Paths     []string `json:"paths,omitempty"`
}

type MetricsRequest struct {
	ServiceID string `json:"service_id"`
	UserID    string `json:"user_id"`
	Period    string `json:"period,omitempty"` // hour, day, week, month
}

// Socket Communication types (matches frontend types)
type SocketMessage struct {
	Type      string          `json:"type"`
	Content   string          `json:"content,omitempty"`
	Plan      *ExecutionPlan  `json:"plan,omitempty"`
	Status    *StatusResponse `json:"status,omitempty"`
	SessionID string          `json:"session_id"`
	Success   bool            `json:"success,omitempty"`
	Message   string          `json:"message,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}

// Error types
type ErrorEvent struct {
	Type      string    `json:"type"`
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	ServiceID string    `json:"service_id,omitempty"`
	UserID    string    `json:"user_id,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// Health Check types
type HealthCheckRequest struct {
	Service   string    `json:"service"`
	Timestamp time.Time `json:"timestamp"`
}

type HealthCheckResponse struct {
	Service   string            `json:"service"`
	Status    string            `json:"status"` // healthy, unhealthy, degraded
	Details   map[string]string `json:"details,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

// Batch operation types
type BatchOperationRequest struct {
	Operations []CDNOperationRequest `json:"operations"`
	UserID     string                `json:"user_id"`
	Timestamp  time.Time             `json:"timestamp"`
}

type BatchOperationResponse struct {
	BatchID   string                 `json:"batch_id"`
	Status    string                 `json:"status"`
	Results   []CDNOperationResponse `json:"results"`
	Error     string                 `json:"error,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// Configuration update types
type ConfigUpdateRequest struct {
	ServiceID string      `json:"service_id"`
	UserID    string      `json:"user_id"`
	Config    interface{} `json:"config"`
	Timestamp time.Time   `json:"timestamp"`
}

type ConfigUpdateResponse struct {
	ServiceID string    `json:"service_id"`
	Status    string    `json:"status"`
	Applied   bool      `json:"applied"`
	Error     string    `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// Analytics and reporting types
type AnalyticsRequest struct {
	ServiceID string    `json:"service_id"`
	UserID    string    `json:"user_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Metrics   []string  `json:"metrics"` // cache_hit_ratio, response_time, requests, bandwidth
}

type AnalyticsResponse struct {
	ServiceID string                 `json:"service_id"`
	Data      map[string]interface{} `json:"data"`
	Period    string                 `json:"period"`
	Error     string                 `json:"error,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// Notification types
type NotificationEvent struct {
	Type      string                 `json:"type"`
	UserID    string                 `json:"user_id"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Level     string                 `json:"level"` // info, warning, error, success
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// Provider-specific types
type ProviderEvent struct {
	Type      string                 `json:"type"`
	Provider  string                 `json:"provider"`
	ServiceID string                 `json:"service_id"`
	Event     string                 `json:"event"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// Audit log types
type AuditEvent struct {
	Type      string                 `json:"type"`
	UserID    string                 `json:"user_id"`
	ServiceID string                 `json:"service_id,omitempty"`
	Action    string                 `json:"action"`
	Resource  string                 `json:"resource"`
	Details   map[string]interface{} `json:"details,omitempty"`
	IPAddress string                 `json:"ip_address,omitempty"`
	UserAgent string                 `json:"user_agent,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}
