package verifycase

import (
	"testing"
	"time"

	"github.com/wyw14/cry-110/internal/gearbox"
	"github.com/wyw14/cry-110/internal/hydraulic"
	"github.com/wyw14/cry-110/internal/turninggear"
)

func TestGearboxWarmupRequiresOilCirculation(t *testing.T) {
	now := time.Now().UTC()
	warmup := gearbox.NewWarmup(18, 2.5, 5)
	if err := warmup.Start("cold-start"); err != nil {
		t.Fatal(err)
	}
	flow := &hydraulic.LubeFlow{}
	if err := flow.Observe(0, 0, 18, now); err != nil {
		t.Fatal(err)
	}
	if state := warmup.Observe(flow.WarmupProof("cold-start")); state == gearbox.WarmupReady {
		t.Fatal("warmup became ready without supply pressure or return flow")
	}
	permit := turninggear.NewPermit(warmup, time.Second)
	if err := permit.Evaluate(now); err == nil {
		t.Fatal("turning gear received a permit without established oil circulation")
	}
}
