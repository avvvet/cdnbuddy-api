package messaging

import (
	"encoding/json"
	"log"

	"github.com/nats-io/nats.go"
)

type Subscriber struct {
	client   *NATSClient
	handlers map[string][]MessageHandler
}

type MessageHandler func(data []byte) error

func NewSubscriber(client *NATSClient) *Subscriber {
	return &Subscriber{
		client:   client,
		handlers: make(map[string][]MessageHandler),
	}
}

// Register handlers for different message types
func (s *Subscriber) RegisterCDNServiceHandler(handler func(event CDNServiceEvent) error) error {
	messageHandler := func(data []byte) error {
		var event CDNServiceEvent
		if err := json.Unmarshal(data, &event); err != nil {
			return err
		}
		return handler(event)
	}

	return s.subscribe(SubjectCDNService, messageHandler)
}

func (s *Subscriber) RegisterDomainHandler(handler func(event DomainEvent) error) error {
	messageHandler := func(data []byte) error {
		var event DomainEvent
		if err := json.Unmarshal(data, &event); err != nil {
			return err
		}
		return handler(event)
	}

	return s.subscribe(SubjectDomain, messageHandler)
}

func (s *Subscriber) RegisterCacheHandler(handler func(event CacheEvent) error) error {
	messageHandler := func(data []byte) error {
		var event CacheEvent
		if err := json.Unmarshal(data, &event); err != nil {
			return err
		}
		return handler(event)
	}

	return s.subscribe(SubjectCache, messageHandler)
}

func (s *Subscriber) RegisterMetricsHandler(handler func(event MetricsEvent) error) error {
	messageHandler := func(data []byte) error {
		var event MetricsEvent
		if err := json.Unmarshal(data, &event); err != nil {
			return err
		}
		return handler(event)
	}

	return s.subscribe(SubjectMetrics, messageHandler)
}

func (s *Subscriber) RegisterOperationHandler(handler func(event OperationEvent) error) error {
	messageHandler := func(data []byte) error {
		var event OperationEvent
		if err := json.Unmarshal(data, &event); err != nil {
			return err
		}
		return handler(event)
	}

	return s.subscribe(SubjectOperation, messageHandler)
}

func (s *Subscriber) RegisterChatHandler(handler func(event ChatEvent) error) error {
	messageHandler := func(data []byte) error {
		var event ChatEvent
		if err := json.Unmarshal(data, &event); err != nil {
			return err
		}
		return handler(event)
	}

	return s.subscribe(SubjectChat, messageHandler)
}

func (s *Subscriber) RegisterExecutionPlanHandler(handler func(event ExecutionPlanEvent) error) error {
	messageHandler := func(data []byte) error {
		var event ExecutionPlanEvent
		if err := json.Unmarshal(data, &event); err != nil {
			return err
		}
		return handler(event)
	}

	return s.subscribe(SubjectExecutionPlan, messageHandler)
}

// RegisterStatusRequestHandler registers handler for CDN status requests
func (s *Subscriber) RegisterStatusRequestHandler(handler func(event StatusRequestEvent) error) error {
	messageHandler := func(data []byte) error {
		var event StatusRequestEvent
		if err := json.Unmarshal(data, &event); err != nil {
			return err
		}
		return handler(event)
	}

	return s.subscribe("cdn.status.request", messageHandler)
}

// Generic subscription method
func (s *Subscriber) subscribe(subject string, handler MessageHandler) error {
	// Add handler to registry
	s.handlers[subject] = append(s.handlers[subject], handler)

	// Subscribe to NATS subject
	_, err := s.client.Subscribe(subject, func(msg *nats.Msg) {
		// Process message with all registered handlers for this subject
		for _, h := range s.handlers[subject] {
			if err := h(msg.Data); err != nil {
				log.Printf("❌ Error processing message on subject %s: %v", subject, err)
			}
		}
	})

	if err != nil {
		return err
	}

	log.Printf("📥 Subscribed to subject: %s", subject)
	return nil
}

// Queue subscription for load balancing
func (s *Subscriber) QueueSubscribe(subject, queue string, handler MessageHandler) error {
	_, err := s.client.QueueSubscribe(subject, queue, func(msg *nats.Msg) {
		if err := handler(msg.Data); err != nil {
			log.Printf("❌ Error processing queued message on subject %s: %v", subject, err)
		}
	})

	if err != nil {
		return err
	}

	log.Printf("📥 Queue subscribed to subject: %s (queue: %s)", subject, queue)
	return nil
}

// Request-Reply pattern
func (s *Subscriber) RegisterRequestHandler(subject string, handler func(data []byte) (interface{}, error)) error {
	_, err := s.client.Subscribe(subject, func(msg *nats.Msg) {
		response, err := handler(msg.Data)
		if err != nil {
			log.Printf("❌ Error processing request on subject %s: %v", subject, err)
			// Send error response
			errorResponse := map[string]string{"error": err.Error()}
			if responseData, marshalErr := json.Marshal(errorResponse); marshalErr == nil {
				msg.Respond(responseData)
			}
			return
		}

		// Send successful response
		if responseData, err := json.Marshal(response); err == nil {
			msg.Respond(responseData)
		} else {
			log.Printf("❌ Error marshaling response: %v", err)
		}
	})

	if err != nil {
		return err
	}

	log.Printf("📥 Request handler registered for subject: %s", subject)
	return nil
}

// RegisterExecuteCommandHandler registers handler for execution commands
func (s *Subscriber) RegisterExecuteCommandHandler(handler func(event ExecuteCommand) error) error {
	messageHandler := func(data []byte) error {
		var event ExecuteCommand
		if err := json.Unmarshal(data, &event); err != nil {
			return err
		}
		return handler(event)
	}

	return s.subscribe("cdnbuddy.execute", messageHandler)
}
