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
func ShortenerRouter(cfg *config.Config, repo storage.URLStorage) chi.Router {
	// Загрузка данных из файла, если используется файловое хранилище.
	if fileRepo, ok := repo.(*storage.FileStorage); ok {
		if err := fileRepo.LoadFromFile(); err != nil {
			log.Printf("Error loading storage from file: %v", err)
		}
	}

	// Регистрация маршрутов.
	r := chi.NewRouter()
	r.Use(logger.Middleware)
	r.Use(compress.GzipMiddleware)
	r.Route("/", func(r chi.Router) {
		r.Post("/", handlers.PostHandler(repo, cfg.BaseURL))
		r.Get("/{id}", handlers.GetHandler(repo))
		r.Get("/ping", handlers.PingHandler(repo))
		r.Post("/api/shorten", handlers.PostJSONHandler(repo, cfg.BaseURL))
	})

	return r
}
