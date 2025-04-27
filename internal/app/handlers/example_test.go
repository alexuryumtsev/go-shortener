package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/alexuryumtsev/go-shortener/config"
	"github.com/alexuryumtsev/go-shortener/internal/app/handlers"
	"github.com/alexuryumtsev/go-shortener/internal/app/models"
	"github.com/alexuryumtsev/go-shortener/internal/app/service/user"
	"github.com/alexuryumtsev/go-shortener/internal/app/storage"
)

func ExamplePostHandler() {
	// Инициализация тестового окружения
	repo := storage.NewMockStorage()
	userService := user.NewMockUserService("test-user")
	cfg := &config.Config{BaseURL: "http://localhost:8080"}

	// Создание обработчика
	handler := handlers.PostHandler(repo, userService, cfg)

	// Создание тестового запроса
	longURL := "https://practicum.yandex.ru"
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(longURL))
	w := httptest.NewRecorder()

	// Выполнение запроса
	handler.ServeHTTP(w, req)

	// Получение результата
	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	fmt.Printf("Status: %d\nShort URL: %s\n", resp.StatusCode, body)
	// Output:
	// Status: 201
	// Short URL: http://localhost:8080/6bdb5b0e
}

func ExamplePostJSONHandler() {
	// Инициализация тестового окружения
	repo := storage.NewMockStorage()
	userService := user.NewMockUserService("test-user")
	cfg := &config.Config{BaseURL: "http://localhost:8080"}

	// Создание обработчика
	handler := handlers.PostJSONHandler(repo, userService, cfg)

	// Подготовка JSON запроса
	reqBody := models.RequestBody{URL: "https://practicum.yandex.ru"}
	jsonData, _ := json.Marshal(reqBody)

	// Создание тестового запроса
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Выполнение запроса
	handler.ServeHTTP(w, req)

	// Получение результата
	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	fmt.Printf("Status: %d\nResponse: %s\n", resp.StatusCode, body)
	// Output:
	// Status: 201
	// Response: {"result":"http://localhost:8080/6bdb5b0e"}
}

func ExamplePostBatchHandler() {
	// Инициализация тестового окружения
	repo := storage.NewMockStorage()
	userService := user.NewMockUserService("test-user")
	cfg := &config.Config{BaseURL: "http://localhost:8080"}

	// Создание обработчика
	handler := handlers.PostBatchHandler(repo, userService, cfg)

	// Подготовка пакетного запроса
	batchRequest := []models.URLBatchModel{
		{
			CorrelationID: "1",
			OriginalURL:   "https://practicum.yandex.ru",
		},
		{
			CorrelationID: "2",
			OriginalURL:   "https://ya.ru",
		},
	}
	jsonData, _ := json.Marshal(batchRequest)

	// Создание тестового запроса
	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Выполнение запроса
	handler.ServeHTTP(w, req)

	// Получение результата
	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	fmt.Printf("Status: %d\nResponse: %s\n", resp.StatusCode, body)
	// Output:
	// Status: 201
	// Response: [{"correlation_id":"1","short_url":"http://localhost:8080/6bdb5b0e"},{"correlation_id":"2","short_url":"http://localhost:8080/e98192e1"}]
}

func ExampleGetUserURLsHandler() {
	// Инициализация тестового окружения
	repo := storage.NewMockStorage()
	userService := user.NewMockUserService("test-user")
	cfg := &config.Config{BaseURL: "http://localhost:8080"}

	// Добавление тестовых данных
	repo.Save(context.Background(), models.URLModel{
		ID:     "abcdef12",
		URL:    "https://practicum.yandex.ru",
		UserID: "test-user",
	})

	// Создание обработчика
	handler := handlers.GetUserURLsHandler(repo, userService, cfg)

	// Создание тестового запроса
	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	w := httptest.NewRecorder()

	// Выполнение запроса
	handler.ServeHTTP(w, req)

	// Получение результата
	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	fmt.Printf("Status: %d\nResponse: %s\n", resp.StatusCode, body)
	// Output:
	// Status: 200
	// Response: [{"short_url":"http://localhost:8080/abcdef12","original_url":"https://practicum.yandex.ru"}]
}
