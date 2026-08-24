package gearbox

import (
	"testing"
	"time"
)

func TestWarmupRequiresAllThreeConditions(t *testing.T) {
	now := time.Now().UTC()
	w := NewWarmup(18, 2.5, 5)
	if err := w.Start("startup"); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Temperature met, but no supply pressure and no return flow -> must NOT be ready.
	state := w.Observe(WarmupProof{
		SessionID: "startup", SumpTemperatureC: 18,
		SupplyPressureBar: 0, ReturnFlowLPM: 0, ObservedAt: now,
	})
	if state == WarmupReady {
		t.Fatalf("ready with temperature only: state=%s", state)
	}

	// Temperature + supply pressure, but no return flow -> circulating, not ready.
	state = w.Observe(WarmupProof{
		SessionID: "startup", SumpTemperatureC: 18,
		SupplyPressureBar: 3, ReturnFlowLPM: 0, ObservedAt: now,
	})
	if state != WarmupCirculating {
		t.Fatalf("expected circulating, got %s", state)
	}

	// All three established -> ready.
	state = w.Observe(WarmupProof{
		SessionID: "startup", SumpTemperatureC: 18,
		SupplyPressureBar: 3, ReturnFlowLPM: 7, ObservedAt: now,
	})
	if state != WarmupReady {
		t.Fatalf("expected ready, got %s", state)
	}
}
