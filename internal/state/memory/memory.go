package memory

import (
	"maxim/internal/models"
	"maxim/internal/state"
	"sync"
)

type Store[T models.Clonable[T]] struct {
	mu    sync.RWMutex
	items map[string]T
}

func (s *Store[T]) Create(v T) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.items[v.Id()]; exists {
		return state.ErrAlreadyExists
	}
	s.items[v.Id()] = v.Clone()
	return nil
}

func (s *Store[T]) Get(id string) (T, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, exists := s.items[id]
	if !exists {
		var zero T
		return zero, false
	}
	return item.Clone(), true
}

func (s *Store[T]) Update(v T) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.items[v.Id()]; !exists {
		return state.ErrNotFound
	}
	s.items[v.Id()] = v.Clone()
	return nil
}

func (s *Store[T]) List() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]string, 0, len(s.items))
	for id := range s.items {
		items = append(items, id)
	}
	return items, nil
}

func NewStore[T models.Clonable[T]]() *Store[T] {
	return &Store[T]{
		items: make(map[string]T),
	}
}

type (
	UserStore    = Store[*models.User]
	ChannelStore = Store[*models.Channel]
)

func NewUserStore() *UserStore {
	return NewStore[*models.User]()
}

func NewChannelStore() *ChannelStore {
	return NewStore[*models.Channel]()
}
