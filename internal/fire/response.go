package fire

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/cry-110/internal/interlock"
	"github.com/wyw14/cry-110/internal/model"
	"github.com/wyw14/cry-110/internal/ventilation"
)

type Response struct {
	mu          sync.Mutex
	ventilation *ventilation.CoolingService
	route       interlock.FireRoute
	incident    model.Incident
	active      bool
}

func NewResponse(ventilationService *ventilation.CoolingService) *Response {
	return &Response{ventilation: ventilationService, route: interlock.NewFireRoute()}
}

func (r *Response) StartPurge(message string, flow float64, now time.Time) (model.Incident, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.active {
		return r.incident, fmt.Errorf("fire response is already active")
	}
	incident, err := model.NewIncident(uuid.NewString(), model.IncidentFire, message, now)
	if err != nil {
		return model.Incident{}, err
	}
	if err := r.ventilation.StartCooling(flow); err != nil {
		return model.Incident{}, err
	}
	r.incident = incident
	r.active = true
	return incident, nil
}

func (r *Response) Status() (bool, model.Incident, interlock.FireRoute) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.active, r.incident, r.route
}

func (r *Response) Reset(inspectionPassed bool, now time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.active {
		return fmt.Errorf("no fire response is active")
	}
	if !inspectionPassed {
		return fmt.Errorf("fire inspection has not passed")
	}
	r.incident = r.incident.Clear(now)
	r.ventilation.Stop()
	r.route = interlock.NewFireRoute()
	r.active = false
	return nil
}
