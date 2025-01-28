package router

import (
	"log"

	"github.com/alexuryumtsev/go-shortener/config"
	"github.com/alexuryumtsev/go-shortener/internal/app/compress"
	"github.com/alexuryumtsev/go-shortener/internal/app/handlers"
	"github.com/alexuryumtsev/go-shortener/internal/app/logger"
	"github.com/alexuryumtsev/go-shortener/internal/app/storage"
	"github.com/go-chi/chi/v5"
)

// ShortenerRouter создает маршруты для приложения.
func ShortenerRouter(cfg *config.Config) chi.Router {
	// Инициализация хранилища.
	var repo storage.URLStorage = storage.NewStorage(cfg.FileStoragePath)

	// Загрузка данных из файла.
	if err := repo.LoadFromFile(); err != nil {
		log.Printf("Error loading storage from file: %v", err)
	}

	// Регистрация маршрутов.
	r := chi.NewRouter()
	r.Use(logger.Middleware)
	r.Use(compress.GzipMiddleware)
	r.Route("/", func(r chi.Router) {
		r.Post("/", handlers.PostHandler(repo, cfg.BaseURL))
		r.Get("/{id}", handlers.GetHandler(repo))
		r.Post("/api/shorten", handlers.PostJSONHandler(repo, cfg.BaseURL))
	})

	return r
}
