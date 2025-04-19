package storage

import (
	"context"

	"github.com/alexuryumtsev/go-shortener/internal/app/models"
)

type MockStorage struct {
	data map[string]models.URLModel
}

func NewMockStorage() *MockStorage {
	return &MockStorage{data: make(map[string]models.URLModel)}
}

func (m *MockStorage) Save(ctx context.Context, urlModel models.URLModel) error {
	m.data[urlModel.ID] = urlModel
	return nil
}

func (m *MockStorage) SaveBatch(ctx context.Context, urlModels []models.URLModel) error {
	for _, urlModel := range urlModels {
		m.data[urlModel.ID] = urlModel
	}
	return nil
}

func (m *MockStorage) Get(ctx context.Context, id string) (models.URLModel, bool) {
	urlModel, exists := m.data[id]
	return urlModel, exists
}

func (m *MockStorage) GetUserURLs(ctx context.Context, userID string) ([]models.URLModel, error) {
	var userURLs []models.URLModel
	for _, urlModel := range m.data {
		if urlModel.UserID == userID {
			userURLs = append(userURLs, urlModel)
		}
	}
	return userURLs, nil
}

// LoadFromFile имитирует загрузку данных из файла.
func (m *MockStorage) LoadFromFile() error {
	// Можно имитировать ошибку или инициализировать данными для тестов.
	return nil
}

// SaveToFile имитирует сохранение данных в файл.
func (m *MockStorage) SaveToFile(filePath string) error {
	// Для тестов можно просто возвращать успешный результат.
	return nil
}

func (m *MockStorage) DeleteUserURLs(ctx context.Context, userID string, shortURLs []string) error {
	for _, shortURL := range shortURLs {
		if urlModel, exists := m.data[shortURL]; exists && urlModel.UserID == userID {
			urlModel.Deleted = true
			m.data[shortURL] = urlModel
		}
	}
	return nil
}
