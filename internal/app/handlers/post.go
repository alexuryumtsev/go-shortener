package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/alexuryumtsev/go-shortener/internal/app/service"
	"github.com/alexuryumtsev/go-shortener/internal/app/storage"
)

// PostHandler обрабатывает POST-запросы.
func PostHandler(storage storage.URLWriter, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)

		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		ctx := r.Context()
		originalURL := strings.TrimSpace(string(body))
		shortenedURL, err := service.NewURLService(ctx, storage, baseURL).ShortenerURL(originalURL)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortenedURL))
	}
}

// RequestBody определяет структуру входных данных.
type RequestBody struct {
	URL string `json:"url"`
}

// ResponseBody определяет структуру ответа.
type ResponseBody struct {
	ShortURL string `json:"result"`
}

// PostHandler обрабатывает POST-запросы для создания коротких URL.
func PostJSONHandler(storage storage.URLWriter, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RequestBody
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		ctx := r.Context()
		shortenedURL, err := service.NewURLService(ctx, storage, baseURL).ShortenerURL(req.URL)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp := ResponseBody{
			ShortURL: shortenedURL,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	}
}
