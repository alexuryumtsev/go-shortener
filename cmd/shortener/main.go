package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/alexuryumtsev/go-shortener/config"
	"github.com/alexuryumtsev/go-shortener/internal/app/db"
	"github.com/alexuryumtsev/go-shortener/internal/app/logger"
	"github.com/alexuryumtsev/go-shortener/internal/app/router"
	"github.com/alexuryumtsev/go-shortener/internal/app/storage"
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
		pool, err := db.NewDatabaseConnection(ctx, cfg.DatabaseDSN)
		if err != nil {
			log.Fatalf("Failed connect to db: %v", err)
		}
		defer pool.Close()
		repo = storage.NewDatabaseStorage(pool)
	} else if cfg.FileStoragePath != "" {
		repo = storage.NewFileStorage(cfg.FileStoragePath)
	} else {
		repo = storage.NewInMemoryStorage()
	}

	// Запуск сервера
	fmt.Println("Server started at", cfg.ServerAddress)
	err = http.ListenAndServe(cfg.ServerAddress, router.ShortenerRouter(cfg, repo))
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
