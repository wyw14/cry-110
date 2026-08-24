package interlock

import (
	"fmt"
	"time"
)

type ThermalState struct {
	EnergyKJ    float64
	LimitKJ     float64
	Temperature float64
	ObservedAt  time.Time
}

type YawLimit struct {
	warnFraction float64
}

func NewYawLimit(warnFraction float64) YawLimit {
	if warnFraction <= 0 || warnFraction >= 1 {
		warnFraction = 0.8
	}
	return YawLimit{warnFraction: warnFraction}
}

func (l YawLimit) AllowedSpeed(requested float64, thermal ThermalState) (float64, error) {
	if requested <= 0 {
		return 0, fmt.Errorf("yaw speed must be positive")
	}
	if thermal.LimitKJ <= 0 || thermal.ObservedAt.IsZero() {
		return 0, fmt.Errorf("yaw brake thermal proof is unavailable")
	}
	fraction := thermal.EnergyKJ / thermal.LimitKJ
	if fraction >= 1 {
		return 0, fmt.Errorf("yaw brake thermal capacity exhausted")
	}
	if fraction >= l.warnFraction {
		remaining := (1 - fraction) / (1 - l.warnFraction)
		return requested * remaining, nil
	}
	return requested, nil
}

func (l YawLimit) Ready(thermal ThermalState) bool {
	return thermal.LimitKJ > 0 && thermal.EnergyKJ < thermal.LimitKJ
}
