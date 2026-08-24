package turninggear

import (
	"fmt"
	"math"

	"github.com/wyw14/cry-110/internal/gearbox"
)

type Controller struct {
	phase        *gearbox.PhaseService
	maximumRPM   float64
	proportional float64
}

func NewController(phase *gearbox.PhaseService, maximumRPM, proportional float64) *Controller {
	return &Controller{phase: phase, maximumRPM: maximumRPM, proportional: proportional}
}

func (c *Controller) Error(target, rawEncoder float64) (float64, error) {
	current, err := c.phase.Normalize(rawEncoder)
	if err != nil {
		return 0, err
	}
	return target - current, nil
}

func (c *Controller) Command(target, rawEncoder float64) (float64, error) {
	errorDegrees, err := c.Error(target, rawEncoder)
	if err != nil {
		return 0, err
	}
	if math.Abs(errorDegrees) < 0.05 {
		return 0, nil
	}
	command := errorDegrees * c.proportional
	if command > c.maximumRPM {
		command = c.maximumRPM
	}
	if command < -c.maximumRPM {
		command = -c.maximumRPM
	}
	if math.IsNaN(command) || math.IsInf(command, 0) {
		return 0, fmt.Errorf("turning gear command is not finite")
	}
	return command, nil
}
