package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexuryumtsev/go-shortener/internal/app/models"
	"github.com/alexuryumtsev/go-shortener/internal/app/service/url"
	"github.com/alexuryumtsev/go-shortener/internal/app/service/user"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestGetUserURLsHandler(t *testing.T) {
	// Тестовые случаи
	tests := []struct {
		name       string
		userID     string
		setupMock  func(*url.MockURLService)
		wantStatus int
		wantURLs   []models.UserURLModel
	}{
		{
			name:   "Valid User with URLs",
			userID: "test-user",
			setupMock: func(m *url.MockURLService) {
				m.AddURL("0dd11111", "https://practicum.yandex.ru/", "test-user")
			},
			wantStatus: http.StatusOK,
			wantURLs: []models.UserURLModel{
				{
					ShortURL:    "http://localhost/0dd11111",
					OriginalURL: "https://practicum.yandex.ru/",
				},
			},
		},
		{
			name:   "User with no URLs",
			userID: "empty-user",
			setupMock: func(m *url.MockURLService) {
				// Не добавляем URLs
			},
			wantStatus: http.StatusNoContent,
			wantURLs:   nil,
		},
		{
			name:   "User with deleted URLs",
			userID: "user-with-deleted",
			setupMock: func(m *url.MockURLService) {
				m.AddURL("deleted123", "https://deleted.com/", "user-with-deleted")
				m.MarkURLAsDeleted("deleted123")
			},
			wantStatus: http.StatusOK,
			wantURLs: []models.UserURLModel{
				{
					ShortURL:    "http://localhost/deleted123",
					OriginalURL: "https://deleted.com/",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Подготовка тестового окружения
			mockURLService := url.NewMockURLService("http://localhost", nil)
			tt.setupMock(mockURLService)
			mockUserService := user.NewMockUserService(tt.userID)

			// Создание маршрутизатора и обработчика
			r := chi.NewRouter()
			handler := GetUserURLsHandler(mockURLService, mockUserService)
			r.Get("/api/user/urls", handler)

			// Создание тестового запроса
			req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
			rec := httptest.NewRecorder()

			// Выполнение запроса
			r.ServeHTTP(rec, req)

			// Проверка результатов
			assert.Equal(t, tt.wantStatus, rec.Code)

			if tt.wantStatus == http.StatusOK {
				assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

				var gotURLs []models.UserURLModel
				err := json.NewDecoder(rec.Body).Decode(&gotURLs)
				assert.NoError(t, err)
				assert.Equal(t, tt.wantURLs, gotURLs)
			}
		})
	}
}

func BenchmarkGetUserURLsHandler(b *testing.B) {
	urlCounts := []int{1, 10, 100}

	for _, count := range urlCounts {
		b.Run(fmt.Sprintf("URLs_%d", count), func(b *testing.B) {
			// Подготовка тестового окружения
			mockURLService := url.NewMockURLService("http://localhost", nil)
			userID := "test-user"

			// Добавление тестовых URL
			for i := 0; i < count; i++ {
				mockURLService.AddURL(
					fmt.Sprintf("id%d", i),
					fmt.Sprintf("https://example%d.com", i),
					userID,
				)
			}

			mockUserService := user.NewMockUserService(userID)
			handler := GetUserURLsHandler(mockURLService, mockUserService)

			r := chi.NewRouter()
			r.Get("/api/user/urls", handler)

			b.ResetTimer()
			b.StopTimer()

			for i := 0; i < b.N; i++ {
				req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
				rec := httptest.NewRecorder()

				b.StartTimer()
				r.ServeHTTP(rec, req)
				b.StopTimer()

				if rec.Code != http.StatusOK {
					b.Fatalf("handler returned wrong status code: got %v want %v",
						rec.Code, http.StatusOK)
				}

				// Очистка ответа
				res := rec.Result()
				io.Copy(io.Discard, res.Body)
				res.Body.Close()
			}
		})
	}
}
