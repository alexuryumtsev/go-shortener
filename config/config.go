// Package config содержит функции и структуры для работы с конфигурацией приложения.
package config

import (
	"flag"
	"os"
	"path/filepath"
	"strconv"

	"github.com/alexuryumtsev/go-shortener/internal/app/validator"
)

// Config содержит настройки конфигурации приложения.
// Включает параметры сервера, базы данных и другие настройки.
type Config struct {
	// ServerAddress определяет адрес запуска HTTP-сервера
	// По умолчанию: ":8080"
	ServerAddress string

	// BaseURL определяет базовый адрес для сокращённых URL
	// По умолчанию: "http://localhost:8080/"
	BaseURL string

	// FileStoragePath указывает путь к файлу хранилища
	// По умолчанию: "/tmp/storage.json"
	FileStoragePath string

	// DatabaseDSN определяет строку подключения к PostgreSQL
	// По умолчанию: "" (пустая строка)
	DatabaseDSN string

	// BatchSize определяет размер батча для пакетных операций
	// По умолчанию: 10
	BatchSize int

	// Debug включает режим отладки
	// По умолчанию: false
	Debug bool
}

// Значения по умолчанию.
const (
	defaultServerAddress = ":8080"
	defaultBaseURL       = "http://localhost:8080/"
	defaultStoragePath   = "/tmp/storage.json"
	defaultDatabaseDSN   = ""
	defaultBatchSize     = 10
	defaultDebug         = false
)

// InitConfig инициализирует конфигурацию приложения.
// Читает параметры из переменных окружения и флагов командной строки.
// Возвращает указатель на Config и ошибку в случае некорректных параметров.
func InitConfig() (*Config, error) {
	cfg := &Config{}

	// Получаем значения из переменных окружения.
	envServerAddress := os.Getenv("SERVER_ADDRESS")
	envBaseURL := os.Getenv("BASE_URL")
	envPath := os.Getenv("FILE_STORAGE_PATH")
	envFileStorageName := os.Getenv("FILE_STORAGE_NAME")
	envDatabaseDSN := os.Getenv("DATABASE_DSN")
	envBatchSize := os.Getenv("BATCH_SIZE")
	envDebug := os.Getenv("DEBUG")

	debug := defaultDebug
	if envDebug != "" {
		debug = envDebug == "true"
	}

	// Определяем флаги
	flag.StringVar(&cfg.ServerAddress, "a", "", "HTTP server address, host:port")
	flag.StringVar(&cfg.BaseURL, "b", "", "Base URL for shortened links")
	flag.StringVar(&cfg.FileStoragePath, "f", "", "Path to file storage")
	flag.StringVar(&cfg.DatabaseDSN, "d", envDatabaseDSN, "Строка подключения к базе данных (DSN)")
	flag.IntVar(&cfg.BatchSize, "batch", defaultBatchSize, "Batch size for bulk operations")
	flag.BoolVar(&cfg.Debug, "debug", debug, "Enable debug mode")

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

	// Установка размера батча из переменной окружения, если указана
	if envBatchSize != "" {
		size, parseErr := strconv.Atoi(envBatchSize) // используем другое имя переменной
		if parseErr == nil {
			cfg.BatchSize = size
		}
	}

	if cfg.BatchSize <= 0 {
		cfg.BatchSize = defaultBatchSize
	}

	// Проверка корректности URL
	err = validator.ValidateBaseURL(cfg.BaseURL)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
