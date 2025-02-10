package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexuryumtsev/go-shortener/internal/app/models"
	"github.com/alexuryumtsev/go-shortener/internal/app/storage"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestGetHandler(t *testing.T) {
	// тестовое хранилище и добавляем тестовые данные.
	id := "0dd11111"
	repo := storage.NewMockStorage()
	repo.Save(context.Background(), models.URLModel{ID: id, URL: "https://practicum.yandex.ru/"})

	// Инициализация маршрутизатора.
	r := chi.NewRouter()
	r.Get("/{id}", GetHandler(repo))

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
			requestPath: "/" + id,
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
			// тестовый HTTP-запрос.
			req := httptest.NewRequest(http.MethodGet, tc.requestPath, nil)
			rec := httptest.NewRecorder()

			// Отправляем запрос через маршрутизатор.
			r.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			t.Log("Value:", req)

			assert.Equal(t, tc.want.code, res.StatusCode)
			assert.Equal(t, tc.want.header, rec.Header().Get("Location"))

			if tc.name != "Valid ID" {
				assert.Equal(t, tc.want.contentType, res.Header.Get("Content-Type"))
			}
		})
	}
}
