package pitch

import (
	"fmt"
	"sync"

	"github.com/wyw14/cry-110/internal/interlock"
)

type BladeDemand struct {
	Blade           int
	RemainingAngle  float64
	LitersPerDegree float64
}

func (d BladeDemand) RequiredLiters() float64 {
	if d.RemainingAngle <= 0 || d.LitersPerDegree <= 0 {
		return 0
	}
	return d.RemainingAngle * d.LitersPerDegree
}

type EnergyBudget struct {
	mu      sync.Mutex
	demands map[int]BladeDemand
	reserve float64
}

func NewEnergyBudget(reserveLiters float64) *EnergyBudget {
	return &EnergyBudget{demands: make(map[int]BladeDemand), reserve: reserveLiters}
}

func (b *EnergyBudget) SetDemand(demand BladeDemand) error {
	if demand.Blade < 1 || demand.Blade > 3 || demand.LitersPerDegree <= 0 {
		return fmt.Errorf("invalid blade energy demand")
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.demands[demand.Blade] = demand
	return nil
}

func (b *EnergyBudget) Required() (float64, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.demands) != 3 {
		return 0, fmt.Errorf("all three blade demands are required")
	}
	total := b.reserve
	for blade := 1; blade <= 3; blade++ {
		total += b.demands[blade].RequiredLiters()
	}
	return total, nil
}

func (b *EnergyBudget) Proof(usableLiters, pressureBar, prechargeBar float64) (interlock.EnergyProof, error) {
	required, err := b.Required()
	if err != nil {
		return interlock.EnergyProof{}, err
	}
	return interlock.EnergyProof{
		UsableLiters: usableLiters, RequiredLiters: required,
		PressureBar: pressureBar, PrechargeBar: prechargeBar,
	}, nil
}
