package generator

import (
	"fmt"
	"sync"

	"github.com/wyw14/cry-110/internal/hydraulic"
)

type State string

const (
	StateStopped  State = "stopped"
	StateStarting State = "starting"
	StateRunning  State = "running"
	StateTripped  State = "tripped"
)

type Service struct {
	mu        sync.Mutex
	hydraulic *hydraulic.Service
	state     State
	outputKW  float64
}

func NewService(hydraulicService *hydraulic.Service) *Service {
	return &Service{hydraulic: hydraulicService, state: StateStopped}
}

func (s *Service) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state != StateStopped {
		return fmt.Errorf("generator cannot start from %s", s.state)
	}
	if err := s.hydraulic.Ready(); err != nil {
		return fmt.Errorf("generator run permit denied: %w", err)
	}
	s.state = StateStarting
	return nil
}

func (s *Service) Synchronize(outputKW float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state != StateStarting || outputKW < 0 {
		return fmt.Errorf("generator cannot synchronize")
	}
	s.outputKW = outputKW
	s.state = StateRunning
	return nil
}

func (s *Service) Trip() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.outputKW = 0
	s.state = StateTripped
}

func (s *Service) Reset() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state != StateTripped {
		return fmt.Errorf("generator is not tripped")
	}
	s.state = StateStopped
	return nil
}

func (s *Service) Status() (State, float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state, s.outputKW
}
