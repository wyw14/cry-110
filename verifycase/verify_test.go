package verifycase

import (
	"testing"
	"time"

	"github.com/wyw14/cry-110/internal/brake"
	"github.com/wyw14/cry-110/internal/interlock"
	"github.com/wyw14/cry-110/internal/turninggear"
)

func TestTurningGearDisengageRequiresPawlRetracted(t *testing.T) {
	now := time.Now().UTC()
	permit := interlock.NewRotorPermit(time.Second)
	disengage := turninggear.NewDisengage(1.2, permit)
	if err := disengage.Start("turning-session"); err != nil {
		t.Fatal(err)
	}
	if state := disengage.ApplyCurrent(0.5, "meshed", now); state == turninggear.DisengageComplete {
		t.Fatal("turning gear completed disengagement with the pawl still meshed")
	}
	mainShaftBrake := brake.NewRelease(permit)
	if err := mainShaftBrake.Release("turning-session", now); err == nil {
		t.Fatal("main shaft brake released without a retracted-pawl proof")
	}
}
