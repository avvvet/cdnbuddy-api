package cdn

import (
	"context"

	"github.com/avvvet/cdnbuddy-api/internal/domain"
)

// CDNProvider interface that all providers must implement
type CDNProvider interface {
	// Basic operations
	CreateService(ctx context.Context, config *ServiceConfig) (*domain.CDNService, error)
	ListServices(ctx context.Context) ([]domain.CDNService, error)
	UpdateService(ctx context.Context, serviceID string, config *ServiceConfig) error
	DeleteService(ctx context.Context, serviceID string) error

	// Domain management
	AddDomain(ctx context.Context, serviceID, domain string) error
	RemoveDomain(ctx context.Context, serviceID, domain string) error
	ListDomains(ctx context.Context, serviceID string) ([]domain.Domain, error)

	// Cache management
	PurgeCache(ctx context.Context, serviceID string, paths []string) error
	PurgeAll(ctx context.Context, serviceID string) error

	// Metrics
	GetMetrics(ctx context.Context, serviceID string) (*domain.Metrics, error)

	// Configuration
	UpdateCacheRules(ctx context.Context, serviceID string, rules []CacheRule) error
	UpdateOriginSettings(ctx context.Context, serviceID string, origin OriginConfig) error
}

type ServiceConfig struct {
	Name   string            `json:"name"`
	Origin OriginConfig      `json:"origin"`
	Rules  []CacheRule       `json:"rules"`
	SSL    SSLConfig         `json:"ssl"`
	Custom map[string]string `json:"custom"`
}

type OriginConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Path     string `json:"path"`
}

type CacheRule struct {
	Path        string `json:"path"`
	TTL         int    `json:"ttl"`         // seconds
	BrowserTTL  int    `json:"browser_ttl"` // seconds
	AlwaysCache bool   `json:"always_cache"`
}

type SSLConfig struct {
	Enabled     bool   `json:"enabled"`
	Certificate string `json:"certificate,omitempty"`
	PrivateKey  string `json:"private_key,omitempty"`
}
