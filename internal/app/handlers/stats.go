package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/alexuryumtsev/go-shortener/internal/app/service/url"
	"github.com/alexuryumtsev/go-shortener/internal/app/storage"
)

// StatsResponse представляет ответ со статистикой
type StatsResponse struct {
	URLs  int `json:"urls"`
	Users int `json:"users"`
}

// GetStatsHandler возвращает статистику сервиса
func GetStatsHandler(urlService url.URLService, urlStorage storage.URLStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Создаем контекст с таймаутом
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Получаем статистику
		stats, err := urlStorage.GetStats(ctx)
		if err != nil {
			http.Error(w, "Failed to get statistics", http.StatusInternalServerError)
			return
		}

		// Формируем ответ
		response := StatsResponse{
			URLs:  stats.URLs,
			Users: stats.Users,
		}

		// Отправляем ответ
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}
