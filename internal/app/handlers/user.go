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
