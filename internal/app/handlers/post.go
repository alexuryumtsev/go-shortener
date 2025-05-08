package handlers

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/alexuryumtsev/go-shortener/internal/app/middleware"
	"github.com/alexuryumtsev/go-shortener/internal/app/models"
	"github.com/alexuryumtsev/go-shortener/internal/app/service/url"
	"github.com/alexuryumtsev/go-shortener/internal/app/service/user"
)

// PostHandler обрабатывает POST-запросы для создания короткого URL.
func PostHandler(urlService url.URLService, userService user.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Создаем контекст с таймаутом
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Чтение тела запроса
		body, readErr := io.ReadAll(r.Body)
		if readErr != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		defer func() {
			if closeErr := r.Body.Close(); closeErr != nil {
				// Логируем ошибку закрытия тела запроса
				log.Printf("Error closing request body: %v", closeErr)
			}
		}()

		// Получаем данные
		originalURL := strings.TrimSpace(string(body))
		userID := userService.GetUserIDFromCookie(r)

		// Вызываем бизнес-логику
		shortenedURL, shortenerErr := urlService.ShortenerURL(ctx, originalURL, userID)

		// Обрабатываем результат
		if shortenerErr != nil {
			middleware.ProcessError(w, shortenerErr, shortenedURL, true)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortenedURL))
	}
}

// PostJSONHandler обрабатывает POST-запросы для создания короткого URL в формате JSON.
func PostJSONHandler(urlService url.URLService, userService user.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Создаем контекст с таймаутом
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Чтение и декодирование JSON
		var req models.RequestBody
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// Получаем данные
		userID := userService.GetUserIDFromCookie(r)

		// Вызываем бизнес-логику
		shortenedURL, err := urlService.ShortenerURL(ctx, req.URL, userID)

		// Обрабатываем результат
		if err != nil {
			middleware.ProcessError(w, err, shortenedURL, false)
			return
		}

		// Формируем ответ
		resp := models.ResponseBody{
			ShortURL: shortenedURL,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}

// PostBatchHandler обрабатывает POST-запросы для создания множества коротких URL.
func PostBatchHandler(urlService url.URLService, userService user.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Создаем контекст с таймаутом
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Чтение и декодирование JSON
		var batchModels []models.URLBatchModel
		if err := json.NewDecoder(r.Body).Decode(&batchModels); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// Получаем данные
		userID := userService.GetUserIDFromCookie(r)

		// Вызываем бизнес-логику
		responseModels, err := urlService.SaveBatchShortenerURL(ctx, batchModels, userID)

		// Обрабатываем результат
		if err != nil {
			middleware.ProcessError(w, err, "", false)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(responseModels); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}
