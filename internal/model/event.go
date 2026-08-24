package model

import (
	"fmt"
	"time"
)

type Event struct {
	Sequence    uint64         `json:"sequence"`
	AggregateID string         `json:"aggregate_id"`
	Kind        string         `json:"kind"`
	At          time.Time      `json:"at"`
	Attributes  map[string]any `json:"attributes,omitempty"`
}

func (e Event) Validate() error {
	if e.AggregateID == "" || e.Kind == "" || e.At.IsZero() {
		return fmt.Errorf("event aggregate, kind and timestamp are required")
	}
	return nil
}
