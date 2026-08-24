package verifycase

import (
	"testing"

	"github.com/wyw14/cry-110/internal/hydraulic"
	"github.com/wyw14/cry-110/internal/interlock"
	"github.com/wyw14/cry-110/internal/pitch"
)

func TestEmergencyPitchRequiresUsableAccumulatorVolume(t *testing.T) {
	budget := pitch.NewEnergyBudget(1.5)
	for blade := 1; blade <= 3; blade++ {
		if err := budget.SetDemand(pitch.BladeDemand{Blade: blade, RemainingAngle: 90, LitersPerDegree: 0.025}); err != nil {
			t.Fatal(err)
		}
	}
	service := hydraulic.NewService(hydraulic.Accumulator{
		NominalLiters: 16,
		PrechargeBar:  20,
		HeaderBar:     220,
		OilLiters:     12,
	}, 140, budget, interlock.NewRunPermit(140))
	if err := service.Ready(); err == nil {
		t.Fatal("normal header pressure was treated as proof of enough emergency pitch volume")
	}
}
