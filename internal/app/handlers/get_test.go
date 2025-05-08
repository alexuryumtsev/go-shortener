package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexuryumtsev/go-shortener/internal/app/models"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

// MockURLServiceForGet - мок-реализация URLService для тестирования Get
type MockURLServiceForGet struct {
	urls map[string]string
}

func NewMockURLServiceForGet() *MockURLServiceForGet {
	return &MockURLServiceForGet{
		urls: map[string]string{
			"0dd11111": "https://practicum.yandex.ru/",
		},
	}
}

func (m *MockURLServiceForGet) ShortenerURL(ctx context.Context, originalURL, userID string) (string, error) {
	return "http://localhost/shortid", nil
}

func (m *MockURLServiceForGet) SaveBatchShortenerURL(ctx context.Context, batchModels []models.URLBatchModel, userID string) ([]models.BatchResponseModel, error) {
	return nil, nil
}

func (m *MockURLServiceForGet) DeleteUserURLsBatch(ctx context.Context, userID string, shortURLs []string) error {
	return nil
}

func (m *MockURLServiceForGet) GetURLByID(ctx context.Context, id string) (string, bool, error) {
	url, exists := m.urls[id]
	if !exists {
		return "", false, nil
	}
	return url, true, nil
}

func (m *MockURLServiceForGet) GetUserURLs(ctx context.Context, userID string) ([]models.UserURLModel, error) {
	return nil, nil
}

func TestGetHandler(t *testing.T) {
	// Инициализация мок-сервиса и маршрутизатора
	mockURLService := NewMockURLServiceForGet()
	r := chi.NewRouter()
	r.Get("/{id}", GetHandler(mockURLService))

	type want struct {
		code        int
		header      string
		contentType string
	}

	testCases := []struct {
		name        string
		requestPath string
		want        want
	}{
		{
			name:        "Valid ID",
			requestPath: "/0dd11111",
			want: want{
				code:        http.StatusTemporaryRedirect,
				header:      "https://practicum.yandex.ru/",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:        "Invalid ID",
			requestPath: "/1111",
			want: want{
				code:        http.StatusNotFound,
				header:      "",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// тестовый HTTP-запрос
			req := httptest.NewRequest(http.MethodGet, tc.requestPath, nil)
			rec := httptest.NewRecorder()

			// Отправляем запрос через маршрутизатор
			r.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, tc.want.code, res.StatusCode)
			assert.Equal(t, tc.want.header, rec.Header().Get("Location"))

			if tc.name != "Valid ID" {
				assert.Equal(t, tc.want.contentType, res.Header.Get("Content-Type"))
			}
		})
	}
}

func BenchmarkGetHandler(b *testing.B) {
	// Подготовка тестового окружения
	mockURLService := NewMockURLServiceForGet()
	r := chi.NewRouter()
	r.Get("/{id}", GetHandler(mockURLService))

	benchmarks := []struct {
		name        string
		requestPath string
		wantStatus  int
	}{
		{
			name:        "Existing URL",
			requestPath: "/0dd11111",
			wantStatus:  http.StatusTemporaryRedirect,
		},
		{
			name:        "Non-existing URL",
			requestPath: "/nonexistent",
			wantStatus:  http.StatusNotFound,
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			// Сброс таймера перед началом измерений
			b.StopTimer()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				req := httptest.NewRequest(http.MethodGet, bm.requestPath, nil)
				rr := httptest.NewRecorder()

				b.StartTimer() // Начинаем измерение
				r.ServeHTTP(rr, req)
				b.StopTimer() // Останавливаем измерение

				// Проверяем статус код для уверенности в корректной работе
				if status := rr.Code; status != bm.wantStatus {
					b.Fatalf("handler returned wrong status code: got %v want %v",
						status, bm.wantStatus)
				}
			}
		})
	}
}
