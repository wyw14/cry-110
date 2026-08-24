package brake

import (
	"fmt"
	"sync"
	"time"

	"github.com/wyw14/cry-110/internal/interlock"
)

type Service struct {
	mu       sync.Mutex
	thermal  *ThermalModel
	limit    interlock.YawLimit
	pressure bool
}

func NewService(thermal *ThermalModel, limit interlock.YawLimit) *Service {
	return &Service{thermal: thermal, limit: limit}
}

func (s *Service) Apply(inertia, initialRPM, finalRPM float64, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	state, err := s.thermal.ApplyBraking(inertia, initialRPM, finalRPM, now)
	if err != nil {
		return err
	}
	if !s.limit.Ready(state) {
		return fmt.Errorf("brake thermal capacity is exhausted")
	}
	s.pressure = true
	return nil
}

func (s *Service) ReleasePressure(now time.Time) interlock.ThermalState {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pressure = false
	return s.thermal.ReleasePressure(now)
}

func (s *Service) Thermal(now time.Time) interlock.ThermalState {
	return s.thermal.State(now)
}

func (s *Service) Pressurized() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pressure
}
