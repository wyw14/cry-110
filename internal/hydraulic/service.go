package hydraulic

import (
	"fmt"
	"sync"

	"github.com/wyw14/cry-110/internal/interlock"
	"github.com/wyw14/cry-110/internal/pitch"
)

type Service struct {
	mu          sync.Mutex
	accumulator Accumulator
	minimumBar  float64
	budget      *pitch.EnergyBudget
	permit      interlock.RunPermit
}

func NewService(accumulator Accumulator, minimumBar float64, budget *pitch.EnergyBudget, permit interlock.RunPermit) *Service {
	return &Service{accumulator: accumulator, minimumBar: minimumBar, budget: budget, permit: permit}
}

func (s *Service) Update(accumulator Accumulator) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accumulator = accumulator
}

func (s *Service) EmergencyPitchProof() (interlock.EnergyProof, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	usable, err := s.accumulator.UsableVolume(s.minimumBar)
	if err != nil {
		return interlock.EnergyProof{}, err
	}
	return s.budget.Proof(usable, s.accumulator.HeaderBar, s.accumulator.PrechargeBar)
}

func (s *Service) Ready() error {
	proof, err := s.EmergencyPitchProof()
	if err != nil {
		return fmt.Errorf("calculate emergency pitch energy: %w", err)
	}
	return s.permit.EmergencyPitchReady(proof)
}
