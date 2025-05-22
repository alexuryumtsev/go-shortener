// Package config содержит функции и структуры для работы с конфигурацией приложения.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
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

	// EnableHTTPS включает HTTPS режим
	// По умолчанию: false
	EnableHTTPS bool

	// ConfigPath указывает путь к файлу конфигурации
	ConfigPath string
}

// JSONConfig представляет структуру JSON файла конфигурации
type JSONConfig struct {
	ServerAddress   string `json:"server_address"`
	BaseURL         string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDSN     string `json:"database_dsn"`
	BatchSize       int    `json:"batch_size"`
	Debug           bool   `json:"debug"`
	EnableHTTPS     bool   `json:"enable_https"`
}

// Значения по умолчанию.
const (
	defaultServerAddress = ":8080"
	defaultBaseURL       = "http://localhost:8080/"
	defaultStoragePath   = "/tmp/storage.json"
	defaultDatabaseDSN   = ""
	defaultBatchSize     = 10
	defaultDebug         = false
	defaultEnableHTTPS   = false
)

// loadJSONConfig загружает конфигурацию из JSON файла
func loadJSONConfig(path string) (*JSONConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var jsonCfg JSONConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&jsonCfg); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	return &jsonCfg, nil
}

// InitConfig инициализирует конфигурацию приложения.
// Читает параметры из файла конфигурации, переменных окружения и флагов командной строки.
// Приоритет (от низкого к высокому): файл конфигурации < переменные окружения < флаги.
// Возвращает указатель на Config и ошибку в случае некорректных параметров.
func InitConfig() (*Config, error) {
	cfg := &Config{}

	// Определяем флаг для пути к конфигурационному файлу
	var configPath string
	flag.StringVar(&configPath, "c", "", "Path to configuration file")
	flag.StringVar(&configPath, "config", "", "Path to configuration file")

	// Получаем путь к конфигурационному файлу из переменной окружения
	envConfigPath := os.Getenv("CONFIG")

	// Получаем значения из переменных окружения.
	envServerAddress := os.Getenv("SERVER_ADDRESS")
	envBaseURL := os.Getenv("BASE_URL")
	envPath := os.Getenv("FILE_STORAGE_PATH")
	envFileStorageName := os.Getenv("FILE_STORAGE_NAME")
	envDatabaseDSN := os.Getenv("DATABASE_DSN")
	envBatchSize := os.Getenv("BATCH_SIZE")
	envDebug := os.Getenv("DEBUG")
	envEnableHTTPS := os.Getenv("ENABLE_HTTPS")

	debug := defaultDebug
	if envDebug != "" {
		debug = envDebug == "true"
	}

	enableHTTPS := defaultEnableHTTPS
	if envEnableHTTPS != "" {
		enableHTTPS = envEnableHTTPS == "true"
	}

	// Определяем флаги
	flag.StringVar(&cfg.ServerAddress, "a", "", "HTTP server address, host:port")
	flag.StringVar(&cfg.BaseURL, "b", "", "Base URL for shortened links")
	flag.StringVar(&cfg.FileStoragePath, "f", "", "Path to file storage")
	flag.StringVar(&cfg.DatabaseDSN, "d", envDatabaseDSN, "Строка подключения к базе данных (DSN)")
	flag.IntVar(&cfg.BatchSize, "batch", defaultBatchSize, "Batch size for bulk operations")
	flag.BoolVar(&cfg.Debug, "debug", debug, "Enable debug mode")
	flag.BoolVar(&cfg.EnableHTTPS, "s", enableHTTPS, "Enable HTTPS")

	// Обрабатываем флаги
	flag.Parse()

	// Инициализируем значениями по умолчанию
	cfg.ServerAddress = defaultServerAddress
	cfg.BaseURL = defaultBaseURL
	cfg.FileStoragePath = defaultStoragePath
	cfg.DatabaseDSN = defaultDatabaseDSN
	cfg.BatchSize = defaultBatchSize
	cfg.Debug = defaultDebug
	cfg.EnableHTTPS = defaultEnableHTTPS

	// Определяем путь к конфигурационному файлу
	if configPath == "" && envConfigPath != "" {
		configPath = envConfigPath
	}
	cfg.ConfigPath = configPath

	// Загружаем конфигурацию из JSON файла (если указан)
	if configPath != "" {
		jsonCfg, err := loadJSONConfig(configPath)
		if err != nil {
			// Если файл указан явно, но не может быть загружен - это ошибка
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}

		// Применяем значения из JSON файла
		if jsonCfg.ServerAddress != "" {
			cfg.ServerAddress = jsonCfg.ServerAddress
		}
		if jsonCfg.BaseURL != "" {
			cfg.BaseURL = jsonCfg.BaseURL
		}
		if jsonCfg.FileStoragePath != "" {
			cfg.FileStoragePath = jsonCfg.FileStoragePath
		}
		if jsonCfg.DatabaseDSN != "" {
			cfg.DatabaseDSN = jsonCfg.DatabaseDSN
		}
		if jsonCfg.BatchSize > 0 {
			cfg.BatchSize = jsonCfg.BatchSize
		}
		// Для булевых значений проверяем явное указание в JSON
		cfg.Debug = jsonCfg.Debug
		cfg.EnableHTTPS = jsonCfg.EnableHTTPS
	}

	// Применяем переменные окружения (более высокий приоритет)
	if envServerAddress != "" {
		cfg.ServerAddress = envServerAddress
	}

	if envBaseURL != "" {
		cfg.BaseURL = envBaseURL
	}

	// Обработка файлового хранилища из переменных окружения
	if envPath != "" || envFileStorageName != "" {
		if envPath != "" && envFileStorageName != "" {
			cfg.FileStoragePath = filepath.Join(envPath, envFileStorageName)
		} else if envPath != "" {
			cfg.FileStoragePath = filepath.Join(envPath, "storage.json")
		} else if envFileStorageName != "" {
			cfg.FileStoragePath = envFileStorageName
		}
	}

	if envDatabaseDSN != "" {
		cfg.DatabaseDSN = envDatabaseDSN
	}

	if envBatchSize != "" {
		size, parseErr := strconv.Atoi(envBatchSize)
		if parseErr == nil && size > 0 {
			cfg.BatchSize = size
		}
	}

	if envDebug != "" {
		cfg.Debug = envDebug == "true"
	}

	if envEnableHTTPS != "" {
		cfg.EnableHTTPS = envEnableHTTPS == "true"
	}

	// Применяем флаги командной строки (наивысший приоритет)
	// flag.Parse() уже был вызван, проверяем были ли флаги установлены
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "a":
			cfg.ServerAddress = f.Value.String()
		case "b":
			cfg.BaseURL = f.Value.String()
		case "f":
			path := f.Value.String()
			if path != "" {
				cfg.FileStoragePath = filepath.Join(path, "storage.json")
			}
		case "d":
			cfg.DatabaseDSN = f.Value.String()
		case "batch":
			if size, err := strconv.Atoi(f.Value.String()); err == nil && size > 0 {
				cfg.BatchSize = size
			}
		case "debug":
			cfg.Debug = f.Value.String() == "true"
		case "s":
			cfg.EnableHTTPS = f.Value.String() == "true"
		}
	})

	// Валидация значений
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = defaultBatchSize
	}

	// Проверка формата host:port
	err := validator.ValidateServerAddress(cfg.ServerAddress)
	if err != nil {
		return nil, err
	}

	// Проверка корректности URL
	err = validator.ValidateBaseURL(cfg.BaseURL)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
