package url

import (
	"context"
	"testing"

	"github.com/alexuryumtsev/go-shortener/internal/app/models"
	"github.com/alexuryumtsev/go-shortener/internal/app/storage"
	"github.com/stretchr/testify/assert"
)

func TestURLService(t *testing.T) {
	mockStorage := storage.NewMockStorage()
	baseURL := "http://localhost:8080"
	batchSize := 10
	service := NewURLService(mockStorage, baseURL, batchSize)
	ctx := context.Background()

	t.Run("ShortenerURL", func(t *testing.T) {
		originalURL := "https://example.com"
		userID := "test-user"

		shortURL, err := service.ShortenerURL(ctx, originalURL, userID)
		assert.NoError(t, err)
		assert.Contains(t, shortURL, baseURL)
	})

	t.Run("SaveBatchShortenerURL", func(t *testing.T) {
		batch := []models.URLBatchModel{
			{CorrelationID: "1", OriginalURL: "https://example1.com"},
			{CorrelationID: "2", OriginalURL: "https://example2.com"},
		}
		userID := "test-user"

		responseModels, err := service.SaveBatchShortenerURL(ctx, batch, userID)
		assert.NoError(t, err)
		assert.Len(t, responseModels, 2)
		for _, model := range responseModels {
			assert.Contains(t, model.ShortURL, baseURL)
		}
	})

	t.Run("GetURLByID", func(t *testing.T) {
		// Сначала сохраним URL
		originalURL := "https://example.com"
		userID := "test-user"
		shortURL, err := service.ShortenerURL(ctx, originalURL, userID)
		assert.NoError(t, err)

		// Извлекаем ID из короткого URL
		id := shortURL[len(baseURL)+1:]

		// Теперь пробуем получить URL по ID
		retrievedURL, exists, err := service.GetURLByID(ctx, id)
		assert.True(t, exists)
		assert.NoError(t, err)
		assert.Equal(t, originalURL, retrievedURL)
	})

	t.Run("GetUserURLs", func(t *testing.T) {
		userID := "test-user"

		// Получаем URL пользователя
		userURLs, err := service.GetUserURLs(ctx, userID)
		assert.NoError(t, err)
		assert.NotEmpty(t, userURLs)
	})
}
