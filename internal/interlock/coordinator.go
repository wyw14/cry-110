package interlock

import (
	"fmt"
	"sync"
	"time"

	"github.com/wyw14/cry-110/internal/model"
)

type Coordinator struct {
	mu         sync.Mutex
	isolation  model.Isolation
	hasSession bool
	incidents  map[string]model.Incident
}

func NewCoordinator() *Coordinator {
	return &Coordinator{incidents: make(map[string]model.Incident)}
}

func (c *Coordinator) StartIsolation(id, turbineID, reason string, now time.Time) (model.Isolation, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.hasSession && !c.isolation.Terminal() {
		return model.Isolation{}, fmt.Errorf("isolation %s is already active", c.isolation.ID)
	}
	isolation, err := model.NewIsolation(id, turbineID, reason, now)
	if err != nil {
		return model.Isolation{}, err
	}
	c.isolation = isolation
	c.hasSession = true
	return isolation, nil
}

func (c *Coordinator) AdvanceIsolation(next model.IsolationPhase, now time.Time) (model.Isolation, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.hasSession {
		return model.Isolation{}, fmt.Errorf("no isolation is active")
	}
	advanced, err := c.isolation.Advance(next, now)
	if err != nil {
		return model.Isolation{}, err
	}
	c.isolation = advanced
	return c.isolation, nil
}

func (c *Coordinator) Isolation() (model.Isolation, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.isolation, c.hasSession
}

func (c *Coordinator) RaiseIncident(incident model.Incident) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.incidents[incident.ID] = incident
}

func (c *Coordinator) Incidents() []model.Incident {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]model.Incident, 0, len(c.incidents))
	for _, incident := range c.incidents {
		result = append(result, incident)
	}
	return result
}
