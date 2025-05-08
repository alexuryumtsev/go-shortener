package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/alexuryumtsev/go-shortener/internal/app/service/url"
	"github.com/alexuryumtsev/go-shortener/internal/app/service/user"
)

// GetUserURLsHandler возвращает все URL текущего пользователя.
func GetUserURLsHandler(urlService url.URLService, userService user.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Создаем контекст с таймаутом
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Получаем ID пользователя
		userID := userService.GetUserIDFromCookie(r)
		if userID == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Вызываем бизнес-логику
		userURLs, err := urlService.GetUserURLs(ctx, userID)

		if err != nil {
			http.Error(w, "Failed to get user URLs", http.StatusInternalServerError)
			return
		}

		if len(userURLs) == 0 {
			// Если у пользователя нет URL, возвращаем 204 No Content
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Отправляем ответ
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(userURLs); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}

// DeleteUserURLsHandler удаляет URL пользователя.
func DeleteUserURLsHandler(urlService url.URLService, userService user.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Получаем ID пользователя
		userID := userService.GetUserIDFromCookie(r)
		if userID == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Декодируем JSON с URLs для удаления
		var shortURLs []string
		if err := json.NewDecoder(r.Body).Decode(&shortURLs); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Запускаем асинхронный процесс удаления
		go func() {
			// Создаем отдельный контекст для асинхронной операции
			deleteCtx, deleteCancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer deleteCancel()

			if err := urlService.DeleteUserURLsBatch(deleteCtx, userID, shortURLs); err != nil {
				if err := urlService.DeleteUserURLsBatch(r.Context(), userID, shortURLs); err != nil {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
			}
		}()

		// Сразу возвращаем статус 202, т.к. удаление асинхронное
		w.WriteHeader(http.StatusAccepted)
	}
}
