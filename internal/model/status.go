package model

import (
	"errors"
	"fmt"
	"time"
)

type IsolationPhase string

const (
	IsolationRequested  IsolationPhase = "requested"
	IsolationCoasting   IsolationPhase = "coasting"
	IsolationFeathered  IsolationPhase = "feathered"
	IsolationBraked     IsolationPhase = "braked"
	IsolationPinned     IsolationPhase = "pinned"
	IsolationDischarged IsolationPhase = "discharged"
	IsolationReleased   IsolationPhase = "released"
)

var isolationOrder = []IsolationPhase{
	IsolationRequested,
	IsolationCoasting,
	IsolationFeathered,
	IsolationBraked,
	IsolationPinned,
	IsolationDischarged,
	IsolationReleased,
}

type Isolation struct {
	ID        string         `json:"id"`
	TurbineID string         `json:"turbine_id"`
	Phase     IsolationPhase `json:"phase"`
	Reason    string         `json:"reason"`
	UpdatedAt time.Time      `json:"updated_at"`
}

func NewIsolation(id, turbineID, reason string, now time.Time) (Isolation, error) {
	if id == "" || turbineID == "" {
		return Isolation{}, errors.New("isolation identity is required")
	}
	if reason == "" {
		return Isolation{}, errors.New("isolation reason is required")
	}
	return Isolation{ID: id, TurbineID: turbineID, Phase: IsolationRequested, Reason: reason, UpdatedAt: now}, nil
}

func (i Isolation) Advance(next IsolationPhase, now time.Time) (Isolation, error) {
	currentIndex, nextIndex := -1, -1
	for index, phase := range isolationOrder {
		if phase == i.Phase {
			currentIndex = index
		}
		if phase == next {
			nextIndex = index
		}
	}
	if currentIndex < 0 || nextIndex != currentIndex+1 {
		return i, fmt.Errorf("invalid isolation transition %s -> %s", i.Phase, next)
	}
	i.Phase = next
	i.UpdatedAt = now
	return i, nil
}

func (i Isolation) Terminal() bool {
	return i.Phase == IsolationReleased
}

type IncidentKind string

const (
	IncidentLightning  IncidentKind = "lightning"
	IncidentFire       IncidentKind = "fire"
	IncidentIcing      IncidentKind = "icing"
	IncidentMechanical IncidentKind = "mechanical"
)

type Incident struct {
	ID        string       `json:"id"`
	Kind      IncidentKind `json:"kind"`
	Message   string       `json:"message"`
	Latched   bool         `json:"latched"`
	RaisedAt  time.Time    `json:"raised_at"`
	ClearedAt *time.Time   `json:"cleared_at,omitempty"`
}

func NewIncident(id string, kind IncidentKind, message string, now time.Time) (Incident, error) {
	if id == "" || message == "" {
		return Incident{}, errors.New("incident identity and message are required")
	}
	switch kind {
	case IncidentLightning, IncidentFire, IncidentIcing, IncidentMechanical:
	default:
		return Incident{}, fmt.Errorf("unknown incident kind %q", kind)
	}
	return Incident{ID: id, Kind: kind, Message: message, Latched: true, RaisedAt: now}, nil
}

func (i Incident) Clear(now time.Time) Incident {
	i.Latched = false
	i.ClearedAt = &now
	return i
}
