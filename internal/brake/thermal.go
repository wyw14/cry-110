package brake

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/wyw14/cry-110/internal/interlock"
)

type ThermalModel struct {
	mu           sync.Mutex
	energyKJ     float64
	limitKJ      float64
	ambientC     float64
	temperatureC float64
	coolingTau   time.Duration
	updatedAt    time.Time
}

func NewThermalModel(limitKJ, ambientC float64, coolingTau time.Duration, now time.Time) *ThermalModel {
	return &ThermalModel{limitKJ: limitKJ, ambientC: ambientC, temperatureC: ambientC, coolingTau: coolingTau, updatedAt: now}
}

func (m *ThermalModel) cool(now time.Time) {
	if !now.After(m.updatedAt) || m.coolingTau <= 0 {
		return
	}
	factor := math.Exp(-float64(now.Sub(m.updatedAt)) / float64(m.coolingTau))
	m.energyKJ *= factor
	m.temperatureC = m.ambientC + (m.temperatureC-m.ambientC)*factor
	m.updatedAt = now
}

func (m *ThermalModel) ApplyBraking(inertia, initialRPM, finalRPM float64, now time.Time) (interlock.ThermalState, error) {
	if inertia <= 0 || initialRPM < 0 || finalRPM < 0 || finalRPM > initialRPM {
		return interlock.ThermalState{}, fmt.Errorf("invalid braking energy inputs")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cool(now)
	initial := initialRPM * 2 * math.Pi / 60
	final := finalRPM * 2 * math.Pi / 60
	deltaKJ := 0.5 * inertia * (initial*initial - final*final) / 1000
	m.energyKJ += deltaKJ
	m.temperatureC += deltaKJ * 0.08
	return m.state(now), nil
}

func (m *ThermalModel) State(now time.Time) interlock.ThermalState {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cool(now)
	return m.state(now)
}

func (m *ThermalModel) state(now time.Time) interlock.ThermalState {
	return interlock.ThermalState{EnergyKJ: m.energyKJ, LimitKJ: m.limitKJ, Temperature: m.temperatureC, ObservedAt: now}
}

func (m *ThermalModel) ReleasePressure(now time.Time) interlock.ThermalState {
	return m.State(now)
}
