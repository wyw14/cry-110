package brake

import (
	"fmt"
	"sync"

	"github.com/wyw14/cry-110/internal/pitch"
)

type SequenceState string

const (
	SequenceWaitingForFeather SequenceState = "waiting_for_feather"
	SequenceNormalBrake       SequenceState = "normal_brake"
	SequenceAsymmetricBrake   SequenceState = "asymmetric_brake"
	SequenceComplete          SequenceState = "complete"
)

type Sequence struct {
	mu    sync.Mutex
	state SequenceState
}

func NewSequence() *Sequence {
	return &Sequence{state: SequenceWaitingForFeather}
}

func (s *Sequence) FeatherUpdate(group pitch.GroupState, unsafeBlades []int) SequenceState {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch group {
	case pitch.GroupSafe:
		s.state = SequenceNormalBrake
	case pitch.GroupDegraded:
		if len(unsafeBlades) > 0 {
			s.state = SequenceAsymmetricBrake
		}
	default:
		s.state = SequenceWaitingForFeather
	}
	return s.state
}

func (s *Sequence) Complete() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state != SequenceNormalBrake && s.state != SequenceAsymmetricBrake {
		return fmt.Errorf("brake sequence has no valid feather state")
	}
	s.state = SequenceComplete
	return nil
}

func (s *Sequence) State() SequenceState {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state
}
