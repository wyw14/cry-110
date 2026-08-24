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
	return math.Min(a.OilLiters, a.NominalLiters), nil
}
