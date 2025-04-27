package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexuryumtsev/go-shortener/config"
	"github.com/alexuryumtsev/go-shortener/internal/app/models"
	"github.com/alexuryumtsev/go-shortener/internal/app/service/user"
	"github.com/alexuryumtsev/go-shortener/internal/app/storage"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestGetUserURLsHandler(t *testing.T) {
	// тестовое хранилище и добавляем тестовые данные.
	baseURL := "http://localhost"
	config := &config.Config{BaseURL: baseURL}
	userID := "test-user"
	repo := storage.NewMockStorage()
	repo.Save(context.Background(), models.URLModel{ID: "0dd11111", URL: "https://practicum.yandex.ru/", UserID: userID})

	// Инициализация маршрутизатора.
	r := chi.NewRouter()
	mockUserService := user.NewMockUserService(userID)
	r.Get("/api/user/urls", GetUserURLsHandler(repo, mockUserService, config))

	type want struct {
		code        int
		body        []UserURL
		contentType string
	}

	testCases := []struct {
		name   string
		userID string
		want   want
	}{
		{
			name:   "Valid User ID",
			userID: userID,
			want: want{
				code: http.StatusOK,
				body: []UserURL{
					{
						ShortURL:    baseURL + "/0dd11111",
						OriginalURL: "https://practicum.yandex.ru/",
					},
				},
				contentType: "application/json",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// тестовый HTTP-запрос.
			req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
			rec := httptest.NewRecorder()

			// Отправляем запрос через маршрутизатор.
			r.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, tc.want.code, res.StatusCode)
			assert.Equal(t, tc.want.contentType, res.Header.Get("Content-Type"))

			if tc.want.body != nil {
				var resBody []UserURL
				err := json.NewDecoder(res.Body).Decode(&resBody)
				assert.NoError(t, err)
				assert.Equal(t, tc.want.body, resBody)
			}
		})
	}
}

func BenchmarkGetUserURLsHandler(b *testing.B) {
	// Подготовка тестовых данных
	baseURL := "http://localhost"
	config := &config.Config{BaseURL: baseURL}
	userID := "test-user"
	repo := storage.NewMockStorage()

	// Добавляем разное количество URL для тестирования производительности
	urlCounts := []int{1, 10, 100}

	for _, count := range urlCounts {
		// Добавляем указанное количество URL
		for i := 0; i < count; i++ {
			repo.Save(context.Background(), models.URLModel{
				ID:     fmt.Sprintf("id%d", i),
				URL:    fmt.Sprintf("https://example%d.com", i),
				UserID: userID,
			})
		}

		b.Run(fmt.Sprintf("URLs_%d", count), func(b *testing.B) {
			// Инициализация маршрутизатора
			r := chi.NewRouter()
			mockUserService := user.NewMockUserService(userID)
			r.Get("/api/user/urls", GetUserURLsHandler(repo, mockUserService, config))

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
				rr := httptest.NewRecorder()

				r.ServeHTTP(rr, req)

				// Проверяем статус код
				if status := rr.Code; status != http.StatusOK {
					b.Fatalf("handler returned wrong status code: got %v want %v",
						status, http.StatusOK)
				}

				// Очищаем тело ответа
				res := rr.Result()
				io.Copy(io.Discard, res.Body)
				res.Body.Close()
			}
		})

		// Очищаем хранилище перед следующим тестом
		repo = storage.NewMockStorage()
	}
}
