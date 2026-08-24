package api

import (
	"fmt"
	"net/http"
	"time"
)

func (s *Server) getCrane(writer http.ResponseWriter, request *http.Request) {
	angle, target := s.system.Crane.Status()
	parkState, hookMeters, boomAngle, observedAt := s.system.CranePark.Status()
	respondJSON(writer, http.StatusOK, map[string]any{
		"angle": angle, "target": target, "swept_volume_owners": s.system.Reservations.Owners(),
		"park": map[string]any{
			"state": parkState, "hook_meters": hookMeters,
			"boom_angle": boomAngle, "observed_at": observedAt,
		},
	})
}

func (s *Server) startCrane(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		Target  float64 `json:"target"`
		Seconds float64 `json:"seconds"`
	}
	if err := decodeJSON(request, &input); err != nil {
		respondError(writer, http.StatusBadRequest, err)
		return
	}
	motion, volumes, err := s.system.Crane.Plan(input.Target, input.Seconds)
	if err != nil {
		respondError(writer, http.StatusConflict, err)
		return
	}
	if err := s.system.Crane.Commit(motion, volumes); err != nil {
		respondError(writer, http.StatusConflict, err)
		return
	}
	respondJSON(writer, http.StatusAccepted, motion)
}

func (s *Server) updateCrane(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		Action      string  `json:"action"`
		OperationID string  `json:"operation_id"`
		HookMeters  float64 `json:"hook_meters"`
		BoomAngle   float64 `json:"boom_angle"`
	}
	if err := decodeJSON(request, &input); err != nil {
		respondError(writer, http.StatusBadRequest, err)
		return
	}
	now := time.Now().UTC()
	var err error
	switch input.Action {
	case "complete_slew":
		err = s.system.Crane.Complete(input.OperationID)
	case "deploy":
		err = s.system.CranePark.Deploy(input.HookMeters, input.BoomAngle, now)
	case "begin_park":
		s.system.CranePark.BeginPark(now)
	case "confirm_parked":
		err = s.system.CranePark.ConfirmSecured(input.HookMeters, input.BoomAngle, now)
	default:
		err = fmt.Errorf("unknown crane action")
	}
	if err != nil {
		respondError(writer, http.StatusConflict, err)
		return
	}
	respondJSON(writer, http.StatusOK, map[string]string{"action": input.Action})
}
