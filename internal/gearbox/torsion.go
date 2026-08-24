package gearbox

import (
	"math"
	"sync"
	"time"
)

type TorsionSample struct {
	TorqueNm   float64
	InstantRPM float64
	At         time.Time
}

type TorsionState struct {
	PeakTorqueNm float64
	PeakRPM      float64
	Stable       bool
	ObservedAt   time.Time
}

type TorsionAnalyzer struct {
	mu              sync.Mutex
	maximumTorqueNm float64
	maximumRPM      float64
	window          time.Duration
	samples         []TorsionSample
}

func NewTorsionAnalyzer(maximumTorqueNm, maximumRPM float64, window time.Duration) *TorsionAnalyzer {
	return &TorsionAnalyzer{maximumTorqueNm: maximumTorqueNm, maximumRPM: maximumRPM, window: window}
}

func (a *TorsionAnalyzer) Observe(sample TorsionSample) TorsionState {
	a.mu.Lock()
	defer a.mu.Unlock()
	cutoff := sample.At.Add(-a.window)
	kept := a.samples[:0]
	for _, existing := range a.samples {
		if !existing.At.Before(cutoff) {
			kept = append(kept, existing)
		}
	}
	a.samples = append(kept, sample)
	state := TorsionState{Stable: len(a.samples) > 1, ObservedAt: sample.At}
	for _, existing := range a.samples {
		state.PeakTorqueNm = math.Max(state.PeakTorqueNm, math.Abs(existing.TorqueNm))
		state.PeakRPM = math.Max(state.PeakRPM, math.Abs(existing.InstantRPM))
	}
	state.Stable = state.Stable && state.PeakTorqueNm <= a.maximumTorqueNm && state.PeakRPM <= a.maximumRPM
	return state
}

func (a *TorsionAnalyzer) Reset() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.samples = nil
}
