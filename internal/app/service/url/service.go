// Package url предоставляет реализацию сервиса для работы с URL.
package url

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/alexuryumtsev/go-shortener/internal/app/models"
	"github.com/alexuryumtsev/go-shortener/internal/app/storage"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// URLService определяет интерфейс для работы с URL.
type URLService interface {
	// ShortenerURL создает короткий URL для переданного оригинального URL
	ShortenerURL(ctx context.Context, originalURL, userID string) (string, error)

	// SaveBatchShortenerURL сохраняет пакет URL и возвращает их сокращенные версии
	SaveBatchShortenerURL(ctx context.Context, batchModels []models.URLBatchModel, userID string) ([]models.BatchResponseModel, error)

	// DeleteUserURLsBatch помечает URL пользователя как удаленные
	DeleteUserURLsBatch(ctx context.Context, userID string, shortURLs []string) error

	// GetURLByID получает оригинальный URL по ID
	GetURLByID(ctx context.Context, id string) (string, bool, error)

	// GetUserURLs получает все URL пользователя
	GetUserURLs(ctx context.Context, userID string) ([]models.UserURLModel, error)
}

// urlService реализация URLService
type urlService struct {
	storage   storage.URLStorage
	baseURL   string
	batchSize int
}

// NewURLService создаёт новый экземпляр сервиса для работы с URL.
func NewURLService(storage storage.URLStorage, baseURL string, batchSize int) URLService {
	return &urlService{
		storage:   storage,
		baseURL:   strings.TrimSuffix(baseURL, "/"),
		batchSize: batchSize,
	}
}

// ShortenerURL сокращает URL и сохраняет в базе
func (s *urlService) ShortenerURL(ctx context.Context, originalURL, userID string) (string, error) {
	if originalURL == "" {
		return "", fmt.Errorf("empty URL")
	}

	id := generateID(originalURL)
	urlModel := models.URLModel{ID: id, URL: originalURL, UserID: userID}
	shortenedURL := s.baseURL + "/" + id

	err := s.storage.Save(ctx, urlModel)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return shortenedURL, err
		}
		return "", err
	}

	return shortenedURL, nil
}

// SaveBatchShortenerURL сохраняет пакет URL и возвращает их сокращенные версии
func (s *urlService) SaveBatchShortenerURL(ctx context.Context, batchModels []models.URLBatchModel, userID string) ([]models.BatchResponseModel, error) {
	if len(batchModels) == 0 {
		return nil, fmt.Errorf("empty batch")
	}

	var urlModels []models.URLModel
	for _, req := range batchModels {
		if req.OriginalURL == "" {
			return nil, fmt.Errorf("empty URL in batch")
		}

		urlModels = append(urlModels, models.URLModel{
			ID:     generateID(req.OriginalURL),
			URL:    req.OriginalURL,
			UserID: userID,
		})
	}

	// Подготовка ответа
	var responseModels []models.BatchResponseModel
	for i, urlModel := range urlModels {
		responseModels = append(responseModels, models.BatchResponseModel{
			CorrelationID: batchModels[i].CorrelationID,
			ShortURL:      s.baseURL + "/" + urlModel.ID,
		})
	}

	err := s.storage.SaveBatch(ctx, urlModels)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return responseModels, err
		}
		return nil, err
	}

	return responseModels, nil
}

// DeleteUserURLsBatch помечает URL пользователя как удаленные
func (s *urlService) DeleteUserURLsBatch(ctx context.Context, userID string, shortURLs []string) error {
	if len(shortURLs) == 0 {
		return nil
	}

	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Извлекаем ID из полных URLs, если переданы полные URL
	var ids []string
	for _, urlStr := range shortURLs {
		if strings.HasPrefix(urlStr, s.baseURL) {
			// Если это полный URL, извлекаем ID
			id := strings.TrimPrefix(urlStr, s.baseURL+"/")
			ids = append(ids, id)
		} else {
			// Иначе считаем, что это уже ID
			ids = append(ids, urlStr)
		}
	}

	// Разбиваем на партии для обработки
	for i := 0; i < len(ids); i += s.batchSize {
		end := i + s.batchSize
		if end > len(ids) {
			end = len(ids)
		}

		batch := ids[i:end]
		if err := s.storage.DeleteUserURLs(ctx, userID, batch); err != nil {
			return fmt.Errorf("failed to delete batch: %w", err)
		}
	}

	return nil
}

// GetURLByID получает оригинальный URL по ID
func (s *urlService) GetURLByID(ctx context.Context, id string) (string, bool, error) {
	urlModel, exists := s.storage.Get(ctx, id)
	if !exists {
		return "", false, nil
	}

	if urlModel.Deleted {
		return "", true, fmt.Errorf("URL was deleted")
	}

	return urlModel.URL, true, nil
}

// GetUserURLs получает все URL пользователя
func (s *urlService) GetUserURLs(ctx context.Context, userID string) ([]models.UserURLModel, error) {
	urls, err := s.storage.GetUserURLs(ctx, userID)
	if err != nil {
		return nil, err
	}

	var userURLs []models.UserURLModel
	for _, urlModel := range urls {
		if !urlModel.Deleted {
			userURLs = append(userURLs, models.UserURLModel{
				ShortURL:    s.baseURL + "/" + urlModel.ID,
				OriginalURL: urlModel.URL,
			})
		}
	}

	return userURLs, nil
}

// generateID создает короткий идентификатор для URL
func generateID(url string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(url)))[:8]
}
