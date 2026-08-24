package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/wyw14/cry-110/internal/hydraulic"
)

func (s *Server) getHydraulic(writer http.ResponseWriter, request *http.Request) {
	proof, err := s.system.Hydraulic.EmergencyPitchProof()
	if err != nil {
		respondError(writer, http.StatusConflict, err)
		return
	}
	_, warmupState := s.system.Warmup.Proof()
	respondJSON(writer, http.StatusOK, map[string]any{
		"emergency_pitch": proof,
		"warmup_state":    warmupState,
		"lube_available":  s.system.Lube.Available(5*time.Second, time.Now().UTC()),
	})
}

func (s *Server) updateHydraulic(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		Action          string  `json:"action"`
		SessionID       string  `json:"session_id"`
		NominalLiters   float64 `json:"nominal_liters"`
		PrechargeBar    float64 `json:"precharge_bar"`
		HeaderBar       float64 `json:"header_bar"`
		OilLiters       float64 `json:"oil_liters"`
		SupplyPressure  float64 `json:"supply_pressure_bar"`
		ReturnFlowLPM   float64 `json:"return_flow_lpm"`
		SumpTemperature float64 `json:"sump_temperature_c"`
	}
	if err := decodeJSON(request, &input); err != nil {
		respondError(writer, http.StatusBadRequest, err)
		return
	}
	now := time.Now().UTC()
	switch input.Action {
	case "accumulator":
		s.system.Hydraulic.Update(hydraulic.Accumulator{
			NominalLiters: input.NominalLiters, PrechargeBar: input.PrechargeBar,
			HeaderBar: input.HeaderBar, OilLiters: input.OilLiters,
		})
	case "start_warmup":
		if err := s.system.Warmup.Start(input.SessionID); err != nil {
			respondError(writer, http.StatusConflict, err)
			return
		}
	case "lube":
		if err := s.system.Lube.Observe(input.SupplyPressure, input.ReturnFlowLPM, input.SumpTemperature, now); err != nil {
			respondError(writer, http.StatusBadRequest, err)
			return
		}
		state := s.system.Warmup.Observe(s.system.Lube.WarmupProof(input.SessionID))
		respondJSON(writer, http.StatusOK, map[string]any{"action": input.Action, "warmup_state": state})
		return
	default:
		respondError(writer, http.StatusBadRequest, fmt.Errorf("unknown hydraulic action"))
		return
	}
	respondJSON(writer, http.StatusOK, map[string]string{"action": input.Action})
}

func (s *Server) getGenerator(writer http.ResponseWriter, request *http.Request) {
	state, outputKW := s.system.Generator.Status()
	respondJSON(writer, http.StatusOK, map[string]any{"state": state, "output_kw": outputKW})
}

func (s *Server) updateGenerator(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		Action   string  `json:"action"`
		OutputKW float64 `json:"output_kw"`
	}
	if err := decodeJSON(request, &input); err != nil {
		respondError(writer, http.StatusBadRequest, err)
		return
	}
	var err error
	switch input.Action {
	case "start":
		err = s.system.Generator.Start()
	case "synchronize":
		err = s.system.Generator.Synchronize(input.OutputKW)
	case "trip":
		s.system.Generator.Trip()
	case "reset":
		err = s.system.Generator.Reset()
	default:
		err = fmt.Errorf("unknown generator action")
	}
	if err != nil {
		respondError(writer, http.StatusConflict, err)
		return
	}
	state, outputKW := s.system.Generator.Status()
	respondJSON(writer, http.StatusOK, map[string]any{"state": state, "output_kw": outputKW})
}

func (s *Server) getVentilation(writer http.ResponseWriter, request *http.Request) {
	mode, flow, route := s.system.Ventilation.Status()
	respondJSON(writer, http.StatusOK, map[string]any{"mode": mode, "flow": flow, "route": route})
}

func (s *Server) updateVentilation(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		Action string  `json:"action"`
		Flow   float64 `json:"flow"`
	}
	if err := decodeJSON(request, &input); err != nil {
		respondError(writer, http.StatusBadRequest, err)
		return
	}
	var err error
	switch input.Action {
	case "cool":
		err = s.system.Ventilation.StartCooling(input.Flow)
	case "stop":
		s.system.Ventilation.Stop()
	default:
		err = fmt.Errorf("unknown ventilation action")
	}
	if err != nil {
		respondError(writer, http.StatusConflict, err)
		return
	}
	mode, flow, route := s.system.Ventilation.Status()
	respondJSON(writer, http.StatusOK, map[string]any{"mode": mode, "flow": flow, "route": route})
}

func (s *Server) getAccess(writer http.ResponseWriter, request *http.Request) {
	respondJSON(writer, http.StatusOK, map[string]int{"occupancy": s.system.Access.Occupancy()})
}

func (s *Server) updateAccess(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		Action string `json:"action"`
	}
	if err := decodeJSON(request, &input); err != nil {
		respondError(writer, http.StatusBadRequest, err)
		return
	}
	var err error
	switch input.Action {
	case "enter":
		err = s.system.Access.Enter()
	case "leave":
		err = s.system.Access.Leave()
	default:
		err = fmt.Errorf("unknown access action")
	}
	if err != nil {
		respondError(writer, http.StatusConflict, err)
		return
	}
	respondJSON(writer, http.StatusOK, map[string]int{"occupancy": s.system.Access.Occupancy()})
}
