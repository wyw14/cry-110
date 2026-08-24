package pitch

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	mu          sync.Mutex
	coordinator *Coordinator
	actuators   map[int]*Actuator
	sessionID   string
}

func NewService(coordinator *Coordinator, actuators ...*Actuator) *Service {
	byBlade := make(map[int]*Actuator)
	for _, actuator := range actuators {
		if actuator != nil {
			telemetry := actuator.Telemetry(time.Now())
			byBlade[telemetry.Blade] = actuator
		}
	}
	return &Service{coordinator: coordinator, actuators: byBlade}
}

func (s *Service) StartFeather() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.actuators) != 3 {
		return "", fmt.Errorf("three pitch actuators are required")
	}
	s.sessionID = uuid.NewString()
	if err := s.coordinator.Begin(s.sessionID); err != nil {
		return "", err
	}
	return s.sessionID, nil
}

func (s *Service) Step(target, maximumStep, pressure float64, now time.Time) (GroupState, []int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.sessionID == "" {
		return GroupIdle, nil, fmt.Errorf("no pitch session is active")
	}
	for blade := 1; blade <= 3; blade++ {
		telemetry := s.actuators[blade].MoveToward(target, maximumStep, pressure, now)
		if err := s.coordinator.Observe(s.sessionID, telemetry, now); err != nil {
			return GroupIdle, nil, err
		}
	}
	state, unsafe := s.coordinator.Status(now)
	return state, unsafe, nil
}

func (s *Service) Status(now time.Time) (string, GroupState, []int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	state, unsafe := s.coordinator.Status(now)
	return s.sessionID, state, unsafe
}
