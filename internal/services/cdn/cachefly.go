package cdn

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/avvvet/cdnbuddy-api/internal/domain"
	"github.com/cachefly/cachefly-go-sdk/pkg/cachefly"
	api "github.com/cachefly/cachefly-go-sdk/pkg/cachefly/api/v2_5"
	"github.com/google/uuid"
)

// CacheFlyProvider implements CDNProvider interface for CacheFly
type CacheFlyProvider struct {
	client   *cachefly.Client
	apiToken string
}

// NewCacheFlyProvider creates a new CacheFly provider
func NewCacheFlyProvider() (*CacheFlyProvider, error) {
	// Get API token from environment
	token := os.Getenv("CACHEFLY_API_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("CACHEFLY_API_TOKEN environment variable is required")
	}

	// Initialize CacheFly client
	client := cachefly.NewClient(
		cachefly.WithToken(token),
	)

	return &CacheFlyProvider{
		client:   client,
		apiToken: token,
	}, nil
}

// CreateService creates a new CDN service with origin configuration
func (p *CacheFlyProvider) CreateService(ctx context.Context, config *ServiceConfig) (*domain.CDNService, error) {
	// Generate service name from config or auto-generate
	serviceName := generateServiceName(config.Name)
	uniqueName := fmt.Sprintf("%s-%s", serviceName, uuid.New().String()[:8])

	// Step 1: Create CacheFly service
	createReq := api.CreateServiceRequest{
		Name:        serviceName,
		UniqueName:  uniqueName,
		Description: "CDN service created by CDNBuddy",
	}

	service, err := p.client.Services.Create(ctx, createReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create CacheFly service: %w", err)
	}

	// Step 2: Configure service options (including origin via reverseProxy)
	if err := p.configureServiceOptions(ctx, service.ID, config); err != nil {
		// Cleanup: try to deactivate the service if options fail
		_, err = p.client.Services.DeactivateServiceByID(ctx, service.ID)
		return nil, fmt.Errorf("failed to configure service options: %w", err)
	}

	// Step 3: Build and return domain.CDNService
	cdnService := &domain.CDNService{
		ID:       service.ID,
		Provider: domain.ProviderCacheFly,
		Name:     service.Name,
		Status:   service.Status,
		Config:   p.buildConfigJSON(service, config),
	}

	return cdnService, nil
}

// configureServiceOptions configures origin and performance settings with best practices
func (p *CacheFlyProvider) configureServiceOptions(ctx context.Context, serviceID string, config *ServiceConfig) error {
	// Determine origin scheme
	originScheme := "HTTPS"
	if config.Origin.Protocol != "" {
		originScheme = strings.ToUpper(config.Origin.Protocol)
	}

	// Get best practices configuration with origin details
	options := GetBestPracticesOptions(config.Name, config.Origin.Host, originScheme)

	// Add custom cache rules if provided (override defaults)
	if len(config.Rules) > 0 {
		options["expiryHeaders"] = p.buildExpiryHeaders(config.Rules)
	}

	// Update service options
	_, err := p.client.ServiceOptions.UpdateOptions(ctx, serviceID, options)
	if err != nil {
		return fmt.Errorf("failed to update service options: %w", err)
	}

	return nil
}

// buildExpiryHeaders converts cache rules to CacheFly expiry headers format
func (p *CacheFlyProvider) buildExpiryHeaders(rules []CacheRule) []interface{} {
	headers := make([]interface{}, 0, len(rules))

	for _, rule := range rules {
		header := map[string]interface{}{
			"path":       rule.Path,
			"expiryTime": rule.TTL,
		}
		headers = append(headers, header)
	}

	return headers
}

// AddDomain adds a custom domain to the service
func (p *CacheFlyProvider) AddDomain(ctx context.Context, serviceID, domainName string) error {
	req := api.CreateServiceDomainRequest{
		Name:        domainName,
		Description: fmt.Sprintf("Domain added by CDNBuddy for %s", domainName),
	}

	_, err := p.client.ServiceDomains.Create(ctx, serviceID, req)
	if err != nil {
		return fmt.Errorf("failed to add domain %s: %w", domainName, err)
	}

	return nil
}

// UpdateService updates service configuration
func (p *CacheFlyProvider) UpdateService(ctx context.Context, serviceID string, config *ServiceConfig) error {
	// Update service options with new configuration
	if err := p.configureServiceOptions(ctx, serviceID, config); err != nil {
		return fmt.Errorf("failed to update service: %w", err)
	}

	return nil
}

// DeleteService deactivates a CDN service (CacheFly doesn't support deletion)
func (p *CacheFlyProvider) DeleteService(ctx context.Context, serviceID string) error {
	_, err := p.client.Services.DeactivateServiceByID(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("failed to deactivate service: %w", err)
	}

	return nil
}

// ListServices lists all CDN services for the account
func (p *CacheFlyProvider) ListServices(ctx context.Context) ([]domain.CDNService, error) {
	opts := api.ListOptions{
		Offset:          0,
		Limit:           100, // Adjust as needed
		Status:          "ACTIVE",
		IncludeFeatures: false,
		ResponseType:    "",
	}

	resp, err := p.client.Services.List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	// Convert CacheFly services to domain.CDNService
	services := make([]domain.CDNService, 0, len(resp.Services))
	for _, svc := range resp.Services {
		// Build config JSON for each service
		configData := map[string]interface{}{
			"cachefly_service_id": svc.ID,
			"unique_name":         svc.UniqueName,
			"test_url":            fmt.Sprintf("https://%s.cachefly.net", svc.UniqueName),
			"auto_ssl":            svc.AutoSSL,
			"status":              svc.Status,
			"configuration_mode":  svc.ConfigurationMode,
		}
		configJSON, _ := json.Marshal(configData)

		services = append(services, domain.CDNService{
			ID:       svc.ID,
			Provider: domain.ProviderCacheFly,
			Name:     svc.Name,
			Status:   svc.Status,
			Config:   string(configJSON),
			// UserID and timestamps would be filled from database
		})
	}

	return services, nil
}

