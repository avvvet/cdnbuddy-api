// cmd/server/main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/sirupsen/logrus"

	"github.com/avvvet/cdnbuddy-api/internal/config"
	"github.com/avvvet/cdnbuddy-api/internal/models"
	"github.com/avvvet/cdnbuddy-api/internal/services/cdn"
	"github.com/avvvet/cdnbuddy-api/internal/services/messaging"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logrus.Fatalf("Failed to load config: %v", err)
	}

	// Setup logrus
	setupLogger(cfg.LogLevel, cfg.Environment)

	logrus.Info("🚀 Starting CDNBuddy API Server...")

	// Initialize CacheFly provider
	cacheFlyProvider, err := cdn.NewCacheFlyProvider()
	if err != nil {
		logrus.Fatalf("Failed to initialize CacheFly provider: %v", err)
	}

	// Initialize CDN service
	cdnService := cdn.NewService(cacheFlyProvider)

	// Initialize database
	/*
		logrus.Info("📊 Connecting to database...")
		db, err := storage.NewPostgresConnection(cfg.DatabaseURL)
		if err != nil {
			logrus.Fatalf("Failed to connect to database: %v", err)
		}
		defer db.Close()
		logrus.Info("✅ Database connected")

		// Initialize repositories and services
		repo := storage.NewRepository(db)

	*/

	// Initialize NATS messaging
	logrus.Info("📡 Connecting to NATS...")
	msgClient, err := messaging.NewClient(cfg.NATSUrl)
	if err != nil {
		logrus.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer msgClient.Close()
	logrus.Info("✅ NATS connected")

	publisher := msgClient.Publisher()

	// Setup event handlers for AI Intent Service responses
	setupEventHandlers(msgClient, cdnService)

	// Create Chi router
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Custom middleware for logging request details
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			logrus.WithFields(logrus.Fields{
				"method":   r.Method,
				"path":     r.URL.Path,
				"duration": time.Since(start),
			}).Info("📥 Request processed")
		})
	})

	// Setup routes
	setupRoutes(r, publisher) // I will add db object here

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logrus.WithFields(logrus.Fields{
			"port":        cfg.Port,
			"environment": cfg.Environment,
			"database":    "connected",
			"nats":        "connected",
		}).Info("🌟 CDNBuddy API Server started")

		logrus.Info("🎯 Ready for AI Intent Service integration")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("🛑 Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server gracefully
	if err := srv.Shutdown(ctx); err != nil {
		logrus.Fatalf("Server forced to shutdown: %v", err)
	}

	logrus.Info("✅ CDNBuddy API Server exited gracefully")
}

// setupLogger configures logrus based on environment and log level
func setupLogger(logLevel, environment string) {
	// Set log level
	switch logLevel {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}

	// Set formatter based on environment
	if environment == "production" {
		// JSON formatter for production (better for log aggregation)
		logrus.SetFormatter(&logrus.JSONFormatter{})
	} else {
		// Text formatter for development (more readable)
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
			ForceColors:   true,
		})
	}
}

