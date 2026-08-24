package crane

import (
	"fmt"
	"sync"
	"time"
)

type ParkState string

const (
	ParkDeployed  ParkState = "deployed"
	ParkReturning ParkState = "returning"
	ParkSecured   ParkState = "secured"
)

type Park struct {
	mu        sync.Mutex
	state     ParkState
	hookM     float64
	boomAngle float64
	updatedAt time.Time
}

func NewPark() *Park {
	return &Park{state: ParkSecured}
}

func (p *Park) Deploy(hookMeters, boomAngle float64, now time.Time) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if hookMeters < 0 {
		return fmt.Errorf("hook extension must not be negative")
	}
	p.hookM = hookMeters
	p.boomAngle = boomAngle
	p.updatedAt = now
	p.state = ParkDeployed
	return nil
}

func (p *Park) BeginPark(now time.Time) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state = ParkReturning
	p.updatedAt = now
}

func (p *Park) ConfirmSecured(hookMeters, boomAngle float64, now time.Time) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if hookMeters > 0.2 || boomAngle < -1 || boomAngle > 1 {
		return fmt.Errorf("crane is outside the secured envelope")
	}
	p.hookM = hookMeters
	p.boomAngle = boomAngle
	p.updatedAt = now
	p.state = ParkSecured
	return nil
}

func (p *Park) Status() (ParkState, float64, float64, time.Time) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.state, p.hookM, p.boomAngle, p.updatedAt
}
