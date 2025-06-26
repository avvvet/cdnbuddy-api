package domain

import (
	"time"
)

type CDNProvider string

const (
	ProviderCacheFly   CDNProvider = "cachefly"
	ProviderCloudflare CDNProvider = "cloudflare"
)

type CDNService struct {
	ID        string      `json:"id" db:"id"`
	UserID    string      `json:"user_id" db:"user_id"`
	Provider  CDNProvider `json:"provider" db:"provider"`
	Name      string      `json:"name" db:"name"`
	Status    string      `json:"status" db:"status"`
	Config    string      `json:"config" db:"config"` // JSON config
	CreatedAt time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt time.Time   `json:"updated_at" db:"updated_at"`
}

type Domain struct {
	ID           string    `json:"id" db:"id"`
	CDNServiceID string    `json:"cdn_service_id" db:"cdn_service_id"`
	Name         string    `json:"name" db:"name"`
	Status       string    `json:"status" db:"status"`
	Regions      int       `json:"regions" db:"regions"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type Metrics struct {
	ID              string    `json:"id" db:"id"`
	CDNServiceID    string    `json:"cdn_service_id" db:"cdn_service_id"`
	CacheHitRatio   float64   `json:"cache_hit_ratio" db:"cache_hit_ratio"`
	AvgResponseTime int       `json:"avg_response_time" db:"avg_response_time"` // milliseconds
	TotalRequests   int64     `json:"total_requests" db:"total_requests"`
	Timestamp       time.Time `json:"timestamp" db:"timestamp"`
}

// CDN Management Operations
type CDNOperation struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Status    string                 `json:"status"`
	Params    map[string]interface{} `json:"params"`
	Result    map[string]interface{} `json:"result,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}
