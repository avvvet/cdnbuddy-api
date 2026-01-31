package cdn

import (
	api "github.com/cachefly/cachefly-go-sdk/pkg/cachefly/api/v2_5"
)

// GetBestPracticesOptions returns optimized service options following industry best practices
func GetBestPracticesOptions(domain, originHostname, originScheme string) api.ServiceOptions {
	return api.ServiceOptions{
		// ============================================
		// ORIGIN CONFIGURATION (Required)
		// ============================================
		"reverseProxy": map[string]interface{}{
			"enabled":           true,
			"mode":              "WEB",
			"hostname":          originHostname,
			"originScheme":      originScheme,
			"ttl":               2678400, // 31 days for aggressive caching
			"cacheByQueryParam": true,    // Cache varies by query params
			"useRobotsTxt":      true,    // Respect robots.txt
		},

		// Origin host header
		"originhostheader": map[string]interface{}{
			"enabled": true,
			"value":   []string{originHostname},
		},

		// ============================================
		// PERFORMANCE OPTIMIZATIONS
		// ============================================

		// Connection optimization
		"allowretry":       true, // Retry failed requests for better reliability
		"forceorigqstring": true, // Preserve query strings for dynamic content
		"send-xff":         true, // Send X-Forwarded-For header for analytics

		// Compression (CRITICAL for performance)
		"brotli_support": true, // Enable Brotli compression (better than gzip)

		// Timeout optimization
		"ttfb_timeout": map[string]interface{}{
			"enabled": true,
			"value":   30, // 30 seconds - reasonable for most origins
		},
		"contimeout": map[string]interface{}{
			"enabled": true,
			"value":   10, // 10 seconds connection timeout
		},

		// Connection pooling
		"maxcons": map[string]interface{}{
			"enabled": true,
			"value":   100, // Max 100 concurrent connections per origin
		},

		// ============================================
		// CACHING OPTIMIZATION (AGGRESSIVE)
		// ============================================

		"servestale":           true, // Serve stale content if origin is down
		"normalizequerystring": true, // Normalize query strings for better cache hit ratio
		"cachebygeocountry":    true, // Cache by country for geo-specific content
		"cachebyregion":        true, // Cache by region for better performance

		// Cache purge optimization
		"purgenoquery": true, // Purge ignores query strings
		"purgemode": map[string]interface{}{
			"enabled": true,
			"value":   "2", // Smart purge mode
		},
		"dirpurgeskip": map[string]interface{}{
			"enabled": true,
			"value":   1, // Skip directory purge for better performance
		},

		// Error caching
		"error_ttl": map[string]interface{}{
			"enabled": true,
			"value":   300, // Cache errors for 5 minutes (prevent origin overload)
		},

		// ============================================
		// DELIVERY OPTIMIZATION
		// ============================================

		"cors":          true, // Enable CORS for modern web apps
		"autoRedirect":  true, // Auto-redirect for better UX
		"livestreaming": true, // Enable live streaming support
		"linkpreheat":   true, // Preload linked resources

		// File encoding optimization
		"skip_encoding_ext": map[string]interface{}{
			"enabled": true,
			"value":   []string{".zip", ".gz", ".tar", ".rar", ".7z", ".bz2"}, // Don't re-compress
		},

		// HTTP methods (secure defaults)
		"httpmethods": map[string]interface{}{
			"enabled": true,
			"value": map[string]interface{}{
				"GET":     true,  // Read operations
				"POST":    true,  // Form submissions
				"HEAD":    true,  // Metadata requests
				"OPTIONS": true,  // CORS preflight
				"PUT":     false, // Disabled for security
				"DELETE":  false, // Disabled for security
				"PATCH":   false, // Disabled for security
			},
		},

		// ============================================
		// SECURITY (Basic but important)
		// ============================================

		"protectServeKeyEnabled": true, // Enable secure token authentication
		"apiKeyEnabled":          true, // Enable API key authentication

		// Don't protect common static assets (for better performance)
		"skip_pserve_ext": map[string]interface{}{
			"enabled": true,
			"value":   []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".css", ".js", ".woff", ".woff2", ".ttf", ".svg", ".ico"},
		},

		// Empty expiry headers (let CDN manage)
		"expiryHeaders": []interface{}{},
	}
}

// GetOptimizationsSummary returns a human-readable list of applied optimizations
func GetOptimizationsSummary() []string {
	return []string{
		"Aggressive caching for static assets (31-day TTL)",
		"Brotli compression enabled (30% better than gzip)",
		"Serve stale content if origin is unavailable",
		"Smart query string normalization",
		"Geographic caching optimization",
		"Connection pooling (100 concurrent connections)",
		"CORS enabled for modern web apps",
		"Auto-redirect and link preheating",
		"Optimized timeouts (10s connect, 30s TTFB)",
		"Secure HTTP methods (GET, POST, HEAD, OPTIONS only)",
		"Smart file encoding (skip pre-compressed files)",
		"Error caching (5-minute TTL to protect origin)",
	}
}

// GetOptimizationsCount returns the number of optimizations applied
func GetOptimizationsCount() int {
	return len(GetOptimizationsSummary())
}
