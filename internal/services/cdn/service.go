package cdn

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/avvvet/cdnbuddy-api/internal/domain"
	"github.com/avvvet/cdnbuddy-api/internal/models"
)

type Service struct {
	provider CDNProvider
}

func NewService(provider CDNProvider) *Service {
	return &Service{
		provider: provider,
	}
}

// ListServices returns all CDN services (exposed for API handlers)
func (s *Service) ListServices(ctx context.Context) ([]domain.CDNService, error) {
	return s.provider.ListServices(ctx)
}

// ExecuteIntent handles intent responses and executes CDN operations
func (s *Service) ExecuteIntent(ctx context.Context, intent *models.IntentResponse) (string, error) {
	if intent.Action == nil {
		return "", fmt.Errorf("no action specified")
	}

	switch *intent.Action {
	case "SETUP_CDN":
		return s.handleSetupCDN(ctx, intent.Parameters)
	case "ADD_DOMAIN":
		return s.handleAddDomain(ctx, intent.Parameters)
	case "LIST_SERVICES":
		return s.handleListServices(ctx)
	default:
		return "", fmt.Errorf("unknown action: %s", *intent.Action)
	}
}

func (s *Service) handleSetupCDN(ctx context.Context, params map[string]*string) (string, error) {
	// Extract parameters
	domain := getParam(params, "domain")
	origin := getParam(params, "origin_hostname")
	if domain == "" || origin == "" {
		return "", fmt.Errorf("missing required parameters")
	}

	// Step 1: Create service (this now automatically applies best practices)
	config := &ServiceConfig{
		Name: domain,
		Origin: OriginConfig{
			Host:     origin,
			Protocol: "https",
		},
		SSL: SSLConfig{
			Enabled: true,
		},
	}

	service, err := s.provider.CreateService(ctx, config)
	if err != nil {
		return "", fmt.Errorf("failed to create service: %w", err)
	}

	// Step 2: Add domain
	err = s.provider.AddDomain(ctx, service.ID, domain)
	if err != nil {
		return "", fmt.Errorf("failed to add domain: %w", err)
	}

	// Extract test URL from config
	var configData map[string]interface{}
	json.Unmarshal([]byte(service.Config), &configData)
	testURL := configData["test_url"].(string)
	uniqueName := configData["unique_name"].(string)

	// ============================================
	// Build enhanced response with optimizations
	// ============================================
	optimizations := GetOptimizationsSummary()
	optimizationCount := GetOptimizationsCount()

	response := fmt.Sprintf(`✅ CDN configured successfully with %d optimizations!

🧪 Test URL: %s
🌐 Domain: %s (Status: Waiting for DNS)
📡 Origin: %s

🚀 Applied Optimizations:
   • %s
   • %s
   • %s
   • %s
   • %s
   • ...and %d more optimizations

📌 To activate your domain:
   1. Update DNS: Type: CNAME, Name: %s, Value: %s.cachefly.net, TTL: 300
   2. Wait 5-10 minutes for DNS propagation

Your CDN is ready to test now!`,
		optimizationCount,
		testURL,
		domain,
		origin,
		optimizations[0],
		optimizations[1],
		optimizations[2],
		optimizations[3],
		optimizations[4],
		optimizationCount-5,
		domain,
		uniqueName,
	)

	return response, nil
}

func (s *Service) handleAddDomain(ctx context.Context, params map[string]*string) (string, error) {
	serviceID := getParam(params, "service_id")
	domain := getParam(params, "domain")

	if serviceID == "" || domain == "" {
		return "", fmt.Errorf("missing required parameters")
	}

	err := s.provider.AddDomain(ctx, serviceID, domain)
	if err != nil {
		return "", fmt.Errorf("failed to add domain: %w", err)
	}

	return fmt.Sprintf("✅ Domain %s added to CDN service!", domain), nil
}

func (s *Service) handleListServices(ctx context.Context) (string, error) {
	services, err := s.provider.ListServices(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list services: %w", err)
	}

	if len(services) == 0 {
		return "You don't have any CDN services yet.", nil
	}

	response := fmt.Sprintf("You have %d CDN service(s):\n\n", len(services))
	for i, svc := range services {
		response += fmt.Sprintf("%d. %s (Status: %s)\n", i+1, svc.Name, svc.Status)
	}

	return response, nil
}

func getParam(params map[string]*string, key string) string {
	if val, ok := params[key]; ok && val != nil {
		return *val
	}
	return ""
}
