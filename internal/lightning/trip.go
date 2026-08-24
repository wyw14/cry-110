package lightning

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/cry-110/internal/crane"
	"github.com/wyw14/cry-110/internal/model"
	"github.com/wyw14/cry-110/internal/yaw"
)

type Trip struct {
	mu       sync.Mutex
	yawBrake *yaw.BrakeController
	crane    *crane.Park
	incident model.Incident
	active   bool
}

func NewTrip(yawBrake *yaw.BrakeController, cranePark *crane.Park) *Trip {
	return &Trip{yawBrake: yawBrake, crane: cranePark}
}

func (t *Trip) Trigger(message string, now time.Time) (model.Incident, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.active {
		return t.incident, fmt.Errorf("lightning trip is already active")
	}
	incident, err := model.NewIncident(uuid.NewString(), model.IncidentLightning, message, now)
	if err != nil {
		return model.Incident{}, err
	}
	if err := t.yawBrake.Lock(8000, 0.2, now); err != nil {
		return model.Incident{}, err
	}
	t.crane.BeginPark(now)
	t.incident = incident
	t.active = true
	return incident, nil
}

func (t *Trip) ConfirmCraneParked(now time.Time) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.active {
		return fmt.Errorf("no lightning trip is active")
	}
	return t.crane.ConfirmSecured(0, 0, now)
}

func (t *Trip) Status() (bool, model.Incident) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.active, t.incident
}

func (t *Trip) Reset(inspectionPassed bool, now time.Time) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.active {
		return fmt.Errorf("no lightning trip is active")
	}
	if !inspectionPassed {
		return fmt.Errorf("lightning inspection has not passed")
	}
	state, _, _, _ := t.crane.Status()
	if state != crane.ParkSecured {
		return fmt.Errorf("crane is not secured after the lightning trip")
	}
	t.yawBrake.Unlock(now)
	t.incident = t.incident.Clear(now)
	t.active = false
	return nil
}
