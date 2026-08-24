package access

import (
	"fmt"
	"sync"

	"github.com/wyw14/cry-110/internal/interlock"
)

type Service struct {
	mu        sync.Mutex
	exclusion *interlock.Exclusion
	present   int
}

func NewService(exclusion *interlock.Exclusion) *Service {
	return &Service{exclusion: exclusion}
}

func (s *Service) Enter() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.exclusion.AccessAllowed() {
		return fmt.Errorf("tower base access is excluded by an active hazard")
	}
	s.present++
	return nil
}

func (s *Service) Leave() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.present == 0 {
		return fmt.Errorf("tower base occupancy is already empty")
	}
	s.present--
	return nil
}

func (s *Service) Occupancy() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.present
}
