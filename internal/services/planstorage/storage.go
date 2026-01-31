package planstorage

import (
	"fmt"
	"sync"
	"time"

	"github.com/avvvet/cdnbuddy-api/internal/models"
	"github.com/sirupsen/logrus"
)

// Storage manages pending execution plans in memory
type Storage struct {
	plans map[string]*models.ExecutionPlan
	mu    sync.RWMutex
}

// NewStorage creates a new plan storage
func NewStorage() *Storage {
	s := &Storage{
		plans: make(map[string]*models.ExecutionPlan),
	}

	// Start cleanup goroutine for expired plans
	go s.cleanupExpired()

	return s
}

// Store saves an execution plan
func (s *Storage) Store(plan models.ExecutionPlan) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.plans[plan.ID] = &plan
	logrus.WithField("plan_id", plan.ID).Info("📦 Stored execution plan")
	return nil
}

// Get retrieves a plan by ID
func (s *Storage) Get(planID string) (*models.ExecutionPlan, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	plan, exists := s.plans[planID]
	if !exists {
		return nil, fmt.Errorf("plan not found: %s", planID)
	}

	// Check if expired
	if time.Now().After(plan.ExpiresAt) {
		return nil, fmt.Errorf("plan expired: %s", planID)
	}

	return plan, nil
}

// Delete removes a plan by ID
func (s *Storage) Delete(planID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.plans, planID)
	logrus.WithField("plan_id", planID).Info("🗑️ Deleted execution plan")
}

// cleanupExpired removes expired plans periodically
func (s *Storage) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		count := 0

		for id, plan := range s.plans {
			if now.After(plan.ExpiresAt) {
				delete(s.plans, id)
				count++
			}
		}

		if count > 0 {
			logrus.WithField("count", count).Info("🧹 Cleaned up expired plans")
		}
		s.mu.Unlock()
	}
}
