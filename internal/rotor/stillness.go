package rotor

import (
	"math"
	"sync"
	"time"

	"github.com/wyw14/cry-110/internal/interlock"
	"github.com/wyw14/cry-110/internal/model"
)

type StillnessConfig struct {
	MaximumAverageRPM float64
	MaximumInstantRPM float64
	MaximumTorqueNm   float64
	RequiredDuration  time.Duration
	MaximumGap        time.Duration
}

type StillnessObserver struct {
	mu          sync.Mutex
	config      StillnessConfig
	stableSince time.Time
	last        model.RotorTelemetry
}

func NewStillnessObserver(config StillnessConfig) *StillnessObserver {
	return &StillnessObserver{config: config}
}

func (o *StillnessObserver) Observe(sample model.RotorTelemetry) interlock.StillnessProof {
	o.mu.Lock()
	defer o.mu.Unlock()
	// The proof must not rely on the averaged RPM alone: a torsional rebound
	// can spike the instantaneous RPM and shaft torque while the time-averaged
	// RPM stays near zero. All three conditions — sustained low speed, low
	// instantaneous speed, and a quiescent shaft torque — must hold on every
	// sample, otherwise the RequiredDuration continuity below is reset.
	valid := sample.Validate() == nil &&
		math.Abs(sample.AverageRPM) <= o.config.MaximumAverageRPM &&
		math.Abs(sample.InstantRPM) <= o.config.MaximumInstantRPM &&
		math.Abs(sample.ShaftTorqueNm) <= o.config.MaximumTorqueNm
	if !o.last.ObservedAt.IsZero() && sample.ObservedAt.Sub(o.last.ObservedAt) > o.config.MaximumGap {
		valid = false
	}
	if valid {
		if o.stableSince.IsZero() {
			o.stableSince = sample.ObservedAt
		}
	} else {
		o.stableSince = time.Time{}
	}
	o.last = sample
	return interlock.StillnessProof{
		StableSince: o.stableSince,
		ObservedAt:  sample.ObservedAt,
		AverageRPM:  sample.AverageRPM,
		InstantRPM:  sample.InstantRPM,
		TorqueNm:    sample.ShaftTorqueNm,
		Valid:       valid && !o.stableSince.IsZero() && sample.ObservedAt.Sub(o.stableSince) >= o.config.RequiredDuration,
	}
}

func (o *StillnessObserver) Reset() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.stableSince = time.Time{}
	o.last = model.RotorTelemetry{}
}
