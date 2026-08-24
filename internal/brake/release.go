package brake

import (
	"fmt"
	"sync"
	"time"

	"github.com/wyw14/cry-110/internal/interlock"
)

type Release struct {
	mu       sync.Mutex
	permit   *interlock.RotorPermit
	applied  bool
	released bool
}

func NewRelease(permit *interlock.RotorPermit) *Release {
	return &Release{permit: permit, applied: true}
}

func (r *Release) Apply() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.applied = true
	r.released = false
}

func (r *Release) Release(operationID string, now time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.applied {
		return fmt.Errorf("main shaft brake is not applied")
	}
	if err := r.permit.Allow(operationID, now); err != nil {
		return err
	}
	r.applied = false
	r.released = true
	return nil
}

func (r *Release) State() (applied, released bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.applied, r.released
}
