package handlers

import (
	"net/http"

	"github.com/alexuryumtsev/go-shortener/internal/app/storage"
	"github.com/go-chi/chi/v5"
)

// GetHandler обрабатывает GET-запросы с динамическими id.
func GetHandler(storage storage.URLReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		ctx := r.Context()
		urlModel, exists := storage.Get(ctx, id)
		if !exists {
			http.Error(w, "URL not found", http.StatusNotFound)
			return
		}

		// Ответ с редиректом на оригинальный URL.
		w.Header().Set("Location", urlModel.URL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}
