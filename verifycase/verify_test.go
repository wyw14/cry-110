package verifycase

import (
	"math"
	"testing"

	"github.com/wyw14/cry-110/internal/gearbox"
	"github.com/wyw14/cry-110/internal/rotor"
	"github.com/wyw14/cry-110/internal/turninggear"
)

func TestTurningGearPhaseRemainsContinuousAcrossWrap(t *testing.T) {
	encoder := &rotor.Encoder{}
	phase := gearbox.NewPhaseService(encoder, 0)
	controller := turninggear.NewController(phase, 0.5, 0.08)
	if command, err := controller.Command(359.8, 359.8); err != nil || command != 0 {
		t.Fatalf("initial command = %.4f, err = %v", command, err)
	}
	command, err := controller.Command(359.8, 0.1)
	if err != nil {
		t.Fatal(err)
	}
	if command >= 0 || math.Abs(command) > 0.1 {
		t.Fatalf("wrapped command = %.4f, want a small negative correction", command)
	}
}
