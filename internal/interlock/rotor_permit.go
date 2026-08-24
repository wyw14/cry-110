package interlock

import (
	"fmt"
	"sync"
	"time"
)

type DisengageProof struct {
	OperationID string
	MotorIdle   bool
	PawlState   string
	ObservedAt  time.Time
}

type RotorPermit struct {
	mu      sync.Mutex
	proof   DisengageProof
	maxAge  time.Duration
	blocked string
}

func NewRotorPermit(maxAge time.Duration) *RotorPermit {
	return &RotorPermit{maxAge: maxAge}
}

func (p *RotorPermit) Update(proof DisengageProof) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.proof = proof
	p.blocked = ""
	if !proof.MotorIdle {
		p.blocked = "turning gear motor remains loaded"
	}
}

func (p *RotorPermit) Allow(operationID string, now time.Time) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if operationID == "" || p.proof.OperationID != operationID {
		return fmt.Errorf("turning gear proof belongs to another operation")
	}
	if p.blocked != "" {
		return fmt.Errorf("rotor motion blocked: %s", p.blocked)
	}
	if now.Sub(p.proof.ObservedAt) < 0 || now.Sub(p.proof.ObservedAt) > p.maxAge {
		return fmt.Errorf("turning gear proof is stale")
	}
	return nil
}

func (p *RotorPermit) BlockReason() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.blocked
}
