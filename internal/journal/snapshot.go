package journal

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Snapshot struct {
	Version   uint64         `json:"version"`
	CreatedAt time.Time      `json:"created_at"`
	State     map[string]any `json:"state"`
}

type SnapshotStore struct {
	mu   sync.Mutex
	path string
}

func OpenSnapshots(path string) (*SnapshotStore, error) {
	if path == "" {
		return nil, errors.New("snapshot path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create snapshot directory: %w", err)
	}
	return &SnapshotStore{path: path}, nil
}

func (s *SnapshotStore) Save(snapshot Snapshot) error {
	if snapshot.Version == 0 || snapshot.CreatedAt.IsZero() {
		return errors.New("snapshot version and timestamp are required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("encode snapshot: %w", err)
	}
	temporary := s.path + ".new"
	if err := os.WriteFile(temporary, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write snapshot: %w", err)
	}
	if err := os.Rename(temporary, s.path); err != nil {
		return fmt.Errorf("replace snapshot: %w", err)
	}
	return nil
}

func (s *SnapshotStore) Load() (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.path)
	if err != nil {
		return Snapshot{}, err
	}
	var snapshot Snapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return Snapshot{}, fmt.Errorf("decode snapshot: %w", err)
	}
	return snapshot, nil
}
