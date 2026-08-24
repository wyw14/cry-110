package api

import (
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/wyw14/cry-110/internal/model"
)

func (s *Server) getIsolation(writer http.ResponseWriter, request *http.Request) {
	isolation, exists := s.system.Coordinator.Isolation()
	if !exists {
		snapshot, err := s.system.Snapshots.Load()
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			respondError(writer, http.StatusInternalServerError, err)
			return
		}
		respondJSON(writer, http.StatusOK, map[string]any{"active": false, "last_snapshot": snapshot})
		return
	}
	respondJSON(writer, http.StatusOK, map[string]any{
		"active": !isolation.Terminal(), "isolation": isolation,
		"pin_state": s.system.Isolation.PinState(), "brake_sequence": s.system.BrakeSequence.State(),
	})
}

func (s *Server) advanceIsolation(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		Action         string  `json:"action"`
		AverageRPM     float64 `json:"average_rpm"`
		InstantRPM     float64 `json:"instant_rpm"`
		ShaftTorqueNm  float64 `json:"shaft_torque_nm"`
		EncoderDegrees float64 `json:"encoder_degrees"`
	}
	if err := decodeJSON(request, &input); err != nil {
		respondError(writer, http.StatusBadRequest, err)
		return
	}
	now := time.Now().UTC()
	var isolation model.Isolation
	var err error
	switch input.Action {
	case "coasting":
		isolation, err = s.system.Isolation.ConfirmCoasting(model.RotorTelemetry{
			AverageRPM: input.AverageRPM, InstantRPM: input.InstantRPM,
			ShaftTorqueNm: input.ShaftTorqueNm, EncoderDegrees: input.EncoderDegrees, ObservedAt: now,
		}, now)
	case "feather":
		isolation, err = s.system.Isolation.Feather(now)
	case "brake":
		isolation, err = s.system.Isolation.Brake(now)
	case "pin":
		isolation, err = s.system.Isolation.InsertPin(now)
	case "discharge":
		isolation, err = s.system.Isolation.ConfirmDischarged(now)
	case "release":
		isolation, err = s.system.Isolation.Release(now)
	default:
		respondError(writer, http.StatusBadRequest, errors.New("unknown isolation action"))
		return
	}
	if err != nil {
		respondError(writer, http.StatusConflict, err)
		return
	}
	respondJSON(writer, http.StatusOK, isolation)
}

func (s *Server) startIsolation(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		TurbineID string `json:"turbine_id"`
		Reason    string `json:"reason"`
	}
	if err := decodeJSON(request, &input); err != nil {
		respondError(writer, http.StatusBadRequest, err)
		return
	}
	now := time.Now().UTC()
	isolation, err := s.system.Isolation.Start(input.TurbineID, input.Reason, now)
	if err != nil {
		respondError(writer, http.StatusConflict, err)
		return
	}
	respondJSON(writer, http.StatusAccepted, isolation)
}
