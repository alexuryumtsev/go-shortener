package url

import (
	"context"
	"fmt"
	"strings"

	"github.com/alexuryumtsev/go-shortener/internal/app/models"
)

// MockURLService - это структура, которая реализует интерфейс URLService для тестирования
type MockURLService struct {
	baseURL     string
	urls        map[string]string   // мапа для хранения id -> original_url
	userURLs    map[string][]string // мапа для хранения user_id -> []short_url
	deletedURLs map[string]bool     // мапа для хранения удаленных URLs
	err         error
}

// NewMockURLService создает новый экземпляр MockURLService с заданными параметрами
func NewMockURLService(baseURL string, err error) *MockURLService {
	return &MockURLService{
		baseURL:     strings.TrimSuffix(baseURL, "/"),
		urls:        make(map[string]string),
		userURLs:    make(map[string][]string),
		deletedURLs: make(map[string]bool),
		err:         err,
	}
}

// ShortenerURL возвращает предустановленный короткий URL
func (m *MockURLService) ShortenerURL(ctx context.Context, originalURL, userID string) (string, error) {
	if originalURL == "" {
		return "", fmt.Errorf("empty URL")
	}

	id := fmt.Sprintf("%x", len(m.urls)+1) // простая генерация ID
	m.urls[id] = originalURL
	m.userURLs[userID] = append(m.userURLs[userID], id)

	return fmt.Sprintf("%s/%s", m.baseURL, id), m.err
}

// SaveBatchShortenerURL сохраняет пакет URL и возвращает их сокращенные версии
func (m *MockURLService) SaveBatchShortenerURL(ctx context.Context, batchModels []models.URLBatchModel, userID string) ([]models.BatchResponseModel, error) {
	if m.err != nil {
		return nil, m.err
	}

	result := make([]models.BatchResponseModel, len(batchModels))
	for i, model := range batchModels {
		id := fmt.Sprintf("%x", len(m.urls)+i+1)
		m.urls[id] = model.OriginalURL
		m.userURLs[userID] = append(m.userURLs[userID], id)

		result[i] = models.BatchResponseModel{
			CorrelationID: model.CorrelationID,
			ShortURL:      fmt.Sprintf("%s/%s", m.baseURL, id),
		}
	}

	return result, nil
}

// DeleteUserURLsBatch помечает URL пользователя как удаленные
func (m *MockURLService) DeleteUserURLsBatch(ctx context.Context, userID string, shortURLs []string) error {
	if m.err != nil {
		return m.err
	}

	for _, url := range shortURLs {
		m.deletedURLs[url] = true
	}
	return nil
}

// GetURLByID возвращает предустановленный оригинальный URL
func (m *MockURLService) GetURLByID(ctx context.Context, id string) (string, bool, error) {
	if m.err != nil {
		return "", false, m.err
	}

	url := m.urls[id]
	isDeleted := m.deletedURLs[id]
	return url, isDeleted, nil
}

// GetUserURLs возвращает все URLs пользователя
func (m *MockURLService) GetUserURLs(ctx context.Context, userID string) ([]models.UserURLModel, error) {
	if m.err != nil {
		return nil, m.err
	}

	urls, exists := m.userURLs[userID]
	if !exists || len(urls) == 0 {
		return []models.UserURLModel{}, nil
	}

	result := make([]models.UserURLModel, 0, len(urls))
	for _, id := range urls {
		if url, exists := m.urls[id]; exists {
			result = append(result, models.UserURLModel{
				ShortURL:    fmt.Sprintf("%s/%s", m.baseURL, id),
				OriginalURL: url,
			})
		}
	}
	return result, nil
}

// SetError устанавливает ошибку для тестирования
func (m *MockURLService) SetError(err error) {
	m.err = err
}

// AddURL добавляет URL для тестирования
func (m *MockURLService) AddURL(id, originalURL, userID string) {
	m.urls[id] = originalURL
	m.userURLs[userID] = append(m.userURLs[userID], id)
}

// MarkURLAsDeleted помечает URL как удаленный
func (m *MockURLService) MarkURLAsDeleted(id string) {
	m.deletedURLs[id] = true
}
