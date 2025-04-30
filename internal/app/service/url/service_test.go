package url

import (
	"context"
	"testing"

	"github.com/alexuryumtsev/go-shortener/internal/app/models"
	"github.com/alexuryumtsev/go-shortener/internal/app/storage"
	"github.com/stretchr/testify/assert"
)

func TestURLService(t *testing.T) {
	ctx := context.Background()
	mockStorage := storage.NewMockStorage()
	baseURL := "http://localhost:8080"
	batchSize := 10
	service := NewURLService(ctx, mockStorage, baseURL, batchSize)

	t.Run("ShortenerURL", func(t *testing.T) {
		originalURL := "https://example.com"
		userID := "test-user"

		shortURL, err := service.ShortenerURL(originalURL, userID)
		assert.NoError(t, err)
		assert.Contains(t, shortURL, baseURL)
	})

	t.Run("SaveBatchShortenerURL", func(t *testing.T) {
		batch := []models.URLBatchModel{
			{CorrelationID: "1", OriginalURL: "https://example1.com"},
			{CorrelationID: "2", OriginalURL: "https://example2.com"},
		}
		userID := "test-user"

		shortURLs, err := service.SaveBatchShortenerURL(batch, userID)
		assert.NoError(t, err)
		assert.Len(t, shortURLs, 2)
		for _, url := range shortURLs {
			assert.Contains(t, url, baseURL)
		}
	})

	t.Run("DeleteUserURLsBatch", func(t *testing.T) {
		userID := "test-user"
		shortURLs := []string{"abc123", "def456"}

		err := service.DeleteUserURLsBatch(ctx, userID, shortURLs)
		assert.NoError(t, err)
	})
}