// RemoveDomain removes a domain from the service
func (p *CacheFlyProvider) RemoveDomain(ctx context.Context, serviceID, domainName string) error {
	// List domains to find the one to delete
	opts := api.ListServiceDomainsOptions{
		Offset: 0,
		Limit:  100,
	}

	resp, err := p.client.ServiceDomains.List(ctx, serviceID, opts)
	if err != nil {
		return fmt.Errorf("failed to list domains: %w", err)
	}

	// Find domain ID by name
	var domainID string
	for _, d := range resp.Domains {
		if d.Name == domainName {
			domainID = d.ID
			break
		}
	}

	if domainID == "" {
		return fmt.Errorf("domain %s not found", domainName)
	}

	// Delete the domain by ID
	err = p.client.ServiceDomains.DeleteByID(ctx, serviceID, domainID)
	if err != nil {
		return fmt.Errorf("failed to remove domain: %w", err)
	}

	return nil
}

// ListDomains lists all domains for a service
func (p *CacheFlyProvider) ListDomains(ctx context.Context, serviceID string) ([]domain.Domain, error) {
	opts := api.ListServiceDomainsOptions{
		Offset: 0,
		Limit:  100,
	}

	resp, err := p.client.ServiceDomains.List(ctx, serviceID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list domains: %w", err)
	}

	// Convert CacheFly domains to our domain type
	domains := make([]domain.Domain, 0, len(resp.Domains))
	for _, d := range resp.Domains {
		domains = append(domains, domain.Domain{
			ID:           d.ID,
			CDNServiceID: serviceID,
			Name:         d.Name,
			Status:       d.ValidationStatus,
			// Regions: not available in CacheFly API
		})
	}

	return domains, nil
}

// PurgeCache purges cache for specific paths
func (p *CacheFlyProvider) PurgeCache(ctx context.Context, serviceID string, paths []string) error {
	// CacheFly purge implementation would go here
	// This depends on CacheFly SDK purge methods
	return fmt.Errorf("purge cache not yet implemented")
}

// PurgeAll purges all cache for a service
func (p *CacheFlyProvider) PurgeAll(ctx context.Context, serviceID string) error {
	// CacheFly purge all implementation would go here
	return fmt.Errorf("purge all cache not yet implemented")
}

// GetMetrics retrieves metrics for a service
func (p *CacheFlyProvider) GetMetrics(ctx context.Context, serviceID string) (*domain.Metrics, error) {
	// CacheFly metrics implementation would go here
	// This depends on CacheFly SDK metrics methods
	return nil, fmt.Errorf("get metrics not yet implemented")
}

// UpdateCacheRules updates cache rules for a service
func (p *CacheFlyProvider) UpdateCacheRules(ctx context.Context, serviceID string, rules []CacheRule) error {
	// Get current options
	currentOptions, err := p.client.ServiceOptions.GetOptions(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("failed to get current options: %w", err)
	}

	// Update expiry headers
	currentOptions["expiryHeaders"] = p.buildExpiryHeaders(rules)

	// Save updated options
	_, err = p.client.ServiceOptions.UpdateOptions(ctx, serviceID, currentOptions)
	if err != nil {
		return fmt.Errorf("failed to update cache rules: %w", err)
	}

	return nil
}

// UpdateOriginSettings updates origin configuration
func (p *CacheFlyProvider) UpdateOriginSettings(ctx context.Context, serviceID string, origin OriginConfig) error {
	// Get current options
	currentOptions, err := p.client.ServiceOptions.GetOptions(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("failed to get current options: %w", err)
	}

	// Determine origin scheme
	originScheme := "HTTPS"
	if origin.Protocol != "" {
		originScheme = strings.ToUpper(origin.Protocol)
	}

	// Update reverse proxy settings
	currentOptions["reverseProxy"] = map[string]interface{}{
		"enabled":           true,
		"mode":              "WEB",
		"hostname":          origin.Host,
		"originScheme":      originScheme,
		"ttl":               86400,
		"cacheByQueryParam": false,
		"useRobotsTxt":      true,
	}

	// Save updated options
	_, err = p.client.ServiceOptions.UpdateOptions(ctx, serviceID, currentOptions)
	if err != nil {
		return fmt.Errorf("failed to update origin settings: %w", err)
	}

	return nil
}

// Helper functions

// generateServiceName creates a clean service name from input
func generateServiceName(name string) string {
	if name == "" {
		return "cdnbuddy-service"
	}

	// Convert to lowercase and replace spaces/dots with hyphens
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, ".", "-")
	name = strings.ReplaceAll(name, "_", "-")

	// Remove invalid characters
	var result strings.Builder
	for _, char := range name {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' {
			result.WriteRune(char)
		}
	}

	return result.String()
}

// buildConfigJSON builds the config JSON to store in database
func (p *CacheFlyProvider) buildConfigJSON(service *api.Service, config *ServiceConfig) string {
	configData := map[string]interface{}{
		"cachefly_service_id": service.ID,
		"unique_name":         service.UniqueName,
		"test_url":            fmt.Sprintf("https://%s.cachefly.net", service.UniqueName),
		"auto_ssl":            service.AutoSSL,
		"configuration_mode":  service.ConfigurationMode,
		"origin": map[string]interface{}{
			"host":     config.Origin.Host,
			"protocol": config.Origin.Protocol,
		},
	}

	jsonBytes, _ := json.Marshal(configData)
	return string(jsonBytes)
}
