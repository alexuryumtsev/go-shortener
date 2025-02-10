package storage

import (
	"context"
	"sync"

	"github.com/alexuryumtsev/go-shortener/internal/app/models"
)

// InMemoryStorage управляет сохранением и получением данных в памяти.
type InMemoryStorage struct {
	mu   sync.RWMutex
	data map[string]string
}

// NewInMemoryStorage создаёт новое хранилище в памяти.
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		data: make(map[string]string),
	}
}

// Save сохраняет URL в памяти.
func (s *InMemoryStorage) Save(ctx context.Context, urlModel models.URLModel) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[urlModel.ID] = urlModel.URL
	return nil
}

// Get возвращает оригинальный URL по идентификатору из памяти.
func (s *InMemoryStorage) Get(ctx context.Context, id string) (models.URLModel, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	url, exists := s.data[id]
	return models.URLModel{ID: id, URL: url}, exists
}

// LoadFromFile загружает данные из памяти (не требуется для памяти).
func (s *InMemoryStorage) LoadFromFile() error {
	return nil
}
