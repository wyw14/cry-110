package icing

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/cry-110/internal/interlock"
	"github.com/wyw14/cry-110/internal/model"
	"github.com/wyw14/cry-110/internal/rotor"
)

type ResponseState string

const (
	ResponseInactive   ResponseState = "inactive"
	ResponseStopping   ResponseState = "stopping"
	ResponseStopped    ResponseState = "stopped"
	ResponseTesting    ResponseState = "testing"
	ResponseInspecting ResponseState = "inspecting"
	ResponseCleared    ResponseState = "cleared"
)

type Response struct {
	mu        sync.Mutex
	exclusion *interlock.Exclusion
	cycle     *rotor.TestCycle
	state     ResponseState
	session   string
	incident  model.Incident
}

func NewResponse(exclusion *interlock.Exclusion, cycle *rotor.TestCycle) *Response {
	return &Response{exclusion: exclusion, cycle: cycle, state: ResponseInactive}
}

func (r *Response) Trigger(message string, now time.Time) (model.Incident, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.state != ResponseInactive && r.state != ResponseCleared {
		return model.Incident{}, fmt.Errorf("icing response is already active")
	}
	r.session = uuid.NewString()
	if err := r.exclusion.Lock(r.session); err != nil {
		return model.Incident{}, err
	}
	incident, err := model.NewIncident(r.session, model.IncidentIcing, message, now)
	if err != nil {
		return model.Incident{}, err
	}
	r.incident = incident
	r.state = ResponseStopping
	return incident, nil
}

func (r *Response) OnStopped() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.state != ResponseStopping {
		return fmt.Errorf("icing response is not stopping the rotor")
	}
	r.state = ResponseStopped
	return nil
}

func (r *Response) StartShedCycle() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.state != ResponseStopped {
		return fmt.Errorf("rotor must be stopped before the shed cycle")
	}
	if err := r.cycle.Start(r.session); err != nil {
		return err
	}
	r.state = ResponseTesting
	return nil
}

func (r *Response) Inspect() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.state != ResponseTesting {
		return fmt.Errorf("controlled shed cycle is not active")
	}
	if err := r.cycle.StopForInspection(); err != nil {
		return err
	}
	r.state = ResponseInspecting
	return nil
}

func (r *Response) Clear(noResidualIce bool, now time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.state != ResponseInspecting {
		return fmt.Errorf("icing inspection is not active")
	}
	if err := r.cycle.Complete(noResidualIce); err != nil {
		return err
	}
	r.incident = r.incident.Clear(now)
	r.state = ResponseCleared
	return nil
}

func (r *Response) Status() (ResponseState, string, model.Incident, rotor.TestCycleState, interlock.ExclusionState) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.state, r.session, r.incident, r.cycle.State(), r.exclusion.State()
}
