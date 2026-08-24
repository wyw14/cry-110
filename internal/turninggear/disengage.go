package turninggear

import (
	"fmt"
	"sync"
	"time"

	"github.com/wyw14/cry-110/internal/interlock"
)

type DisengageState string

const (
	DisengageMeshed     DisengageState = "meshed"
	DisengageUnloading  DisengageState = "unloading"
	DisengageRetracting DisengageState = "retracting"
	DisengageComplete   DisengageState = "complete"
	DisengageFailed     DisengageState = "failed"
)

type Disengage struct {
	mu          sync.Mutex
	operationID string
	state       DisengageState
	maximumIdle float64
	permit      *interlock.RotorPermit
}

func NewDisengage(maximumIdleCurrent float64, permit *interlock.RotorPermit) *Disengage {
	return &Disengage{state: DisengageMeshed, maximumIdle: maximumIdleCurrent, permit: permit}
}

func (d *Disengage) Start(operationID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if operationID == "" {
		return fmt.Errorf("disengage operation is required")
	}
	d.operationID = operationID
	d.state = DisengageUnloading
	return nil
}

func (d *Disengage) ApplyCurrent(amperes float64, pawlState string, now time.Time) DisengageState {
	d.mu.Lock()
	defer d.mu.Unlock()
	motorIdle := amperes <= d.maximumIdle
	if !motorIdle {
		d.state = DisengageUnloading
	} else if pawlState != "retracted" {
		d.state = DisengageRetracting
	} else {
		d.state = DisengageComplete
	}
	d.permit.Update(interlock.DisengageProof{
		OperationID: d.operationID, MotorIdle: motorIdle, PawlState: pawlState, ObservedAt: now,
	})
	return d.state
}

func (d *Disengage) State() DisengageState {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.state
}
