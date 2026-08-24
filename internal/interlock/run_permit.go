package interlock

import (
	"errors"
	"fmt"
)

type EnergyProof struct {
	UsableLiters   float64
	RequiredLiters float64
	PressureBar    float64
	PrechargeBar   float64
}

func (p EnergyProof) Deficit() float64 {
	if p.RequiredLiters <= p.UsableLiters {
		return 0
	}
	return p.RequiredLiters - p.UsableLiters
}

type RunPermit struct {
	minimumPressure float64
}

func NewRunPermit(minimumPressure float64) RunPermit {
	return RunPermit{minimumPressure: minimumPressure}
}

func (p RunPermit) EmergencyPitchReady(proof EnergyProof) error {
	if proof.RequiredLiters <= 0 {
		return errors.New("emergency pitch requirement is unavailable")
	}
	if proof.PressureBar < p.minimumPressure {
		return fmt.Errorf("hydraulic pressure %.1f bar is below %.1f bar", proof.PressureBar, p.minimumPressure)
	}
	if proof.PrechargeBar <= 0 || proof.PrechargeBar >= proof.PressureBar {
		return errors.New("accumulator precharge proof is invalid")
	}
	if deficit := proof.Deficit(); deficit > 0 {
		return fmt.Errorf("emergency pitch energy is short by %.2f liters", deficit)
	}
	return nil
}
