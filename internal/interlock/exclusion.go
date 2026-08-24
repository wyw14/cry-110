package interlock

import (
	"fmt"
	"sync"
)

type ExclusionState string

const (
	ExclusionClear    ExclusionState = "clear"
	ExclusionLocked   ExclusionState = "locked"
	ExclusionShedding ExclusionState = "shedding"
	ExclusionInspect  ExclusionState = "inspection"
)

type Exclusion struct {
	mu      sync.Mutex
	state   ExclusionState
	session string
}

func NewExclusion() *Exclusion {
	return &Exclusion{state: ExclusionClear}
}

func (e *Exclusion) Lock(session string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if session == "" {
		return fmt.Errorf("hazard session is required")
	}
	if e.state != ExclusionClear && e.session != session {
		return fmt.Errorf("exclusion zone belongs to another hazard")
	}
	e.session = session
	e.state = ExclusionLocked
	return nil
}

func (e *Exclusion) BeginShed(session string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.session != session || e.state != ExclusionLocked {
		return fmt.Errorf("exclusion zone is not locked for this icing session")
	}
	e.state = ExclusionShedding
	return nil
}

func (e *Exclusion) BeginInspection(session string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.session != session || e.state != ExclusionShedding {
		return fmt.Errorf("controlled shed cycle is incomplete")
	}
	e.state = ExclusionInspect
	return nil
}

func (e *Exclusion) Release(session string, noResidualIce bool) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.session != session || e.state != ExclusionInspect {
		return fmt.Errorf("hazard inspection is incomplete")
	}
	if !noResidualIce {
		return fmt.Errorf("residual ice still blocks access")
	}
	e.session = ""
	e.state = ExclusionClear
	return nil
}

func (e *Exclusion) ReleaseOnStillness(session string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.session != session || e.state != ExclusionLocked {
		return fmt.Errorf("exclusion zone is not locked for this hazard")
	}
	e.session = ""
	e.state = ExclusionClear
	return nil
}

func (e *Exclusion) State() ExclusionState {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.state
}

func (e *Exclusion) AccessAllowed() bool {
	return e.State() == ExclusionClear
}
