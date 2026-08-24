package rotor

import (
	"fmt"
	"sync"
)

type AeroState string

const (
	AeroProducing  AeroState = "producing"
	AeroFeathering AeroState = "feathering"
	AeroSafe       AeroState = "safe"
	AeroAsymmetric AeroState = "asymmetric"
)

type AeroController struct {
	mu     sync.Mutex
	state  AeroState
	unsafe []int
}

func NewAeroController() *AeroController {
	return &AeroController{state: AeroProducing}
}

func (a *AeroController) BeginFeather() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.state = AeroFeathering
	a.unsafe = nil
}

func (a *AeroController) UpdateBladeProofs(safe map[int]bool) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.state != AeroFeathering && a.state != AeroAsymmetric {
		return fmt.Errorf("rotor is not feathering")
	}
	a.unsafe = a.unsafe[:0]
	for blade := 1; blade <= 3; blade++ {
		if !safe[blade] {
			a.unsafe = append(a.unsafe, blade)
		}
	}
	if len(a.unsafe) == 0 {
		a.state = AeroSafe
	} else {
		a.state = AeroAsymmetric
	}
	return nil
}
