package rotor

import (
	"fmt"
	"sync"

	"github.com/wyw14/cry-110/internal/interlock"
)

type TestCycleState string

const (
	TestCycleIdle       TestCycleState = "idle"
	TestCycleRotating   TestCycleState = "rotating"
	TestCycleInspecting TestCycleState = "inspecting"
	TestCycleComplete   TestCycleState = "complete"
)

type TestCycle struct {
	mu        sync.Mutex
	exclusion *interlock.Exclusion
	state     TestCycleState
	session   string
}

func NewTestCycle(exclusion *interlock.Exclusion) *TestCycle {
	return &TestCycle{exclusion: exclusion, state: TestCycleIdle}
}

func (c *TestCycle) Start(session string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.state != TestCycleIdle && c.state != TestCycleComplete {
		return fmt.Errorf("a controlled rotor cycle is already active")
	}
	if err := c.exclusion.BeginShed(session); err != nil {
		return err
	}
	c.session = session
	c.state = TestCycleRotating
	return nil
}

func (c *TestCycle) StopForInspection() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.state != TestCycleRotating {
		return fmt.Errorf("controlled rotor cycle is not rotating")
	}
	if err := c.exclusion.BeginInspection(c.session); err != nil {
		return err
	}
	c.state = TestCycleInspecting
	return nil
}

func (c *TestCycle) Complete(noResidualIce bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.state != TestCycleInspecting {
		return fmt.Errorf("controlled rotor cycle has not reached inspection")
	}
	if err := c.exclusion.Release(c.session, noResidualIce); err != nil {
		return err
	}
	c.state = TestCycleComplete
	return nil
}

func (c *TestCycle) State() TestCycleState {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.state
}
