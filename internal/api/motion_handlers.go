package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/wyw14/cry-110/internal/turninggear"
)

func (s *Server) getPitch(writer http.ResponseWriter, request *http.Request) {
	session, state, unsafe := s.system.Pitch.Status(time.Now().UTC())
	respondJSON(writer, http.StatusOK, map[string]any{"session_id": session, "state": state, "unsafe_blades": unsafe})
}

func (s *Server) startPitch(writer http.ResponseWriter, request *http.Request) {
	session, err := s.system.Pitch.StartFeather()
	if err != nil {
		respondError(writer, http.StatusConflict, err)
		return
	}
	state, unsafe, err := s.system.Pitch.Step(90, 90, 180, time.Now().UTC())
	if err != nil {
		respondError(writer, http.StatusConflict, err)
		return
	}
	respondJSON(writer, http.StatusAccepted, map[string]any{"session_id": session, "state": state, "unsafe_blades": unsafe})
}

func (s *Server) getYaw(writer http.ResponseWriter, request *http.Request) {
	angle, target, thermal := s.system.Yaw.Status(time.Now().UTC())
	respondJSON(writer, http.StatusOK, map[string]any{
		"angle": angle, "target": target, "thermal": thermal,
		"brake_pressurized":      s.system.YawBrakeService.Pressurized(),
		"emergency_brake_locked": s.system.YawBrake.Locked(),
	})
}

func (s *Server) finishYaw(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		OperationID string  `json:"operation_id"`
		Inertia     float64 `json:"inertia"`
		InitialRPM  float64 `json:"initial_rpm"`
	}
	if err := decodeJSON(request, &input); err != nil {
		respondError(writer, http.StatusBadRequest, err)
		return
	}
	if err := s.system.Yaw.Finish(input.OperationID, input.Inertia, input.InitialRPM, time.Now().UTC()); err != nil {
		respondError(writer, http.StatusConflict, err)
		return
	}
	respondJSON(writer, http.StatusOK, map[string]any{"completed": input.OperationID})
}

func (s *Server) startYaw(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		Target float64 `json:"target"`
		Speed  float64 `json:"speed"`
	}
	if err := decodeJSON(request, &input); err != nil {
		respondError(writer, http.StatusBadRequest, err)
		return
	}
	motion, err := s.system.Yaw.Request(input.Target, input.Speed, time.Now().UTC())
	if err != nil {
		respondError(writer, http.StatusConflict, err)
		return
	}
	respondJSON(writer, http.StatusAccepted, motion)
}

func (s *Server) getTurningGear(writer http.ResponseWriter, request *http.Request) {
	operation, target, state := s.system.TurningGear.Status()
	applied, released := s.system.MainShaftBrake.State()
	respondJSON(writer, http.StatusOK, map[string]any{
		"operation_id": operation, "target": target, "disengage_state": state,
		"warmup_permit":    s.system.TurningPermit.Granted(),
		"main_shaft_brake": map[string]bool{"applied": applied, "released": released},
	})
}

func (s *Server) updateTurningGear(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		Action     string  `json:"action"`
		EncoderRaw float64 `json:"encoder_raw"`
		Amperes    float64 `json:"amperes"`
		PawlState  string  `json:"pawl_state"`
	}
	if err := decodeJSON(request, &input); err != nil {
		respondError(writer, http.StatusBadRequest, err)
		return
	}
	now := time.Now().UTC()
	switch input.Action {
	case "update":
		command, err := s.system.TurningGear.Update(input.EncoderRaw)
		if err != nil {
			respondError(writer, http.StatusConflict, err)
			return
		}
		respondJSON(writer, http.StatusOK, map[string]float64{"motor_rpm": command})
	case "begin_disengage":
		if err := s.system.TurningGear.BeginDisengage(); err != nil {
			respondError(writer, http.StatusConflict, err)
			return
		}
		respondJSON(writer, http.StatusOK, map[string]string{"state": "unloading"})
	case "disengage_feedback":
		state, err := s.system.TurningGear.ApplyDisengageFeedback(input.Amperes, input.PawlState, now)
		if err != nil {
			respondError(writer, http.StatusConflict, err)
			return
		}
		operation, _, _ := s.system.TurningGear.Status()
		if state == turninggear.DisengageComplete {
			if err := s.system.MainShaftBrake.Release(operation, now); err != nil {
				respondError(writer, http.StatusConflict, err)
				return
			}
		}
		respondJSON(writer, http.StatusOK, map[string]any{
			"state": state, "block_reason": s.system.MainShaftBrakeReason(),
		})
	case "apply_brake":
		s.system.MainShaftBrake.Apply()
		respondJSON(writer, http.StatusOK, map[string]bool{"applied": true})
	default:
		respondError(writer, http.StatusBadRequest, fmt.Errorf("unknown turning gear action"))
	}
}

func (s *Server) startTurningGear(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		Target     float64 `json:"target"`
		EncoderRaw float64 `json:"encoder_raw"`
	}
	if err := decodeJSON(request, &input); err != nil {
		respondError(writer, http.StatusBadRequest, err)
		return
	}
	operation, command, err := s.system.TurningGear.Start(input.Target, input.EncoderRaw, time.Now().UTC())
	if err != nil {
		respondError(writer, http.StatusConflict, err)
		return
	}
	respondJSON(writer, http.StatusAccepted, map[string]any{"operation_id": operation, "motor_rpm": command})
}
