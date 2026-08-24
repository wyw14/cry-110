package ventilation

import (
	"fmt"
	"sync"

	"github.com/wyw14/cry-110/internal/interlock"
)

type Mode string

const (
	ModeStopped Mode = "stopped"
	ModeCooling Mode = "cooling"
	ModePurge   Mode = "fire_purge"
)

type CoolingService struct {
	mu    sync.Mutex
	mode  Mode
	flow  float64
	route interlock.FireRoute
}

func NewCoolingService() *CoolingService {
	return &CoolingService{mode: ModeStopped, route: interlock.NewFireRoute()}
}

func (s *CoolingService) StartCooling(flow float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if flow <= 0 {
		return fmt.Errorf("cooling flow must be positive")
	}
	if s.route.Latched {
		return fmt.Errorf("fire route blocks normal cooling")
	}
	s.mode = ModeCooling
	s.flow = flow
	return nil
}

func (s *CoolingService) StartFirePurge(route interlock.FireRoute, flow float64) error {
	if err := route.ValidatePurge(); err != nil {
		return err
	}
	if flow <= 0 {
		return fmt.Errorf("purge flow must be positive")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.route = route
	s.mode = ModePurge
	s.flow = flow
	return nil
}

func (s *CoolingService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mode = ModeStopped
	s.flow = 0
}

func (s *CoolingService) Status() (Mode, float64, interlock.FireRoute) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mode, s.flow, s.route
}
