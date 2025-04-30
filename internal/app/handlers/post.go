package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/alexuryumtsev/go-shortener/config"
	"github.com/alexuryumtsev/go-shortener/internal/app/middleware"
	"github.com/alexuryumtsev/go-shortener/internal/app/models"
	"github.com/alexuryumtsev/go-shortener/internal/app/service/url"
	"github.com/alexuryumtsev/go-shortener/internal/app/service/user"
	"github.com/alexuryumtsev/go-shortener/internal/app/storage"
)

// PostHandler обрабатывает POST-запросы для создания короткого URL.
//
// Принимает:
//   - Тело запроса: оригинальный URL в текстовом формате
//
// Возвращает:
//   - В случае успеха:
//     Код: 201 Created
//     Тело: сокращённый URL в текстовом формате
//   - В случае ошибки:
//     Код: 400 Bad Request - при невалидном теле запроса
//     Код: 409 Conflict - если URL уже существует
//     Код: 500 Internal Server Error - при внутренних ошибках сервера
func PostHandler(storage storage.URLWriter, userService user.UserService, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		ctx := r.Context()
		originalURL := strings.TrimSpace(string(body))
		userID := userService.GetUserIDFromCookie(r)
		shortenedURL, err := url.NewURLService(ctx, storage, cfg.BaseURL, cfg.BatchSize).ShortenerURL(originalURL, userID)

		if err != nil {
			middleware.ProcessError(w, err, shortenedURL, true)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortenedURL))
	}
}

// PostJSONHandler обрабатывает POST-запросы для создания короткого URL в формате JSON.
//
// Принимает:
//   - Content-Type: application/json
//   - Тело запроса: {"url": "https://example.com"}
//
// Возвращает:
//   - В случае успеха:
//     Код: 201 Created
//     Content-Type: application/json
//     Тело: {"result": "http://shortener.com/abcdef"}
//   - В случае ошибки:
//     Код: 400 Bad Request - при невалидном JSON или отсутствии обязательных полей
//     Код: 409 Conflict - если URL уже существует
//     Код: 500 Internal Server Error - при внутренних ошибках сервера
func PostJSONHandler(storage storage.URLWriter, userService user.UserService, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.RequestBody
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		ctx := r.Context()
		userID := userService.GetUserIDFromCookie(r)
		shortenedURL, err := url.NewURLService(ctx, storage, cfg.BaseURL, cfg.BatchSize).ShortenerURL(req.URL, userID)

		if err != nil {
			middleware.ProcessError(w, err, shortenedURL, false)
			return
		}

		resp := models.ResponseBody{
			ShortURL: shortenedURL,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	}
}

// PostBatchHandler обрабатывает POST-запросы для создания множества коротких URL.
//
// Принимает:
//   - Content-Type: application/json
//   - Тело запроса: [
//     {"correlation_id": "1", "original_url": "https://example1.com"},
//     {"correlation_id": "2", "original_url": "https://example2.com"}
//     ]
//
// Возвращает:
//   - В случае успеха:
//     Код: 201 Created
//     Content-Type: application/json
//     Тело: [
//     {"correlation_id": "1", "short_url": "http://shortener.com/abc"},
//     {"correlation_id": "2", "short_url": "http://shortener.com/def"}
//     ]
//   - В случае ошибки:
//     Код: 400 Bad Request - при невалидном JSON, пустом батче или отсутствии обязательных полей
//     Код: 409 Conflict - если один из URL уже существует
//     Код: 500 Internal Server Error - при внутренних ошибках сервера
func PostBatchHandler(repo storage.URLStorage, userService user.UserService, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		baseURL := strings.TrimSuffix(cfg.BaseURL, "/")

		var batchModels []models.URLBatchModel
		if err := json.NewDecoder(r.Body).Decode(&batchModels); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if len(batchModels) == 0 {
			http.Error(w, "Empty batch", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		urlService := url.NewURLService(ctx, repo, baseURL, cfg.BatchSize)
		userID := userService.GetUserIDFromCookie(r)
		shortenedURLs, err := urlService.SaveBatchShortenerURL(batchModels, userID)

		if err != nil {
			middleware.ProcessError(w, err, "", false)
			return
		}

		var batchResponseModels []models.BatchResponseModel
		for i, shortenedURL := range shortenedURLs {
			batchResponseModels = append(batchResponseModels, models.BatchResponseModel{
				CorrelationID: batchModels[i].CorrelationID,
				ShortURL:      shortenedURL,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(batchResponseModels); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}
