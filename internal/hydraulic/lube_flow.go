package hydraulic

import (
	"fmt"
	"sync"
	"time"

	"github.com/wyw14/cry-110/internal/gearbox"
)

type LubeFlow struct {
	mu        sync.Mutex
	pressure  float64
	returnLPM float64
	tempC     float64
	at        time.Time
}

func (f *LubeFlow) Observe(pressureBar, returnLPM, sumpTempC float64, at time.Time) error {
	if pressureBar < 0 || returnLPM < 0 || at.IsZero() {
		return fmt.Errorf("invalid lubrication telemetry")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.pressure = pressureBar
	f.returnLPM = returnLPM
	f.tempC = sumpTempC
	f.at = at
	return nil
}

func (f *LubeFlow) WarmupProof(sessionID string) gearbox.WarmupProof {
	f.mu.Lock()
	defer f.mu.Unlock()
	return gearbox.WarmupProof{
		SessionID: sessionID, SumpTemperatureC: f.tempC,
		SupplyPressureBar: f.pressure, ReturnFlowLPM: f.returnLPM, ObservedAt: f.at,
	}
}

func (f *LubeFlow) Available(maxAge time.Duration, now time.Time) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.pressure > 0 && f.returnLPM > 0 && now.Sub(f.at) >= 0 && now.Sub(f.at) <= maxAge
}
