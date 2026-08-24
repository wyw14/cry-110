package model

import (
	"fmt"
	"math"
	"time"
)

type RotorTelemetry struct {
	AverageRPM     float64   `json:"average_rpm"`
	InstantRPM     float64   `json:"instant_rpm"`
	ShaftTorqueNm  float64   `json:"shaft_torque_nm"`
	EncoderDegrees float64   `json:"encoder_degrees"`
	ObservedAt     time.Time `json:"observed_at"`
}

func (t RotorTelemetry) Validate() error {
	values := []float64{t.AverageRPM, t.InstantRPM, t.ShaftTorqueNm, t.EncoderDegrees}
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return fmt.Errorf("rotor telemetry contains a non-finite value")
		}
	}
	if t.ObservedAt.IsZero() {
		return fmt.Errorf("rotor telemetry timestamp is required")
	}
	return nil
}

type BladeTelemetry struct {
	Blade        int       `json:"blade"`
	AngleDegrees float64   `json:"angle_degrees"`
	PressureBar  float64   `json:"pressure_bar"`
	ObservedAt   time.Time `json:"observed_at"`
}

func (b BladeTelemetry) Safe(minAngle, minPressure float64, maxAge time.Duration, now time.Time) bool {
	return b.Blade >= 1 && b.Blade <= 3 &&
		b.AngleDegrees >= minAngle &&
		b.PressureBar >= minPressure &&
		!b.ObservedAt.IsZero() &&
		now.Sub(b.ObservedAt) >= 0 && now.Sub(b.ObservedAt) <= maxAge
}

type MotionTarget struct {
	OperationID string  `json:"operation_id"`
	Axis        string  `json:"axis"`
	Start       float64 `json:"start"`
	End         float64 `json:"end"`
	Seconds     float64 `json:"seconds"`
}

func (m MotionTarget) Validate() error {
	if m.OperationID == "" || m.Axis == "" {
		return fmt.Errorf("motion identity is required")
	}
	if m.Seconds <= 0 || math.IsNaN(m.Seconds) || math.IsInf(m.Seconds, 0) {
		return fmt.Errorf("motion duration must be positive")
	}
	return nil
}
