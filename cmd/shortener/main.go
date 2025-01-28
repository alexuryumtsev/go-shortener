package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/alexuryumtsev/go-shortener/config"
	"github.com/alexuryumtsev/go-shortener/internal/app/logger"
	"github.com/alexuryumtsev/go-shortener/internal/app/router"
)

func main() {
	// Инициализируем конфигурацию
	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	// Инициализируем логгер
	logger.InitLogger()

	// Запуск сервера
	fmt.Println("Server started at", cfg.ServerAddress)
	err = http.ListenAndServe(cfg.ServerAddress, router.ShortenerRouter(cfg))
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
