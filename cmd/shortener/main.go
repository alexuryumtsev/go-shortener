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

// Информация о сборке приложения.
// Заполняется при компиляции с помощью флагов линковщика:
// -ldflags "-X main.buildVer=v1.0.0 -X main.buildDt=2023-05-11 -X main.buildCmt=abc123"
var (
	buildVer string // версия сборки
	buildDt  string // дата сборки
	buildCmt string // коммит, на котором собрана версия
)

// Функция для вывода информации о сборке
func printBuildInfo() {
	version := buildVer
	if version == "" {
		version = "N/A"
	}

	date := buildDt
	if date == "" {
		date = "N/A"
	}

	commit := buildCmt
	if commit == "" {
		commit = "N/A"
	}

	fmt.Printf("Build version: %s\n", version)
	fmt.Printf("Build date: %s\n", date)
	fmt.Printf("Build commit: %s\n", commit)
}

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
		repo = pg.NewDatabaseStorage(pool)
	} else if cfg.FileStoragePath != "" {
		repo = file.NewFileStorage(cfg.FileStoragePath)
	} else {
		repo = memory.NewInMemoryStorage()
	}

	// Инициализируем сервисы
	userService := user.NewUserService("super-secret-key")
	urlService := url.NewURLService(repo, cfg.BaseURL, cfg.BatchSize)

	// Запуск сервера
	fmt.Println("Server started at", cfg.ServerAddress)
	err = http.ListenAndServe(cfg.ServerAddress, router.ShortenerRouter(cfg, repo, userService, urlService))
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
