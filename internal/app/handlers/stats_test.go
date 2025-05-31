package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexuryumtsev/go-shortener/internal/app/models"
	"github.com/alexuryumtsev/go-shortener/internal/app/service/url"
	"github.com/alexuryumtsev/go-shortener/internal/app/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetStatsHandler(t *testing.T) {
	// Создаем тестовое хранилище с данными
	mockStorage := storage.NewMockStorage()
	ctx := context.Background()

	// Добавляем тестовые данные
	testData := []models.URLModel{
		{ID: "abc123", URL: "https://example1.com", UserID: "user1", Deleted: false},
		{ID: "def456", URL: "https://example2.com", UserID: "user1", Deleted: false},
		{ID: "ghi789", URL: "https://example3.com", UserID: "user2", Deleted: false},
		{ID: "jkl012", URL: "https://example4.com", UserID: "user2", Deleted: true}, // Удалённый URL
		{ID: "mno345", URL: "https://example5.com", UserID: "user3", Deleted: false},
	}

	for _, data := range testData {
		mockStorage.Save(ctx, data)
	}

	mockURLService := url.NewMockURLService("http://localhost:8080", nil)
	handler := GetStatsHandler(mockURLService, mockStorage)

	tests := []struct {
		name       string
		wantStatus int
		wantURLs   int
		wantUsers  int
	}{
		{
			name:       "Get statistics successfully",
			wantStatus: http.StatusOK,
			wantURLs:   4, // 5 total - 1 deleted = 4
			wantUsers:  3, // user1, user2, user3
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatus, res.StatusCode)
			assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

			var stats StatsResponse
			err := json.NewDecoder(res.Body).Decode(&stats)
			require.NoError(t, err)

			assert.Equal(t, tt.wantURLs, stats.URLs)
			assert.Equal(t, tt.wantUsers, stats.Users)
		})
	}
}

func BenchmarkGetStatsHandler(b *testing.B) {
	// Подготовка тестовых данных
	mockStorage := storage.NewMockStorage()
	ctx := context.Background()

	// Добавляем много тестовых данных
	for i := 0; i < 1000; i++ {
		userID := "user" + string(rune(i%10))
		mockStorage.Save(ctx, models.URLModel{
			ID:      string(rune(i)),
			URL:     "https://example.com/" + string(rune(i)),
			UserID:  userID,
			Deleted: i%5 == 0, // Каждый 5-й URL удалён
		})
	}

	mockURLService := url.NewMockURLService("http://localhost:8080", nil)
	handler := GetStatsHandler(mockURLService, mockStorage)

	b.ResetTimer()
	b.StopTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
		rec := httptest.NewRecorder()

		b.StartTimer()
		handler.ServeHTTP(rec, req)
		b.StopTimer()

		if rec.Code != http.StatusOK {
			b.Fatalf("handler returned wrong status code: got %v want %v",
				rec.Code, http.StatusOK)
		}
	}
}
