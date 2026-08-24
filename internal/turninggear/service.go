package turninggear

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	mu         sync.Mutex
	controller *Controller
	permit     *Permit
	disengage  *Disengage
	operation  string
	target     float64
}

func NewService(controller *Controller, permit *Permit, disengage *Disengage) *Service {
	return &Service{controller: controller, permit: permit, disengage: disengage}
}

func (s *Service) Start(target, rawEncoder float64, now time.Time) (string, float64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.permit.Evaluate(now); err != nil {
		return "", 0, err
	}
	command, err := s.controller.Command(target, rawEncoder)
	if err != nil {
		return "", 0, err
	}
	s.operation = uuid.NewString()
	s.target = target
	return s.operation, command, nil
}

func (s *Service) Update(rawEncoder float64) (float64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.operation == "" {
		return 0, fmt.Errorf("no turning gear operation is active")
	}
	return s.controller.Command(s.target, rawEncoder)
}

func (s *Service) BeginDisengage() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.operation == "" {
		return fmt.Errorf("no turning gear operation is active")
	}
	return s.disengage.Start(s.operation)
}

func (s *Service) ApplyDisengageFeedback(amperes float64, pawlState string, now time.Time) (DisengageState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.operation == "" {
		return DisengageMeshed, fmt.Errorf("no turning gear operation is active")
	}
	return s.disengage.ApplyCurrent(amperes, pawlState, now), nil
}

func (s *Service) Status() (string, float64, DisengageState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.operation, s.target, s.disengage.State()
}
