package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/alexuryumtsev/go-shortener/internal/app/service/url"
	"github.com/go-chi/chi/v5"
)

// GetHandler обрабатывает GET-запросы с динамическими id.
func GetHandler(urlService url.URLService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Создаем контекст с таймаутом
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, "URL ID is required", http.StatusBadRequest)
			return
		}

		// Вызываем бизнес-логику
		originalURL, exists, err := urlService.GetURLByID(ctx, id)

		// Обрабатываем результат
		if !exists {
			http.Error(w, "URL not found", http.StatusNotFound)
			return
		}

		if err != nil {
			http.Error(w, "This URL is no longer available as it has been deleted by the owner", http.StatusGone)
			return
		}

		// Перенаправляем на оригинальный URL
		w.Header().Set("Location", originalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}
