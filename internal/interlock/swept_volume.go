package interlock

import (
	"errors"
	"math"
	"sort"
	"sync"
)

type SweptVolume struct {
	OperationID string
	Axis        string
	MinAngle    float64
	MaxAngle    float64
	MinRadiusM  float64
	MaxRadiusM  float64
	StartSecond float64
	EndSecond   float64
}

func (v SweptVolume) Validate() error {
	if v.OperationID == "" || v.Axis == "" {
		return errors.New("swept volume identity is required")
	}
	if v.MaxRadiusM < v.MinRadiusM || v.EndSecond < v.StartSecond {
		return errors.New("swept volume bounds are reversed")
	}
	values := []float64{v.MinAngle, v.MaxAngle, v.MinRadiusM, v.MaxRadiusM, v.StartSecond, v.EndSecond}
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return errors.New("swept volume contains a non-finite bound")
		}
	}
	return nil
}

func (v SweptVolume) overlaps(other SweptVolume) bool {
	timeOverlap := v.StartSecond <= other.EndSecond && other.StartSecond <= v.EndSecond
	radiusOverlap := v.MinRadiusM <= other.MaxRadiusM && other.MinRadiusM <= v.MaxRadiusM
	angleOverlap := v.MinAngle <= other.MaxAngle && other.MinAngle <= v.MaxAngle
	return timeOverlap && radiusOverlap && angleOverlap
}

type ReservationBook struct {
	mu           sync.Mutex
	reservations map[string][]SweptVolume
	blocked      []SweptVolume
}

func NewReservationBook(blocked []SweptVolume) *ReservationBook {
	return &ReservationBook{reservations: make(map[string][]SweptVolume), blocked: append([]SweptVolume(nil), blocked...)}
}

func (b *ReservationBook) TryReserve(operationID string, volumes []SweptVolume) error {
	if operationID == "" || len(volumes) == 0 {
		return errors.New("operation and swept volumes are required")
	}
	for _, volume := range volumes {
		if err := volume.Validate(); err != nil {
			return err
		}
		if volume.OperationID != operationID {
			return errors.New("swept volume operation identity mismatch")
		}
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, exists := b.reservations[operationID]; exists {
		return errors.New("operation already owns swept volume")
	}
	for _, candidate := range volumes {
		for _, obstacle := range b.blocked {
			if candidate.overlaps(obstacle) {
				return errors.New("motion intersects a fixed exclusion volume")
			}
		}
		for _, reserved := range b.reservations {
			for _, existing := range reserved {
				if candidate.overlaps(existing) {
					return errors.New("motion intersects a reserved swept volume")
				}
			}
		}
	}
	b.reservations[operationID] = append([]SweptVolume(nil), volumes...)
	return nil
}

func (b *ReservationBook) Release(operationID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.reservations, operationID)
}

func (b *ReservationBook) Owners() []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	owners := make([]string, 0, len(b.reservations))
	for owner := range b.reservations {
		owners = append(owners, owner)
	}
	sort.Strings(owners)
	return owners
}