// setupRoutes configures the API routes
func setupRoutes(r chi.Router, publisher *messaging.Publisher) {
	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
            "status": "healthy",
            "service": "cdnbuddy-api",
            "timestamp": "` + time.Now().UTC().Format(time.RFC3339) + `"
        }`))
	})

	// API version 1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Health check
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
                "status": "healthy",
                "version": "v1",
                "service": "cdnbuddy-api"
            }`))
		})

		// CDN services endpoints
		r.Route("/cdn", func(r chi.Router) {
			r.Get("/services", func(w http.ResponseWriter, r *http.Request) {
				logrus.Info("📋 Listing CDN services")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"services": [], "message": "CDN services endpoint ready"}`))
			})

			r.Post("/services", func(w http.ResponseWriter, r *http.Request) {
				logrus.Info("➕ Creating CDN service")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte(`{"message": "CDN service creation endpoint ready"}`))
			})

			r.Get("/services/{serviceID}", func(w http.ResponseWriter, r *http.Request) {
				serviceID := chi.URLParam(r, "serviceID")
				logrus.WithField("service_id", serviceID).Info("📄 Getting CDN service details")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"service_id": "` + serviceID + `", "message": "Service details endpoint ready"}`))
			})
		})

		// Operations endpoints (for execution plans from AI)
		r.Route("/operations", func(r chi.Router) {
			r.Get("/{operationID}", func(w http.ResponseWriter, r *http.Request) {
				operationID := chi.URLParam(r, "operationID")
				logrus.WithField("operation_id", operationID).Info("📊 Getting operation status")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"operation_id": "` + operationID + `", "status": "pending"}`))
			})

			r.Post("/{operationID}/execute", func(w http.ResponseWriter, r *http.Request) {
				operationID := chi.URLParam(r, "operationID")
				logrus.WithField("operation_id", operationID).Info("⚡ Executing operation")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusAccepted)
				w.Write([]byte(`{"operation_id": "` + operationID + `", "status": "executing"}`))
			})
		})
	})

	logrus.Info("✅ Routes configured")
}

// setupEventHandlers configures NATS event subscribers for AI Intent Service integration
func setupEventHandlers(msgClient *messaging.Client, cdnService *cdn.Service) {
	subscriber := msgClient.Subscriber()

	// Handle AI Intent Service responses (execution plans)
	err := subscriber.RegisterExecutionPlanHandler(func(event messaging.ExecutionPlanEvent) error {
		logrus.WithFields(logrus.Fields{
			"user_id":    event.UserID,
			"session_id": event.SessionID,
		}).Info("🤖 AI Intent execution plan received")

		// Process the execution plan from AI Intent Service
		logrus.WithField("plan", event.Plan).Debug("📋 Execution plan details")

		return nil
	})
	if err != nil {
		logrus.WithError(err).Error("Failed to register execution plan handler")
	}

	// Handle chat messages from socket service (will forward to AI Intent Service)
	err = subscriber.RegisterChatHandler(func(event messaging.ChatEvent) error {
		logrus.WithFields(logrus.Fields{
			"user_id":    event.UserID,
			"session_id": event.SessionID,
		}).Info("💬 Chat message received")

		// Request intent analysis
		intentResponse, err := msgClient.RequestIntentAnalysis(
			context.Background(),
			event.SessionID,
			event.Message,
		)
		if err != nil {
			logrus.WithError(err).Error("❌ Failed to get response from intent service")

			// Send fallback message to user
			return msgClient.SendAIResponse(
				context.Background(),
				event.UserID,
				event.SessionID,
				"I'm sorry, I'm having trouble processing your request right now. Please try again.",
			)
		}

		logrus.WithFields(logrus.Fields{
			"session_id": event.SessionID,
			"status":     intentResponse.Status,
			"action":     intentResponse.Action,
		}).Info("📥 Received response from intent service")

		// Step 3: Handle the response based on status
		var responseMessage string

		switch intentResponse.Status {
		case "ERROR":
			// Handle error response
			if intentResponse.ErrorMessage != nil {
				logrus.WithFields(logrus.Fields{
					"session_id": event.SessionID,
					"error_code": intentResponse.ErrorCode,
					"error_msg":  *intentResponse.ErrorMessage,
				}).Error("❌ Intent service returned error")
			}
			responseMessage = intentResponse.UserMessage
			// Optional: Clear session on error to start fresh
			// msgClient.clearSession(event.SessionID)

		case "NEEDS_INFO":
			// LLM needs more information - continue conversation
			responseMessage = intentResponse.UserMessage

			logrus.WithFields(logrus.Fields{
				"session_id": event.SessionID,
				"message":    intentResponse.UserMessage,
			}).Info("🔍 Requesting more information from user")

		case "READY":
			// LLM has enough info to execute action
			if intentResponse.Action != nil {
				logrus.WithFields(logrus.Fields{
					"session_id": event.SessionID,
					"action":     *intentResponse.Action,
					"parameters": intentResponse.Parameters,
				}).Info("🎯 Executing action")

				// Execute CDN action
				result, err := cdnService.ExecuteIntent(context.Background(), intentResponse)
				if err != nil {
					logrus.WithError(err).Error("❌ Failed to execute CDN action")
					responseMessage = fmt.Sprintf("❌ Failed to complete action: %v", err)
				} else {
					responseMessage = result
				}
			} else {
				responseMessage = intentResponse.UserMessage
			}

		default:
			// Handle unknown status
			logrus.WithFields(logrus.Fields{
				"session_id": event.SessionID,
				"status":     intentResponse.Status,
			}).Warn("⚠️ Unknown intent response status")
			responseMessage = intentResponse.UserMessage
		}

		// Send the response back to the user
		return msgClient.SendAIResponse(
			context.Background(),
			event.UserID,
			event.SessionID,
			responseMessage,
		)
	})

	if err != nil {
		logrus.WithError(err).Error("Failed to register chat handler")
	}

	// Handle CDN operation events
	err = subscriber.RegisterOperationHandler(func(event messaging.OperationEvent) error {
		logrus.WithFields(logrus.Fields{
			"type":         event.Type,
			"operation_id": event.OperationID,
			"user_id":      event.UserID,
		}).Info("⚙️ CDN Operation event")

		switch event.Type {
		case messaging.EventOperationStarted:
			return msgClient.SendAIResponse(
				context.Background(),
				event.UserID,
				"current_session",
				"🔄 Starting operation: "+event.OpType,
			)

		case messaging.EventOperationProgress:
			return msgClient.SendAIResponse(
				context.Background(),
				event.UserID,
				"current_session",
				"📊 Progress: "+event.Progress,
			)

		case messaging.EventOperationCompleted:
			return msgClient.SendAIResponse(
				context.Background(),
				event.UserID,
				"current_session",
				"✅ Operation completed successfully!",
			)

		case messaging.EventOperationFailed:
			return msgClient.SendAIResponse(
				context.Background(),
				event.UserID,
				"current_session",
				"❌ Operation failed: "+event.Error,
			)
		}
		return nil
	})
	if err != nil {
		logrus.WithError(err).Error("Failed to register operation handler")
	}

	// Handle CDN service events
	err = subscriber.RegisterCDNServiceHandler(func(event messaging.CDNServiceEvent) error {
		logrus.WithFields(logrus.Fields{
			"type":       event.Type,
			"service_id": event.ServiceID,
			"user_id":    event.UserID,
			"provider":   event.Provider,
		}).Info("📢 CDN Service event")

		switch event.Type {
		case messaging.EventCDNServiceCreated:
			return msgClient.SendAIResponse(
				context.Background(),
				event.UserID,
				"current_session",
				"✅ CDN service '"+event.Name+"' created successfully with "+event.Provider+"!",
			)
		case messaging.EventCDNServiceUpdated:
			return msgClient.SendAIResponse(
				context.Background(),
				event.UserID,
				"current_session",
				"🔄 CDN service '"+event.Name+"' updated successfully!",
			)
		case messaging.EventCDNServiceDeleted:
			return msgClient.SendAIResponse(
				context.Background(),
				event.UserID,
				"current_session",
				"🗑️ CDN service deleted successfully",
			)
		}
		return nil
	})
	if err != nil {
		logrus.WithError(err).Error("Failed to register CDN service handler")
	}

	// Handle domain events
	err = subscriber.RegisterDomainHandler(func(event messaging.DomainEvent) error {
		logrus.WithFields(logrus.Fields{
			"type":           event.Type,
			"domain":         event.Name,
			"cdn_service_id": event.CDNServiceID,
		}).Info("🌐 Domain event")

		switch event.Type {
		case messaging.EventDomainAdded:
			return msgClient.SendAIResponse(
				context.Background(),
				"user_from_event", // TODO: Get user from event context
				"current_session",
				"🌐 Domain '"+event.Name+"' added to CDN successfully!",
			)
		case messaging.EventDomainStatusChanged:
			return msgClient.SendAIResponse(
				context.Background(),
				"user_from_event",
				"current_session",
				"📊 Domain '"+event.Name+"' status changed to "+event.Status,
			)
		}
		return nil
	})
	if err != nil {
		logrus.WithError(err).Error("Failed to register domain handler")
	}

	// Handle cache events
	err = subscriber.RegisterCacheHandler(func(event messaging.CacheEvent) error {
		logrus.WithFields(logrus.Fields{
			"type":       event.Type,
			"service_id": event.ServiceID,
			"user_id":    event.UserID,
		}).Info("💾 Cache event")

		switch event.Type {
		case messaging.EventCachePurged:
			msg := "🧹 Cache purged successfully!"
			if len(event.Paths) > 0 {
				msg = "🧹 Cache purged for specific paths"
				logrus.WithField("paths", event.Paths).Debug("Purged paths")
			}
			return msgClient.SendAIResponse(
				context.Background(),
				event.UserID,
				"current_session",
				msg,
			)
		}
		return nil
	})
	if err != nil {
		logrus.WithError(err).Error("Failed to register cache handler")
	}

	// Handle CDN status requests from Socket Server
	err = subscriber.RegisterStatusRequestHandler(func(event messaging.StatusRequestEvent) error {
		logrus.WithFields(logrus.Fields{
			"user_id":    event.UserID,
			"session_id": event.SessionID,
		}).Info("📡 CDN status request received")

		// Fetch real services from CacheFly
		ctx := context.Background()
		services, err := cdnService.ListServices(ctx)
		if err != nil {
			logrus.WithError(err).Error("❌ Failed to fetch CDN services")
			// Send empty response on error
			return msgClient.Publisher().PublishStatusResponse(event.UserID, event.SessionID, []messaging.ServiceStatus{})
		}

		// Convert to response format
		statusServices := make([]messaging.ServiceStatus, 0, len(services))
		for _, svc := range services {
			// Parse config JSON to get test URL
			var config map[string]interface{}
			json.Unmarshal([]byte(svc.Config), &config)

			testURL := ""
			if url, ok := config["test_url"].(string); ok {
				testURL = url
			}

			statusServices = append(statusServices, messaging.ServiceStatus{
				ID:       svc.ID,
				Name:     svc.Name,
				Status:   svc.Status,
				TestURL:  testURL,
				Provider: string(svc.Provider),
			})
		}

		logrus.WithField("count", len(statusServices)).Info("✅ Sending CDN status response")

		// Send response back to Socket Server
		return msgClient.Publisher().PublishStatusResponse(event.UserID, event.SessionID, statusServices)
	})
	if err != nil {
		logrus.WithError(err).Error("Failed to register status request handler")
	}

	logrus.Info("✅ Event handlers configured for AI Intent Service integration")
}

func createIntentRequest(event messaging.ChatEvent) models.IntentRequest {
	return models.IntentRequest{
		SessionID:           event.SessionID,
		UserMessage:         event.Message,
		ConversationHistory: []models.ConversationMessage{}, // Empty for now
		AvailableActions:    []models.ActionSchema{},        // Empty for now
	}
}
