package journal

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/wyw14/cry-110/internal/model"
)

type Store struct {
	mu       sync.Mutex
	path     string
	sequence uint64
}

func Open(path string) (*Store, error) {
	if path == "" {
		return nil, errors.New("journal path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create journal directory: %w", err)
	}
	store := &Store{path: path}
	events, err := store.readUnlocked()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	for _, event := range events {
		if event.Sequence > store.sequence {
			store.sequence = event.Sequence
		}
	}
	return store, nil
}

func (s *Store) Append(event model.Event) (model.Event, error) {
	if err := event.Validate(); err != nil {
		return model.Event{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sequence++
	event.Sequence = s.sequence
	file, err := os.OpenFile(s.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return model.Event{}, fmt.Errorf("open journal: %w", err)
	}
	encoder := json.NewEncoder(file)
	writeErr := encoder.Encode(event)
	closeErr := file.Close()
	if writeErr != nil {
		return model.Event{}, fmt.Errorf("append journal: %w", writeErr)
	}
	if closeErr != nil {
		return model.Event{}, fmt.Errorf("close journal: %w", closeErr)
	}
	return event, nil
}

func (s *Store) Events() ([]model.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.readUnlocked()
}

func (s *Store) readUnlocked() ([]model.Event, error) {
	file, err := os.Open(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []model.Event{}, nil
		}
		return nil, err
	}
	defer file.Close()
	var events []model.Event
	scanner := bufio.NewScanner(file)
	buffer := make([]byte, 64*1024)
	scanner.Buffer(buffer, 1024*1024)
	for scanner.Scan() {
		var event model.Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return nil, fmt.Errorf("decode journal event: %w", err)
		}
		events = append(events, event)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan journal: %w", err)
	}
	return events, nil
}

func (s *Store) Path() string {
	return s.path
}
