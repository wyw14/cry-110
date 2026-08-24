package yaw

import (
	"fmt"
	"math"
	"sync"

	"github.com/google/uuid"
	"github.com/wyw14/cry-110/internal/interlock"
	"github.com/wyw14/cry-110/internal/model"
)

type Planner struct {
	mu        sync.Mutex
	angle     float64
	book      *interlock.ReservationBook
	motorGoal *model.MotionTarget
}

func NewPlanner(angle float64, book *interlock.ReservationBook) *Planner {
	return &Planner{angle: angle, book: book}
}

func (p *Planner) Plan(target, seconds float64) (model.MotionTarget, []interlock.SweptVolume, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if seconds <= 0 || math.Abs(target-p.angle) > 15 {
		return model.MotionTarget{}, nil, fmt.Errorf("yaw correction exceeds the service envelope")
	}
	id := uuid.NewString()
	motion := model.MotionTarget{OperationID: id, Axis: "nacelle-yaw", Start: p.angle, End: target, Seconds: seconds}
	volume := interlock.SweptVolume{
		OperationID: id, Axis: motion.Axis,
		MinAngle: math.Min(p.angle, target), MaxAngle: math.Max(p.angle, target),
		MinRadiusM: 2, MaxRadiusM: 7, StartSecond: 0, EndSecond: seconds,
	}
	return motion, []interlock.SweptVolume{volume}, nil
}

func (p *Planner) Commit(motion model.MotionTarget, volumes []interlock.SweptVolume) error {
	if err := motion.Validate(); err != nil {
		return err
	}
	if err := p.book.TryReserve(motion.OperationID, volumes); err != nil {
		return err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	copyTarget := motion
	p.motorGoal = &copyTarget
	return nil
}

func (p *Planner) Complete(operationID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.motorGoal == nil || p.motorGoal.OperationID != operationID {
		return fmt.Errorf("yaw operation is not active")
	}
	p.angle = p.motorGoal.End
	p.motorGoal = nil
	p.book.Release(operationID)
	return nil
}

func (p *Planner) Status() (float64, *model.MotionTarget) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.motorGoal == nil {
		return p.angle, nil
	}
	copyTarget := *p.motorGoal
	return p.angle, &copyTarget
}
