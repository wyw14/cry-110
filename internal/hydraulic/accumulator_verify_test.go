package hydraulic

import (
	"math"
	"testing"

	"github.com/wyw14/cry-110/internal/interlock"
	"github.com/wyw14/cry-110/internal/pitch"
)

// requiredForEmergencyStroke mirrors the default System wiring so the proof
// path is exercised against realistic demand figures.
func requiredForEmergencyStroke(t *testing.T) *pitch.EnergyBudget {
	t.Helper()
	budget := pitch.NewEnergyBudget(1.5)
	for blade := 1; blade <= 3; blade++ {
		if err := budget.SetDemand(pitch.BladeDemand{
			Blade: blade, RemainingAngle: 90, LitersPerDegree: 0.025,
		}); err != nil {
			t.Fatalf("set demand: %v", err)
		}
	}
	return budget
}

func TestUsableVolumeHonorsPrecharge(t *testing.T) {
	// Healthy precharge: the gas cushion holds pressure across the stroke.
	healthy := Accumulator{NominalLiters: 16, PrechargeBar: 90, HeaderBar: 220, OilLiters: 12}
	usable, err := healthy.UsableVolume(140)
	if err != nil {
		t.Fatalf("healthy usable: %v", err)
	}
	// Nominal*P0*(1/min - 1/header) = 16*90*(1/140 - 1/220) ~= 3.58, capped at 12.
	want := 16 * 90 * (1.0/140 - 1.0/220)
	if math.Abs(usable-want) > 1e-9 {
		t.Fatalf("healthy usable = %.4f, want %.4f", usable, want)
	}

	// The described fault: gauge reads normal but the precharge has collapsed.
	// The bottle can no longer hold pressure across the stroke, so the usable
	// reserve collapses far below what the static oil level (12 L) implies.
	degraded := Accumulator{NominalLiters: 16, PrechargeBar: 35, HeaderBar: 220, OilLiters: 12}
	degradedUsable, err := degraded.UsableVolume(140)
	if err != nil {
		t.Fatalf("degraded usable: %v", err)
	}
	if !(degradedUsable < usable) {
		t.Fatalf("degraded precharge should reduce usable, got %.4f >= %.4f", degradedUsable, usable)
	}
	// It must be far below the 8.25 L the emergency stroke needs.
	if degradedUsable > 8.25 {
		t.Fatalf("degraded usable = %.4f, expected it well below the 8.25 L stroke demand", degradedUsable)
	}
}

func TestEmergencyPitchReadyRejectsLostPrecharge(t *testing.T) {
	budget := requiredForEmergencyStroke(t)

	// Instantaneous pressure is normal, but precharge has collapsed: the old
	// implementation treated the oil level as available capacity and reported
	// ready. The proof must instead reflect that the stroke cannot complete.
	lost := Accumulator{NominalLiters: 16, PrechargeBar: 35, HeaderBar: 220, OilLiters: 12}
	usable, err := lost.UsableVolume(140)
	if err != nil {
		t.Fatalf("usable: %v", err)
	}
	proof, err := budget.Proof(usable, lost.HeaderBar, lost.PrechargeBar)
	if err != nil {
		t.Fatalf("proof: %v", err)
	}
	permit := interlock.NewRunPermit(140)
	if err := permit.EmergencyPitchReady(proof); err == nil {
		t.Fatalf("emergency pitch ready must be denied when precharge cannot sustain the stroke")
	}
}
