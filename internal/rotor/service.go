package rotor

import (
	"fmt"
	"sync"
	"time"

	"github.com/wyw14/cry-110/internal/interlock"
	"github.com/wyw14/cry-110/internal/model"
)

type Service struct {
	mu        sync.Mutex
	observer  *StillnessObserver
	pinPermit *interlock.PinPermit
	latest    model.RotorTelemetry
	braked    bool
}

func NewService(observer *StillnessObserver, pinPermit *interlock.PinPermit) *Service {
	return &Service{observer: observer, pinPermit: pinPermit}
}

func (s *Service) Record(sample model.RotorTelemetry) error {
	if err := sample.Validate(); err != nil {
		return err
	}
	proof := s.observer.Observe(sample)
	s.pinPermit.Update(proof)
	s.mu.Lock()
	s.latest = sample
	s.mu.Unlock()
	return nil
}

func (s *Service) ResetStillness() {
	s.observer.Reset()
	s.pinPermit.Withdraw()
	s.mu.Lock()
	s.latest = model.RotorTelemetry{}
	s.braked = false
	s.mu.Unlock()
}

func (s *Service) ApplyBrake() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.latest.ObservedAt.IsZero() {
		return fmt.Errorf("rotor speed is unavailable")
	}
	if s.latest.AverageRPM > 2 || s.latest.AverageRPM < -2 {
		return fmt.Errorf("rotor speed is too high for the mechanical brake")
	}
	s.braked = true
	return nil
}

func (s *Service) InsertPin(now time.Time) error {
	if err := s.pinPermit.Begin(now); err != nil {
		return err
	}
	return s.pinPermit.Complete(now)
}

func (s *Service) Status() (model.RotorTelemetry, bool, interlock.PinState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.latest, s.braked, s.pinPermit.State()
}
