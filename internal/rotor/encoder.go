package rotor

import (
	"fmt"
	"math"
	"sync"
)

type Encoder struct {
	mu          sync.Mutex
	initialized bool
	lastRaw     float64
	continuous  float64
}

func NormalizeAngle(degrees float64) float64 {
	value := math.Mod(degrees, 360)
	if value < 0 {
		value += 360
	}
	return value
}

func ShortestAngleError(target, current float64) float64 {
	errorDegrees := NormalizeAngle(target) - NormalizeAngle(current)
	if errorDegrees > 180 {
		errorDegrees -= 360
	}
	if errorDegrees <= -180 {
		errorDegrees += 360
	}
	return errorDegrees
}

func (e *Encoder) Angle(raw float64) (float64, error) {
	if math.IsNaN(raw) || math.IsInf(raw, 0) {
		return 0, fmt.Errorf("encoder angle is not finite")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	raw = NormalizeAngle(raw)
	if !e.initialized {
		e.initialized = true
		e.lastRaw = raw
		e.continuous = raw
		return e.continuous, nil
	}
	delta := ShortestAngleError(raw, e.lastRaw)
	e.continuous += delta
	e.lastRaw = raw
	return e.continuous, nil
}

func (e *Encoder) Reset(raw float64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.initialized = true
	e.lastRaw = NormalizeAngle(raw)
	e.continuous = e.lastRaw
}
