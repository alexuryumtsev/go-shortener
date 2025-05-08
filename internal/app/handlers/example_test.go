package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/alexuryumtsev/go-shortener/internal/app/handlers"
	"github.com/alexuryumtsev/go-shortener/internal/app/models"
	"github.com/alexuryumtsev/go-shortener/internal/app/service/user"
)

// MockURLService - мок-реализация URLService для примеров
type MockURLService struct{}

func (m *MockURLService) ShortenerURL(ctx context.Context, originalURL, userID string) (string, error) {
	// Фиксированный ответ для примера
	return "http://localhost:8080/6bdb5b0e", nil
}

func (m *MockURLService) SaveBatchShortenerURL(ctx context.Context, batchModels []models.URLBatchModel, userID string) ([]models.BatchResponseModel, error) {
	// Фиксированный ответ для примера
	var response []models.BatchResponseModel
	for i, model := range batchModels {
		id := fmt.Sprintf("6bdb5b0e%d", i)
		if i == 1 {
			id = "e98192e1" // Особый ID для второго элемента
		}
		response = append(response, models.BatchResponseModel{
			CorrelationID: model.CorrelationID,
			ShortURL:      fmt.Sprintf("http://localhost:8080/%s", id),
		})
	}
	return response, nil
}

func (m *MockURLService) DeleteUserURLsBatch(ctx context.Context, userID string, shortURLs []string) error {
	return nil
}

func (m *MockURLService) GetURLByID(ctx context.Context, id string) (string, bool, error) {
	return "https://practicum.yandex.ru", true, nil
}

func (m *MockURLService) GetUserURLs(ctx context.Context, userID string) ([]models.UserURLModel, error) {
	return []models.UserURLModel{
		{
			ShortURL:    "http://localhost:8080/abcdef12",
			OriginalURL: "https://practicum.yandex.ru",
		},
	}, nil
}

func ExamplePostHandler() {
	// Инициализация тестового окружения
	userService := user.NewMockUserService("test-user")
	urlService := &MockURLService{}

	// Создание обработчика
	handler := handlers.PostHandler(urlService, userService)

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
	userService := user.NewMockUserService("test-user")
	urlService := &MockURLService{}

	// Создание обработчика
	handler := handlers.PostJSONHandler(urlService, userService)

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
	userService := user.NewMockUserService("test-user")
	urlService := &MockURLService{}

	// Создание обработчика
	handler := handlers.PostBatchHandler(urlService, userService)

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
	// Response: [{"correlation_id":"1","short_url":"http://localhost:8080/6bdb5b0e0"},{"correlation_id":"2","short_url":"http://localhost:8080/e98192e1"}]
}

func ExampleGetUserURLsHandler() {
	// Инициализация тестового окружения
	userService := user.NewMockUserService("test-user")
	urlService := &MockURLService{}

	// Создание обработчика
	handler := handlers.GetUserURLsHandler(urlService, userService)

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
