package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/alexuryumtsev/go-shortener/config"
	"github.com/alexuryumtsev/go-shortener/internal/app/db"
	"github.com/alexuryumtsev/go-shortener/internal/app/logger"
	"github.com/alexuryumtsev/go-shortener/internal/app/router"
	"github.com/alexuryumtsev/go-shortener/internal/app/service/url"
	"github.com/alexuryumtsev/go-shortener/internal/app/service/user"
	"github.com/alexuryumtsev/go-shortener/internal/app/storage"
	"github.com/alexuryumtsev/go-shortener/internal/app/storage/file"
	"github.com/alexuryumtsev/go-shortener/internal/app/storage/memory"
	"github.com/alexuryumtsev/go-shortener/internal/app/storage/pg"
)

func main() {
	// Инициализируем конфигурацию
	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	// Инициализируем логгер
	logger.InitLogger()

	// Подключаемся к базе данных
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var repo storage.URLStorage
	if cfg.DatabaseDSN != "" {
		pool, errConnect := db.NewDatabaseConnection(ctx, cfg.DatabaseDSN)
		if errConnect != nil {
			log.Fatalf("Failed connect to db: %v", errConnect)
		}
		defer pool.Close()
		repo = pg.NewDatabaseStorage(pool)
	} else if cfg.FileStoragePath != "" {
		repo = file.NewFileStorage(cfg.FileStoragePath)
	} else {
		repo = memory.NewInMemoryStorage()
	}

	// Инициализируем сервис пользователя
	userService := user.NewUserService("super-secret-key")
	urlService := url.NewURLService(ctx, repo, cfg.BaseURL, cfg.BatchSize)

	// Запуск сервера
	fmt.Println("Server started at", cfg.ServerAddress)
	err = http.ListenAndServe(cfg.ServerAddress, router.ShortenerRouter(cfg, repo, userService, urlService))
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
