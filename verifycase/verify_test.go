package verifycase

import (
	"testing"
	"time"

	"github.com/wyw14/cry-110/internal/fire"
	"github.com/wyw14/cry-110/internal/interlock"
	"github.com/wyw14/cry-110/internal/ventilation"
)

func TestFirePurgeBuildsIsolatedExhaustRoute(t *testing.T) {
	cooling := ventilation.NewCoolingService()
	if err := cooling.StartCooling(80); err != nil {
		t.Fatal(err)
	}
	response := fire.NewResponse(cooling)
	if _, err := response.StartPurge("converter cabinet fire", 180, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	mode, flow, route := cooling.Status()
	if mode != ventilation.ModePurge || flow != 180 {
		t.Fatalf("ventilation mode = %s, flow = %.1f", mode, flow)
	}
	if err := route.ValidatePurge(); err != nil {
		t.Fatalf("invalid fire purge route: %v", err)
	}
	if route.OutsideIntake != interlock.DamperClosed || route.SafeExhaust != interlock.DamperOpen {
		t.Fatalf("fire purge route keeps an unsafe air path: %+v", route)
	}
}
