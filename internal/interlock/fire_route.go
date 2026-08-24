package interlock

import "fmt"

type DamperState string

const (
	DamperClosed DamperState = "closed"
	DamperOpen   DamperState = "open"
)

type FireRoute struct {
	OutsideIntake DamperState
	CabinetIntake DamperState
	SafeExhaust   DamperState
	FanDirection  string
	Latched       bool
}

func NewFireRoute() FireRoute {
	return FireRoute{OutsideIntake: DamperOpen, CabinetIntake: DamperOpen, SafeExhaust: DamperClosed, FanDirection: "supply"}
}

func (r FireRoute) Apply() FireRoute {
	r.OutsideIntake = DamperClosed
	r.CabinetIntake = DamperClosed
	r.SafeExhaust = DamperOpen
	r.FanDirection = "exhaust"
	r.Latched = true
	return r
}

func (r FireRoute) ValidatePurge() error {
	if !r.Latched {
		return fmt.Errorf("fire route is not latched")
	}
	if r.OutsideIntake != DamperClosed || r.CabinetIntake != DamperClosed {
		return fmt.Errorf("fire route still supplies outside air")
	}
	if r.SafeExhaust != DamperOpen || r.FanDirection != "exhaust" {
		return fmt.Errorf("fire route has no isolated exhaust path")
	}
	return nil
}
