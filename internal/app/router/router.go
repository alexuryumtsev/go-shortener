// Package router содержит маршруты и обработчики для приложения.
// Он использует библиотеку chi для создания маршрутов и middleware для обработки запросов.
package router

import (
	"log"
	"net/http"
	"net/http/pprof"

	"github.com/alexuryumtsev/go-shortener/config"
	"github.com/alexuryumtsev/go-shortener/internal/app/compress"
	"github.com/alexuryumtsev/go-shortener/internal/app/handlers"
	"github.com/alexuryumtsev/go-shortener/internal/app/logger"
	"github.com/alexuryumtsev/go-shortener/internal/app/middleware"
	"github.com/alexuryumtsev/go-shortener/internal/app/service/url"
	"github.com/alexuryumtsev/go-shortener/internal/app/service/user"
	"github.com/alexuryumtsev/go-shortener/internal/app/storage"
	"github.com/alexuryumtsev/go-shortener/internal/app/storage/file"
	"github.com/go-chi/chi/v5"
)

// setupProfiling добавляет маршруты для профилирования
func setupProfiling(r chi.Router) {
	r.HandleFunc("/debug/pprof/*", pprof.Index)
	r.Get("/debug/pprof/profile", pprof.Profile)
}

// ShortenerRouter создает маршруты для приложения.
func ShortenerRouter(cfg *config.Config, repo storage.URLStorage, userService user.UserService, urlService url.URLService) chi.Router {
	// Загрузка данных из файла, если используется файловое хранилище.
	if fileRepo, ok := repo.(*file.FileStorage); ok {
		if err := fileRepo.LoadFromFile(); err != nil {
			log.Printf("Error loading storage from file: %v", err)
		}
	}

	// Регистрация маршрутов.
	r := chi.NewRouter()

	r.Use(logger.Middleware)
	r.Use(compress.GzipMiddleware)
	r.Use(middleware.ErrorMiddleware)

	// Добавляем middleware с userService
	r.Use(func(next http.Handler) http.Handler {
		return middleware.AuthMiddleware(userService, next)
	})

	r.Route("/", func(r chi.Router) {
		// Добавляем профилирование только в режиме отладки
		log.Println(cfg.Debug)
		if cfg.Debug {
			setupProfiling(r)
		}

		r.Post("/", handlers.PostHandler(repo, userService, cfg))
		r.Get("/{id}", handlers.GetHandler(repo))
		r.Get("/ping", handlers.PingHandler(repo))
		r.Get("/api/user/urls", handlers.GetUserURLsHandler(repo, userService, cfg))
		r.Delete("/api/user/urls", handlers.DeleteUserURLsHandler(urlService, userService))
		r.Post("/api/shorten", handlers.PostJSONHandler(repo, userService, cfg))
		r.Post("/api/shorten/batch", handlers.PostBatchHandler(repo, userService, cfg))
	})

	return r
}
