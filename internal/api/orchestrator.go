package api

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/cry-110/internal/gearbox"
	"github.com/wyw14/cry-110/internal/interlock"
	"github.com/wyw14/cry-110/internal/journal"
	"github.com/wyw14/cry-110/internal/model"
)

type IsolationOrchestrator struct {
	system *System
}

func NewIsolationOrchestrator(system *System) *IsolationOrchestrator {
	return &IsolationOrchestrator{system: system}
}

func (o *IsolationOrchestrator) Start(turbineID, reason string, now time.Time) (model.Isolation, error) {
	o.system.Rotor.ResetStillness()
	o.system.Torsion.Reset()
	isolation, err := o.system.Coordinator.StartIsolation(uuid.NewString(), turbineID, reason, now)
	if err != nil {
		return model.Isolation{}, err
	}
	if err := o.record(isolation, "isolation.requested", map[string]any{
		"turbine_id": turbineID,
		"reason":     reason,
	}, now); err != nil {
		return model.Isolation{}, err
	}
	return isolation, nil
}

func (o *IsolationOrchestrator) ConfirmCoasting(sample model.RotorTelemetry, now time.Time) (model.Isolation, error) {
	torsion := o.system.Torsion.Observe(gearbox.TorsionSample{
		TorqueNm: sample.ShaftTorqueNm, InstantRPM: sample.InstantRPM, At: sample.ObservedAt,
	})
	if err := o.system.Rotor.Record(sample); err != nil {
		return model.Isolation{}, fmt.Errorf("record rotor telemetry: %w", err)
	}
	isolation, err := o.system.Coordinator.AdvanceIsolation(model.IsolationCoasting, now)
	if err != nil {
		return model.Isolation{}, err
	}
	if err := o.record(isolation, "isolation.coasting", map[string]any{
		"average_rpm":    sample.AverageRPM,
		"instant_rpm":    sample.InstantRPM,
		"torque_nm":      sample.ShaftTorqueNm,
		"torsion_stable": torsion.Stable,
	}, now); err != nil {
		return model.Isolation{}, err
	}
	return isolation, nil
}

func (o *IsolationOrchestrator) Feather(now time.Time) (model.Isolation, error) {
	session, err := o.system.Pitch.StartFeather()
	if err != nil {
		return model.Isolation{}, fmt.Errorf("start emergency feather: %w", err)
	}
	state, unsafe, err := o.system.Pitch.Step(90, 90, 180, now)
	if err != nil {
		return model.Isolation{}, fmt.Errorf("advance emergency feather: %w", err)
	}
	if len(unsafe) != 0 {
		return model.Isolation{}, fmt.Errorf("pitch session %s has unsafe blades %v", session, unsafe)
	}
	o.system.BrakeSequence.FeatherUpdate(state, unsafe)
	isolation, err := o.system.Coordinator.AdvanceIsolation(model.IsolationFeathered, now)
	if err != nil {
		return model.Isolation{}, err
	}
	if err := o.record(isolation, "isolation.feathered", map[string]any{
		"pitch_session": session,
		"pitch_state":   state,
	}, now); err != nil {
		return model.Isolation{}, err
	}
	return isolation, nil
}

func (o *IsolationOrchestrator) Brake(now time.Time) (model.Isolation, error) {
	if err := o.system.Rotor.ApplyBrake(); err != nil {
		return model.Isolation{}, fmt.Errorf("apply rotor brake: %w", err)
	}
	if err := o.system.BrakeSequence.Complete(); err != nil {
		return model.Isolation{}, fmt.Errorf("complete brake sequence: %w", err)
	}
	isolation, err := o.system.Coordinator.AdvanceIsolation(model.IsolationBraked, now)
	if err != nil {
		return model.Isolation{}, err
	}
	if err := o.record(isolation, "isolation.braked", nil, now); err != nil {
		return model.Isolation{}, err
	}
	return isolation, nil
}

func (o *IsolationOrchestrator) InsertPin(now time.Time) (model.Isolation, error) {
	if err := o.system.Rotor.InsertPin(now); err != nil {
		return model.Isolation{}, fmt.Errorf("insert mechanical pin: %w", err)
	}
	isolation, err := o.system.Coordinator.AdvanceIsolation(model.IsolationPinned, now)
	if err != nil {
		return model.Isolation{}, err
	}
	if err := o.record(isolation, "isolation.pinned", nil, now); err != nil {
		return model.Isolation{}, err
	}
	return isolation, nil
}

func (o *IsolationOrchestrator) ConfirmDischarged(now time.Time) (model.Isolation, error) {
	if err := o.system.Hydraulic.Ready(); err != nil {
		return model.Isolation{}, fmt.Errorf("hydraulic state is not proven before discharge: %w", err)
	}
	isolation, err := o.system.Coordinator.AdvanceIsolation(model.IsolationDischarged, now)
	if err != nil {
		return model.Isolation{}, err
	}
	if err := o.record(isolation, "isolation.discharged", nil, now); err != nil {
		return model.Isolation{}, err
	}
	return isolation, nil
}

func (o *IsolationOrchestrator) Release(now time.Time) (model.Isolation, error) {
	isolation, err := o.system.Coordinator.AdvanceIsolation(model.IsolationReleased, now)
	if err != nil {
		return model.Isolation{}, err
	}
	if err := o.record(isolation, "isolation.released", nil, now); err != nil {
		return model.Isolation{}, err
	}
	if err := o.system.Snapshots.Save(journal.Snapshot{
		Version:   uint64(isolation.UpdatedAt.UnixNano()),
		CreatedAt: now,
		State: map[string]any{
			"isolation_id":    isolation.ID,
			"isolation_phase": isolation.Phase,
			"turbine_id":      isolation.TurbineID,
		},
	}); err != nil {
		return model.Isolation{}, fmt.Errorf("persist release snapshot: %w", err)
	}
	return isolation, nil
}

func (o *IsolationOrchestrator) record(isolation model.Isolation, kind string, attributes map[string]any, now time.Time) error {
	_, err := o.system.Journal.Append(model.Event{
		AggregateID: isolation.ID,
		Kind:        kind,
		At:          now,
		Attributes:  attributes,
	})
	return err
}

func (o *IsolationOrchestrator) PinState() interlock.PinState {
	_, _, state := o.system.Rotor.Status()
	return state
}
