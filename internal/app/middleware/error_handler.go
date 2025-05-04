// Package middleware содержит функции промежуточной обработки HTTP-запросов.
package middleware

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/alexuryumtsev/go-shortener/internal/app/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// ErrorMiddleware — middleware для обработки ошибок.
func ErrorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			// Поймаем панику, если она возникла
			if rec := recover(); rec != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()

		// Выполнение основной логики запроса
		next.ServeHTTP(w, r)
	})
}

// ProcessError — функция для обработки ошибок в контексте работы с БД.
func ProcessError(w http.ResponseWriter, inputErr error, shortenedURL string, responseString bool) {
	var pgErr *pgconn.PgError
	if errors.As(inputErr, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)

		if responseString {
			_, writeErr := w.Write([]byte(shortenedURL))
			if writeErr != nil {
				http.Error(w, "Failed to write response", http.StatusInternalServerError)
				return
			}
			return
		}

		if encodeErr := json.NewEncoder(w).Encode(models.ResponseBody{
			ShortURL: shortenedURL,
		}); encodeErr != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}

		return
	}

	// Обработка других ошибок
	http.Error(w, inputErr.Error(), http.StatusBadRequest)
}
