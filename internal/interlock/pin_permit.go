package interlock

import (
	"errors"
	"sync"
	"time"
)

type StillnessProof struct {
	StableSince time.Time
	ObservedAt  time.Time
	AverageRPM  float64
	InstantRPM  float64
	TorqueNm    float64
	Valid       bool
}

type PinState string

const (
	PinWithdrawn PinState = "withdrawn"
	PinApproach  PinState = "approach"
	PinInserted  PinState = "inserted"
)

type PinPermit struct {
	mu       sync.Mutex
	state    PinState
	proof    StillnessProof
	minDwell time.Duration
}

func NewPinPermit(minDwell time.Duration) *PinPermit {
	return &PinPermit{state: PinWithdrawn, minDwell: minDwell}
}

func (p *PinPermit) Update(proof StillnessProof) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.proof = proof
	if !p.validAt(proof.ObservedAt) && p.state == PinApproach {
		p.state = PinWithdrawn
	}
}

func (p *PinPermit) Begin(now time.Time) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.state != PinWithdrawn {
		return errors.New("mechanical pin is not withdrawn")
	}
	if !p.validAt(now) {
		return errors.New("continuous rotor stillness proof is unavailable")
	}
	p.state = PinApproach
	return nil
}

func (p *PinPermit) Complete(now time.Time) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.state != PinApproach {
		return errors.New("mechanical pin is not approaching")
	}
	if !p.validAt(now) {
		p.state = PinWithdrawn
		return errors.New("rotor moved while the mechanical pin approached")
	}
	p.state = PinInserted
	return nil
}

func (p *PinPermit) Withdraw() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state = PinWithdrawn
}

func (p *PinPermit) State() PinState {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.state
}

func (p *PinPermit) validAt(now time.Time) bool {
	return p.proof.Valid &&
		!p.proof.StableSince.IsZero() &&
		now.Sub(p.proof.StableSince) >= p.minDwell &&
		now.Sub(p.proof.ObservedAt) >= 0 &&
		now.Sub(p.proof.ObservedAt) <= time.Second
}
