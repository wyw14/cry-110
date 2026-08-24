package crane

import (
	"fmt"
	"math"
	"sync"

	"github.com/google/uuid"
	"github.com/wyw14/cry-110/internal/interlock"
	"github.com/wyw14/cry-110/internal/model"
)

type Slew struct {
	mu        sync.Mutex
	angle     float64
	radius    float64
	book      *interlock.ReservationBook
	motorGoal *model.MotionTarget
}

func NewSlew(angle, radius float64, book *interlock.ReservationBook) *Slew {
	return &Slew{angle: angle, radius: radius, book: book}
}

func (s *Slew) Plan(target, seconds float64) (model.MotionTarget, []interlock.SweptVolume, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if seconds <= 0 || math.Abs(target-s.angle) > 180 {
		return model.MotionTarget{}, nil, fmt.Errorf("crane slew target is outside the safe planning envelope")
	}
	id := uuid.NewString()
	motion := model.MotionTarget{OperationID: id, Axis: "crane-slew", Start: s.angle, End: target, Seconds: seconds}
	volume := interlock.SweptVolume{
		OperationID: id, Axis: motion.Axis,
		MinAngle: math.Min(s.angle, target), MaxAngle: math.Max(s.angle, target),
		MinRadiusM: math.Max(0, s.radius-0.5), MaxRadiusM: s.radius + 0.5,
		StartSecond: 0, EndSecond: seconds,
	}
	return motion, []interlock.SweptVolume{volume}, nil
}

func (s *Slew) Commit(motion model.MotionTarget, volumes []interlock.SweptVolume) error {
	if err := motion.Validate(); err != nil {
		return err
	}
	if err := s.book.TryReserve(motion.OperationID, volumes); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	copyTarget := motion
	s.motorGoal = &copyTarget
	return nil
}

func (s *Slew) Complete(operationID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.motorGoal == nil || s.motorGoal.OperationID != operationID {
		return fmt.Errorf("crane slew operation is not active")
	}
	s.angle = s.motorGoal.End
	s.motorGoal = nil
	s.book.Release(operationID)
	return nil
}

func (s *Slew) Status() (float64, *model.MotionTarget) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.motorGoal == nil {
		return s.angle, nil
	}
	copyTarget := *s.motorGoal
	return s.angle, &copyTarget
}
