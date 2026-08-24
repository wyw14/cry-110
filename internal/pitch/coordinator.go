package pitch

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/wyw14/cry-110/internal/model"
	"github.com/wyw14/cry-110/internal/rotor"
)

type GroupState string

const (
	GroupIdle     GroupState = "idle"
	GroupMoving   GroupState = "moving"
	GroupSafe     GroupState = "safe"
	GroupDegraded GroupState = "degraded"
)

type Coordinator struct {
	mu              sync.Mutex
	sessionID       string
	state           GroupState
	blades          map[int]model.BladeTelemetry
	minimumAngle    float64
	minimumPressure float64
	maximumAge      time.Duration
	aero            *rotor.AeroController
}

func NewCoordinator(minimumAngle, minimumPressure float64, maximumAge time.Duration, aero *rotor.AeroController) *Coordinator {
	return &Coordinator{
		state: GroupIdle, blades: make(map[int]model.BladeTelemetry),
		minimumAngle: minimumAngle, minimumPressure: minimumPressure,
		maximumAge: maximumAge, aero: aero,
	}
}

func (c *Coordinator) Begin(sessionID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if sessionID == "" {
		return fmt.Errorf("pitch session is required")
	}
	c.sessionID = sessionID
	c.blades = make(map[int]model.BladeTelemetry)
	c.state = GroupMoving
	c.aero.BeginFeather()
	return nil
}

func (c *Coordinator) Observe(sessionID string, blade model.BladeTelemetry, now time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.state != GroupMoving && c.state != GroupDegraded {
		return fmt.Errorf("pitch group is not moving")
	}
	if sessionID != c.sessionID {
		return fmt.Errorf("blade proof belongs to another pitch session")
	}
	if blade.Blade < 1 || blade.Blade > 3 {
		return fmt.Errorf("blade number must be between one and three")
	}
	c.blades[blade.Blade] = blade
	safe := make(map[int]bool, 3)
	for id := 1; id <= 3; id++ {
		telemetry, exists := c.blades[id]
		safe[id] = exists && telemetry.Safe(c.minimumAngle, c.minimumPressure, c.maximumAge, now)
	}
	if err := c.aero.UpdateBladeProofs(safe); err != nil {
		return err
	}
	if safe[1] && safe[2] && safe[3] {
		c.state = GroupSafe
	} else {
		c.state = GroupDegraded
	}
	return nil
}

func (c *Coordinator) Status(now time.Time) (GroupState, []int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	unsafe := make([]int, 0, 3)
	for id := 1; id <= 3; id++ {
		blade, exists := c.blades[id]
		if !exists || !blade.Safe(c.minimumAngle, c.minimumPressure, c.maximumAge, now) {
			unsafe = append(unsafe, id)
		}
	}
	sort.Ints(unsafe)
	return c.state, unsafe
}
