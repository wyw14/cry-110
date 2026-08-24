package gearbox

import (
	"fmt"

	"github.com/wyw14/cry-110/internal/rotor"
)

type PhaseService struct {
	encoder *rotor.Encoder
	offset  float64
}

func NewPhaseService(encoder *rotor.Encoder, offset float64) *PhaseService {
	return &PhaseService{encoder: encoder, offset: offset}
}

func (s *PhaseService) Normalize(rawEncoderDegrees float64) (float64, error) {
	continuous, err := s.encoder.Angle(rawEncoderDegrees)
	if err != nil {
		return 0, err
	}
	return rotor.NormalizeAngle(continuous + s.offset), nil
}

func (s *PhaseService) Error(targetDegrees, rawEncoderDegrees float64) (float64, error) {
	current, err := s.Normalize(rawEncoderDegrees)
	if err != nil {
		return 0, fmt.Errorf("read shaft phase: %w", err)
	}
	return rotor.ShortestAngleError(targetDegrees, current), nil
}

func (s *PhaseService) Reset(rawEncoderDegrees float64) {
	s.encoder.Reset(rawEncoderDegrees)
}
