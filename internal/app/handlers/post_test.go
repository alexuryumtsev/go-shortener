package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alexuryumtsev/go-shortener/internal/app/models"
	"github.com/alexuryumtsev/go-shortener/internal/app/service/url"
	"github.com/alexuryumtsev/go-shortener/internal/app/service/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostHandler(t *testing.T) {
	userID := "test-user"
	mockUserService := user.NewMockUserService(userID)
	mockURLService := url.NewMockURLService("http://localhost:8080/", nil)
	handler := PostHandler(mockURLService, mockUserService)

	type want struct {
		code        int
		body        string
		contentType string
	}

	testCases := []struct {
		name     string
		inputURL string
		want     want
	}{
		{
			name:     "Valid URL",
			inputURL: "https://practicum.yandex.ru/",
			want: want{
				code:        http.StatusCreated,
				body:        "http://localhost:8080/",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:     "Empty URL",
			inputURL: "",
			want: want{
				code:        http.StatusBadRequest,
				body:        "empty URL",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// тестовый HTTP-запрос.
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tc.inputURL))
			rec := httptest.NewRecorder()
			handler(rec, req)

			res := rec.Result()
			defer res.Body.Close()
			assert.Equal(t, tc.want.code, res.StatusCode)

			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			assert.True(t, strings.HasPrefix(string(resBody), tc.want.body))
		})
	}
}

func TestPostJsonHandler(t *testing.T) {
	userID := "test-user"

	// тестовое хранилище.
	mockUserService := user.NewMockUserService(userID)
	mockURLService := url.NewMockURLService("http://localhost:8080/", nil)
	handler := PostJSONHandler(mockURLService, mockUserService)

	type want struct {
		code         int
		body         models.RequestBody
		expectedBody models.ResponseBody
		contentType  string
	}

	testCases := []struct {
		name     string
		inputURL string
		want     want
	}{
		{
			name: "Valid URL",
			want: want{
				code: http.StatusCreated,
				body: models.RequestBody{
					URL: "https://practicum.yandex.ru/",
				},
				expectedBody: models.ResponseBody{
					ShortURL: "http://localhost:8080/",
				},
				contentType: "Content-Type: application/json",
			},
		},
		{
			name: "Invalid request body",
			want: want{
				code:         http.StatusBadRequest,
				body:         models.RequestBody{},
				expectedBody: models.ResponseBody{},
				contentType:  "Content-Type: application/json",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// тестовый HTTP-запрос.
			body, _ := json.Marshal(tc.want.body)
			req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()

			handler(rec, req)

			res := rec.Result()
			defer res.Body.Close()
			assert.Equal(t, tc.want.code, res.StatusCode)

			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			var resp models.ResponseBody

			json.Unmarshal(resBody, &resp)

			assert.True(t, strings.HasPrefix(resp.ShortURL, tc.want.expectedBody.ShortURL))
		})
	}
}

func BenchmarkPostHandlers(b *testing.B) {
	// Подготовка тестовых данных
	userID := "test-user"
	mockUserService := user.NewMockUserService(userID)
	mockURLService := url.NewMockURLService("http://localhost:8080/", nil)

	// Подготовка обработчиков
	plainHandler := PostHandler(mockURLService, mockUserService)
	jsonHandler := PostJSONHandler(mockURLService, mockUserService)

	// Тестовые данные
	testURL := "https://practicum.yandex.ru/"
	jsonBody := models.RequestBody{URL: testURL}
	jsonData, _ := json.Marshal(jsonBody)

	benchmarks := []struct {
		name    string
		handler http.HandlerFunc
		path    string
		body    []byte
	}{
		{
			name:    "Plain POST Handler",
			handler: plainHandler,
			path:    "/",
			body:    []byte(testURL),
		},
		{
			name:    "JSON POST Handler",
			handler: jsonHandler,
			path:    "/api/shorten",
			body:    jsonData,
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			// Сброс таймера перед измерениями
			b.StopTimer() // Останавливаем таймер перед началом итерации
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// Создаем новый запрос для каждой итерации
				req := httptest.NewRequest(http.MethodPost, bm.path, bytes.NewReader(bm.body))
				if bm.name == "JSON POST Handler" {
					req.Header.Set("Content-Type", "application/json")
				}
				rr := httptest.NewRecorder()

				b.StartTimer()      // Начинаем измерение
				bm.handler(rr, req) // Выполняем запрос
				b.StopTimer()       // Останавливаем измерение

				// Проверяем статус код
				if status := rr.Code; status != http.StatusCreated {
					b.Fatalf("handler returned wrong status code: got %v want %v",
						status, http.StatusCreated)
				}

				// Очищаем тело ответа
				res := rr.Result()
				io.Copy(io.Discard, res.Body)
				res.Body.Close()
			}
		})
	}
}
