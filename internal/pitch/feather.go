package pitch

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/wyw14/cry-110/internal/model"
)

type Actuator struct {
	mu       sync.Mutex
	blade    int
	angle    float64
	pressure float64
}

func NewActuator(blade int, angle float64) (*Actuator, error) {
	if blade < 1 || blade > 3 {
		return nil, fmt.Errorf("blade number must be between one and three")
	}
	return &Actuator{blade: blade, angle: angle}, nil
}

func (a *Actuator) MoveToward(target, maximumStep, pressure float64, now time.Time) model.BladeTelemetry {
	a.mu.Lock()
	defer a.mu.Unlock()
	delta := target - a.angle
	if math.Abs(delta) > maximumStep {
		delta = math.Copysign(maximumStep, delta)
	}
	if pressure > 0 {
		a.angle += delta
	}
	a.pressure = pressure
	return model.BladeTelemetry{Blade: a.blade, AngleDegrees: a.angle, PressureBar: pressure, ObservedAt: now}
}

func (a *Actuator) Telemetry(now time.Time) model.BladeTelemetry {
	a.mu.Lock()
	defer a.mu.Unlock()
	return model.BladeTelemetry{Blade: a.blade, AngleDegrees: a.angle, PressureBar: a.pressure, ObservedAt: now}
}
