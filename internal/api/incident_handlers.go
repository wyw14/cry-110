package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/wyw14/cry-110/internal/model"
)

func (s *Server) getIncidents(writer http.ResponseWriter, request *http.Request) {
	fireActive, fireIncident, fireRoute := s.system.Fire.Status()
	icingState, icingSession, icingIncident, shedCycle, exclusion := s.system.Icing.Status()
	lightningActive, lightningIncident := s.system.Lightning.Status()
	events, err := s.system.Journal.Events()
	if err != nil {
		respondError(writer, http.StatusInternalServerError, err)
		return
	}
	respondJSON(writer, http.StatusOK, map[string]any{
		"incidents": s.system.Coordinator.Incidents(),
		"fire":      map[string]any{"active": fireActive, "incident": fireIncident, "route": fireRoute},
		"icing": map[string]any{
			"state": icingState, "session": icingSession, "incident": icingIncident,
			"shed_cycle": shedCycle, "exclusion": exclusion,
		},
		"lightning": map[string]any{"active": lightningActive, "incident": lightningIncident},
		"events":    events,
	})
}

func (s *Server) updateIncident(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		Kind             model.IncidentKind `json:"kind"`
		Action           string             `json:"action"`
		InspectionPassed bool               `json:"inspection_passed"`
		NoResidualIce    bool               `json:"no_residual_ice"`
	}
	if err := decodeJSON(request, &input); err != nil {
		respondError(writer, http.StatusBadRequest, err)
		return
	}
	now := time.Now().UTC()
	var err error
	switch input.Kind {
	case model.IncidentFire:
		if input.Action != "reset" {
			err = fmt.Errorf("unknown fire action")
		} else {
			err = s.system.Fire.Reset(input.InspectionPassed, now)
		}
	case model.IncidentIcing:
		switch input.Action {
		case "stopped":
			err = s.system.Icing.OnStopped()
		case "shed":
			err = s.system.Icing.StartShedCycle()
		case "inspect":
			err = s.system.Icing.Inspect()
		case "clear":
			err = s.system.Icing.Clear(input.NoResidualIce, now)
		default:
			err = fmt.Errorf("unknown icing action")
		}
	case model.IncidentLightning:
		switch input.Action {
		case "crane_parked":
			err = s.system.Lightning.ConfirmCraneParked(now)
		case "reset":
			err = s.system.Lightning.Reset(input.InspectionPassed, now)
		default:
			err = fmt.Errorf("unknown lightning action")
		}
	default:
		err = fmt.Errorf("unsupported incident kind %q", input.Kind)
	}
	if err != nil {
		respondError(writer, http.StatusConflict, err)
		return
	}
	respondJSON(writer, http.StatusOK, map[string]string{"kind": string(input.Kind), "action": input.Action})
}

func (s *Server) startIncident(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		Kind    model.IncidentKind `json:"kind"`
		Message string             `json:"message"`
	}
	if err := decodeJSON(request, &input); err != nil {
		respondError(writer, http.StatusBadRequest, err)
		return
	}
	now := time.Now().UTC()
	var incident model.Incident
	var err error
	switch input.Kind {
	case model.IncidentFire:
		incident, err = s.system.Fire.StartPurge(input.Message, 180, now)
	case model.IncidentIcing:
		incident, err = s.system.Icing.Trigger(input.Message, now)
	case model.IncidentLightning:
		incident, err = s.system.Lightning.Trigger(input.Message, now)
	default:
		err = fmt.Errorf("unsupported incident kind %q", input.Kind)
	}
	if err != nil {
		respondError(writer, http.StatusConflict, err)
		return
	}
	s.system.Coordinator.RaiseIncident(incident)
	if _, err := s.system.Journal.Append(model.Event{
		AggregateID: incident.ID, Kind: "incident.raised", At: now,
		Attributes: map[string]any{"kind": incident.Kind, "message": incident.Message},
	}); err != nil {
		respondError(writer, http.StatusInternalServerError, err)
		return
	}
	respondJSON(writer, http.StatusAccepted, incident)
}
