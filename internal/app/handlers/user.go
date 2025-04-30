package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/alexuryumtsev/go-shortener/config"
	"github.com/alexuryumtsev/go-shortener/internal/app/service/url"
	"github.com/alexuryumtsev/go-shortener/internal/app/service/user"
	"github.com/alexuryumtsev/go-shortener/internal/app/storage"
)

type UserURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// GetUserURLsHandler возвращает все URL текущего пользователя.
//
// Принимает:
//   - Cookie: auth_token - JWT токен для аутентификации пользователя
//
// Возвращает:
//   - В случае успеха:
//     Код: 200 OK
//     Content-Type: application/json
//     Тело: [
//     {"short_url": "http://shortener.com/abc", "original_url": "https://example1.com"},
//     {"short_url": "http://shortener.com/def", "original_url": "https://example2.com"}
//     ]
//   - Если URL не найдены:
//     Код: 204 No Content
//   - В случае ошибки:
//     Код: 401 Unauthorized - если токен отсутствует или невалиден
//     Код: 500 Internal Server Error - при внутренних ошибках сервера
func GetUserURLsHandler(repo storage.URLStorage, userService user.UserService, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := userService.GetUserIDFromCookie(r)
		urls, err := repo.GetUserURLs(r.Context(), userID)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if len(urls) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		baseURL := strings.TrimSuffix(cfg.BaseURL, "/")
		var userURLs []UserURL
		for _, url := range urls {
			if url.Deleted {
				continue
			}
			userURLs = append(userURLs, UserURL{
				ShortURL:    baseURL + "/" + url.ID,
				OriginalURL: url.URL,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(userURLs)
	}
}

// DeleteUserURLsHandler удаляет URL пользователя.
//
// Принимает:
//   - Cookie: auth_token - JWT токен для аутентификации пользователя
//   - Content-Type: application/json
//   - Тело: ["url_id1", "url_id2", ...]
//
// Возвращает:
//   - В случае успеха:
//     Код: 202 Accepted
//   - В случае ошибки:
//     Код: 401 Unauthorized - если токен отсутствует или невалиден
//     Код: 400 Bad Request - при невалидном JSON
//     Код: 500 Internal Server Error - при внутренних ошибках сервера
func DeleteUserURLsHandler(urlService url.URLService, userService user.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var shortURLs []string
		if err := json.NewDecoder(r.Body).Decode(&shortURLs); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		userID := userService.GetUserIDFromCookie(r)

		if err := urlService.DeleteUserURLsBatch(r.Context(), userID, shortURLs); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}
