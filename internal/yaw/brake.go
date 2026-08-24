package yaw

import (
	"fmt"
	"sync"
	"time"

	"github.com/wyw14/cry-110/internal/brake"
)

type BrakeController struct {
	mu      sync.Mutex
	service *brake.Service
	locked  bool
}

func NewBrakeController(service *brake.Service) *BrakeController {
	return &BrakeController{service: service}
}

func (c *BrakeController) Lock(inertia, speedRPM float64, now time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.service.Apply(inertia, speedRPM, 0, now); err != nil {
		return fmt.Errorf("lock nacelle yaw: %w", err)
	}
	c.locked = true
	return nil
}

func (c *BrakeController) Unlock(now time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.service.ReleasePressure(now)
	c.locked = false
}

func (c *BrakeController) Locked() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.locked
}
