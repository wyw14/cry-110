package turninggear

import (
	"fmt"
	"sync"
	"time"

	"github.com/wyw14/cry-110/internal/gearbox"
)

type Permit struct {
	mu      sync.Mutex
	warmup  *gearbox.Warmup
	maxAge  time.Duration
	granted bool
}

func NewPermit(warmup *gearbox.Warmup, maxAge time.Duration) *Permit {
	return &Permit{warmup: warmup, maxAge: maxAge}
}

func (p *Permit) Evaluate(now time.Time) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	proof, state := p.warmup.Proof()
	if state != gearbox.WarmupReady {
		p.granted = false
		return fmt.Errorf("gearbox warmup is %s", state)
	}
	if now.Sub(proof.ObservedAt) < 0 || now.Sub(proof.ObservedAt) > p.maxAge {
		p.granted = false
		return fmt.Errorf("gearbox circulation proof is stale")
	}
	p.granted = true
	return nil
}

func (p *Permit) Granted() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.granted
}
