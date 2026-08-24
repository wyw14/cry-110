package hydraulic

import (
	"fmt"
	"math"
)

type Accumulator struct {
	NominalLiters float64
	PrechargeBar  float64
	HeaderBar     float64
	OilLiters     float64
}

// UsableVolume returns the oil that can be expelled while the gas side stays
// above minimumPressureBar. The instantaneous header pressure only proves the
// system is charged *now*; it says nothing about whether that charge can be
// sustained across the full emergency stroke. Precharge is the gas cushion that
// holds the bottle off the empty state, so the usable reserve is the oil released
// as pressure falls from the header to the minimum operating pressure, bounded by
// the oil that is physically present.
//
// Gas side follows the isothermal relation P0·V0 = P·Vg where V0 = NominalLiters
// (the gas fills the whole bottle at precharge). Oil present at pressure P is
// NominalLiters·(1 − PrechargeBar/P), so the oil expelled between header and
// minimum pressure is NominalLiters·PrechargeBar·(1/minimumPressureBar −
// 1/HeaderBar). A low or failed precharge collapses this term and correctly
// reports a short reserve even when the gauge still reads normal.
func (a Accumulator) UsableVolume(minimumPressureBar float64) (float64, error) {
	if a.NominalLiters <= 0 || a.OilLiters < 0 {
		return 0, fmt.Errorf("accumulator volume is invalid")
	}
	if a.PrechargeBar <= 0 || a.HeaderBar <= a.PrechargeBar {
		return 0, fmt.Errorf("accumulator precharge is invalid")
	}
	if minimumPressureBar <= a.PrechargeBar || minimumPressureBar >= a.HeaderBar {
		return 0, fmt.Errorf("minimum operating pressure is outside the discharge range")
	}
	oilAtHeader := a.NominalLiters * (1 - a.PrechargeBar/a.HeaderBar)
	oilAtMinimum := a.NominalLiters * (1 - a.PrechargeBar/minimumPressureBar)
	dischargeable := oilAtHeader - oilAtMinimum
	if dischargeable < 0 {
		dischargeable = 0
	}
	return math.Min(dischargeable, a.OilLiters), nil
}
