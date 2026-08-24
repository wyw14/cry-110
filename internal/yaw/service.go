package yaw

import (
	"fmt"
	"sync"
	"time"

	"github.com/wyw14/cry-110/internal/brake"
	"github.com/wyw14/cry-110/internal/interlock"
	"github.com/wyw14/cry-110/internal/model"
)

type Service struct {
	mu      sync.Mutex
	planner *Planner
	brake   *brake.Service
	limit   interlock.YawLimit
}

func NewService(planner *Planner, brakeService *brake.Service, limit interlock.YawLimit) *Service {
	return &Service{planner: planner, brake: brakeService, limit: limit}
}

func (s *Service) Request(target, requestedSpeed float64, now time.Time) (model.MotionTarget, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	allowed, err := s.limit.AllowedSpeed(requestedSpeed, s.brake.Thermal(now))
	if err != nil {
		return model.MotionTarget{}, err
	}
	angle, _ := s.planner.Status()
	seconds := 1.0
	if delta := target - angle; delta != 0 {
		if delta < 0 {
			delta = -delta
		}
		seconds = delta / allowed
	}
	motion, volumes, err := s.planner.Plan(target, seconds)
	if err != nil {
		return model.MotionTarget{}, err
	}
	if err := s.planner.Commit(motion, volumes); err != nil {
		return model.MotionTarget{}, err
	}
	return motion, nil
}

func (s *Service) Finish(operationID string, inertia, initialRPM float64, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.planner.Complete(operationID); err != nil {
		return err
	}
	if err := s.brake.Apply(inertia, initialRPM, 0, now); err != nil {
		return fmt.Errorf("apply yaw brake: %w", err)
	}
	s.brake.ReleasePressure(now)
	return nil
}

func (s *Service) Status(now time.Time) (float64, *model.MotionTarget, interlock.ThermalState) {
	angle, motion := s.planner.Status()
	return angle, motion, s.brake.Thermal(now)
}
