package gearbox

import (
	"fmt"
	"sync"
	"time"
)

type WarmupProof struct {
	SessionID         string
	SumpTemperatureC  float64
	SupplyPressureBar float64
	ReturnFlowLPM     float64
	ObservedAt        time.Time
}

type WarmupState string

const (
	WarmupCold        WarmupState = "cold"
	WarmupHeating     WarmupState = "heating"
	WarmupCirculating WarmupState = "circulating"
	WarmupReady       WarmupState = "ready"
	WarmupFailed      WarmupState = "failed"
)

type Warmup struct {
	mu          sync.Mutex
	state       WarmupState
	minimumTemp float64
	minimumPSI  float64
	minimumFlow float64
	proof       WarmupProof
}

func NewWarmup(minimumTemp, minimumPressure, minimumFlow float64) *Warmup {
	return &Warmup{state: WarmupCold, minimumTemp: minimumTemp, minimumPSI: minimumPressure, minimumFlow: minimumFlow}
}

func (w *Warmup) Start(sessionID string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if sessionID == "" {
		return fmt.Errorf("warmup session is required")
	}
	w.state = WarmupHeating
	w.proof = WarmupProof{SessionID: sessionID}
	return nil
}

func (w *Warmup) Observe(proof WarmupProof) WarmupState {
	w.mu.Lock()
	defer w.mu.Unlock()
	if proof.SessionID == "" || proof.SessionID != w.proof.SessionID || proof.ObservedAt.IsZero() {
		w.state = WarmupFailed
		return w.state
	}
	w.proof = proof
	if proof.SumpTemperatureC < w.minimumTemp {
		w.state = WarmupHeating
		return w.state
	}
	// Oil temperature alone is not sufficient. The sump heater can lift
	// temperature past the threshold before the lube pump has primed the
	// supply header and opened the return circuit, which leaves the planet
	// carrier bearings starved the moment the gearbox takes load. Demand
	// both supply pressure and a measurable return flow before ready.
	if proof.SupplyPressureBar < w.minimumPSI {
		w.state = WarmupHeating
		return w.state
	}
	if proof.ReturnFlowLPM < w.minimumFlow {
		w.state = WarmupCirculating
		return w.state
	}
	w.state = WarmupReady
	return w.state
}

func (w *Warmup) Proof() (WarmupProof, WarmupState) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.proof, w.state
}
