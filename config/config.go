package config

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/alexuryumtsev/go-shortener/internal/app/validator"
)

type Config struct {
	ServerAddress   string // Адрес запуска HTTP-сервера
	BaseURL         string // Базовый адрес для сокращённых URL
	FileStoragePath string // Путь к файлу хранилища
	DatabaseDSN     string // подключения к PostgreSQL
}

// Значения по умолчанию.
const (
	defaultServerAddress = ":8080"
	defaultBaseURL       = "http://localhost:8080/"
	defaultStoragePath   = "tmp/storage.json"
	defaultDatabaseDSN   = ""
)

func InitConfig() (*Config, error) {
	cfg := &Config{}

	// Получаем значения из переменных окружения.
	envServerAddress := os.Getenv("SERVER_ADDRESS")
	envBaseURL := os.Getenv("BASE_URL")
	envPath := os.Getenv("FILE_STORAGE_PATH")
	envFileStorageName := os.Getenv("FILE_STORAGE_NAME")
	envDatabaseDSN := os.Getenv("DATABASE_DSN")

	// Определяем флаги
	flag.StringVar(&cfg.ServerAddress, "a", "", "HTTP server address, host:port")
	flag.StringVar(&cfg.BaseURL, "b", "", "Base URL for shortened links")
	flag.StringVar(&cfg.FileStoragePath, "f", "", "Path to file storage")
	flag.StringVar(&cfg.DatabaseDSN, "d", envDatabaseDSN, "Строка подключения к базе данных (DSN)")

	// Обрабатываем флаги
	flag.Parse()

	// Проверяем значения флагов и переменных окружения
	if cfg.ServerAddress == "" {
		cfg.ServerAddress = envServerAddress
	}

	if cfg.ServerAddress == "" {
		cfg.ServerAddress = defaultServerAddress
	}

	// Проверка формата host:port
	err := validator.ValidateServerAddress(cfg.ServerAddress)
	if err != nil {
		return nil, err
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = envBaseURL
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}

	if cfg.FileStoragePath != "" {
		cfg.FileStoragePath = filepath.Join(cfg.FileStoragePath, "storage.json")
	}

	if cfg.FileStoragePath == "" {
		cfg.FileStoragePath = filepath.Join(envPath, envFileStorageName)
	}

	if cfg.FileStoragePath == "" {
		cfg.FileStoragePath = defaultStoragePath
	}

	if cfg.DatabaseDSN == "" {
		cfg.DatabaseDSN = defaultDatabaseDSN
	}

	// Проверка корректности URL
	err = validator.ValidateBaseURL(cfg.BaseURL)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
