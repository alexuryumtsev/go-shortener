package handlers

import (
	"log"
	"net/http"

	"github.com/alexuryumtsev/go-shortener/internal/app/storage"
	"github.com/go-chi/chi/v5"
)

// GetHandler обрабатывает GET-запросы с динамическими id.
//
// Принимает:
//   - URL параметр: id - короткий идентификатор URL
//
// Возвращает:
//   - В случае успеха:
//     Код: 307 Temporary Redirect
//     Заголовок: Location содержит оригинальный URL для редиректа
//   - В случае ошибки:
//     Код: 404 Not Found - если URL не найден
//     Код: 410 Gone - если URL был удален владельцем
func GetHandler(storage storage.URLReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		ctx := r.Context()
		urlModel, exists := storage.Get(ctx, id)
		if !exists {
			http.Error(w, "URL not found", http.StatusNotFound)
			return
		}

		log.Printf("Redirecting to %v", urlModel)

		if urlModel.Deleted {
			http.Error(w, "This URL is no longer available as it has been deleted by the owner", http.StatusGone)
			return
		}

		// Ответ с редиректом на оригинальный URL.
		w.Header().Set("Location", urlModel.URL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}
