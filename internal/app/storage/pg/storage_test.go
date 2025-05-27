package pg_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/alexuryumtsev/go-shortener/internal/app/db"
	"github.com/alexuryumtsev/go-shortener/internal/app/models"
	"github.com/alexuryumtsev/go-shortener/internal/app/storage/pg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *db.Database {
	// Получаем строку подключения из переменной окружения или используем тестовую базу
	connString := os.Getenv("DATABASE_CONN_STRING")
	if connString == "" {
		// connString = "postgres://postgres:postgres@127.0.0.1:5432/praktikum?sslmode=disable"
		connString = "postgres://postgres:postgres@postgres:5432/praktikum?sslmode=disable"
	}

	// Выполняем миграции (предполагаем, что таблицы уже созданы или создаем их)
	ctx := context.Background()

	// Создаем экземпляр базы данных
	database, err := db.NewDatabaseConnection(ctx, connString)
	require.NoError(t, err, "Failed to connect to test database")

	_, err = database.Pool.Exec(ctx, `
		DROP TABLE IF EXISTS urls;
		CREATE TABLE IF NOT EXISTS urls (
			id SERIAL PRIMARY KEY,
			user_id TEXT NOT NULL,
			short_url TEXT NOT NULL UNIQUE,
			original_url TEXT NOT NULL,
			is_deleted BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
	`)
	require.NoError(t, err, "Failed to setup test database schema")

	return database
}

func TestDatabaseStorage_Integration(t *testing.T) {
	// Настраиваем тестовую базу данных
	database := setupTestDB(t)
	defer database.Close()

	// Создаем хранилище
	storage := pg.NewDatabaseStorage(database)
	ctx := context.Background()

	// Генерируем тестовые данные
	userID := "test-user-id"
	urlModel := models.URLModel{
		UserID: userID,
		ID:     "abc123",
		URL:    "https://example.com",
	}

	// Тест 1: Проверка функции Save
	t.Run("Save", func(t *testing.T) {
		err := storage.Save(ctx, urlModel)
		assert.NoError(t, err, "Failed to save URL")

		// Проверяем конфликт при повторном сохранении
		err = storage.Save(ctx, urlModel)
		assert.Error(t, err, "Expected conflict error on duplicate save")
	})

	// Тест 2: Проверка функции Get
	t.Run("Get", func(t *testing.T) {
		fetchedURL, found := storage.Get(ctx, urlModel.ID)
		assert.True(t, found, "URL should be found")
		assert.Equal(t, urlModel.URL, fetchedURL.URL, "Original URL should match")
		assert.False(t, fetchedURL.Deleted, "URL should not be marked as deleted")

		// Проверяем несуществующий URL
		_, found = storage.Get(ctx, "non-existent-id")
		assert.False(t, found, "Non-existent URL should not be found")
	})

	// Тест 3: Проверка функции SaveBatch
	t.Run("SaveBatch", func(t *testing.T) {
		batchURLs := []models.URLModel{
			{
				UserID: userID,
				ID:     "batch1",
				URL:    "https://example.com/batch1",
			},
			{
				UserID: userID,
				ID:     "batch2",
				URL:    "https://example.com/batch2",
			},
		}

		err := storage.SaveBatch(ctx, batchURLs)
		assert.NoError(t, err, "Failed to save batch URLs")

		// Проверяем, что все URL из пакета сохранены
		for _, urlModel := range batchURLs {
			fetchedURL, found := storage.Get(ctx, urlModel.ID)
			assert.True(t, found, "Batch URL should be found")
			assert.Equal(t, urlModel.URL, fetchedURL.URL, "Original URL should match")
		}
	})

	// Тест 4: Проверка функции GetUserURLs
	t.Run("GetUserURLs", func(t *testing.T) {
		userURLs, err := storage.GetUserURLs(ctx, userID)
		assert.NoError(t, err, "Failed to get user URLs")

		// У пользователя должно быть не менее 3 URL (1 + 2 из пакета)
		assert.GreaterOrEqual(t, len(userURLs), 3, "User should have at least 3 URLs")

		// Проверяем, что все URL принадлежат пользователю
		for _, url := range userURLs {
			assert.NotEmpty(t, url.ID, "URL ID should not be empty")
			assert.NotEmpty(t, url.URL, "Original URL should not be empty")
		}

		// Проверяем URLs для несуществующего пользователя
		emptyURLs, err := storage.GetUserURLs(ctx, "non-existent-user")
		assert.NoError(t, err, "Should not error for non-existent user")
		assert.Empty(t, emptyURLs, "Non-existent user should have no URLs")
	})

	// Тест 5: Проверка функции DeleteUserURLs
	t.Run("DeleteUserURLs", func(t *testing.T) {
		// Удаляем первый URL
		err := storage.DeleteUserURLs(ctx, userID, []string{urlModel.ID})
		assert.NoError(t, err, "Failed to delete URL")

		// Проверяем, что URL помечен как удаленный
		fetchedURL, found := storage.Get(ctx, urlModel.ID)
		assert.True(t, found, "URL should still be found after deletion")
		assert.True(t, fetchedURL.Deleted, "URL should be marked as deleted")

		// Проверяем, что удаленный URL не возвращается в GetUserURLs
		userURLs, err := storage.GetUserURLs(ctx, userID)
		assert.NoError(t, err, "Failed to get user URLs after deletion")

		// Проверяем, что удаленный URL не входит в список
		for _, url := range userURLs {
			assert.NotEqual(t, urlModel.ID, url.ID, "Deleted URL should not be in user URLs")
		}
	})

	// Тест 6: Проверка функции Ping
	t.Run("Ping", func(t *testing.T) {
		err := storage.Ping(ctx)
		assert.NoError(t, err, "Database ping should succeed")

		// Создаем контекст с таймаутом для проверки таймаута соединения
		timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
		defer cancel()
		err = storage.Ping(timeoutCtx)
		assert.Error(t, err, "Ping with timeout should fail")
	})
}
