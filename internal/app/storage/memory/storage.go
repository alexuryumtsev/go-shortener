package memory

import (
	"context"
	"sync"

	"github.com/alexuryumtsev/go-shortener/internal/app/models"
)

// InMemoryStorage управляет сохранением и получением данных в памяти.
type InMemoryStorage struct {
	mu       sync.RWMutex
	data     map[string]string
	userData map[string][]models.URLModel
}

// NewInMemoryStorage создаёт новое хранилище в памяти.
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		data:     make(map[string]string),
		userData: make(map[string][]models.URLModel),
	}
}

// Save сохраняет URL в памяти.
func (s *InMemoryStorage) Save(ctx context.Context, urlModel models.URLModel) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[urlModel.ID] = urlModel.URL
	userID := urlModel.UserID
	s.userData[userID] = append(s.userData[userID], urlModel)
	return nil
}

// SaveBatch сохраняет множество URL в памяти.
func (s *InMemoryStorage) SaveBatch(ctx context.Context, urlModels []models.URLModel) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, urlModel := range urlModels {
		userID := urlModel.UserID
		s.data[urlModel.ID] = urlModel.URL
		s.userData[userID] = append(s.userData[userID], urlModel)
	}
	return nil
}

// Get возвращает оригинальный URL по идентификатору из памяти.
func (s *InMemoryStorage) Get(ctx context.Context, id string) (models.URLModel, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	url, exists := s.data[id]
	return models.URLModel{ID: id, URL: url}, exists
}

// GetUserURLs возвращает все URL, сокращённые пользователем.
func (s *InMemoryStorage) GetUserURLs(ctx context.Context, userID string) ([]models.URLModel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	urls, exists := s.userData[userID]
	if !exists {
		return nil, nil
	}
	return urls, nil
}

// LoadFromFile загружает данные из памяти (не требуется для памяти).
func (s *InMemoryStorage) LoadFromFile() error {
	return nil
}

// Ping проверяет соединение с памятью (всегда возвращает nil).
func (s *InMemoryStorage) Ping(ctx context.Context) error {
	return nil
}

func (s *InMemoryStorage) DeleteUserURLs(ctx context.Context, userID string, shortURLs []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, shortURL := range shortURLs {
		if url, exists := s.data[shortURL]; exists {
			for i, userURL := range s.userData[userID] {
				if userURL.ID == shortURL {
					s.userData[userID][i].Deleted = true
					break
				}
			}
			s.data[shortURL] = url
		}
	}

	return nil
}
