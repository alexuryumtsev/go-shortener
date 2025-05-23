package config

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strconv"

	"github.com/alexuryumtsev/go-shortener/internal/app/validator"
)

// Config содержит настройки конфигурации приложения.
type Config struct {
	ServerAddress   string `json:"server_address"`
	BaseURL         string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDSN     string `json:"database_dsn"`
	BatchSize       int    `json:"batch_size"`
	Debug           bool   `json:"debug"`
	EnableHTTPS     bool   `json:"enable_https"`
	CertPath        string `json:"cert_path"`
	KeyPath         string `json:"key_path"`
}

// Значения по умолчанию
const (
	defaultServerAddress = ":8080"
	defaultBaseURL       = "http://localhost:8080/"
	defaultStoragePath   = "/tmp/storage.json"
	defaultDatabaseDSN   = ""
	defaultBatchSize     = 10
	defaultDebug         = false
	defaultEnableHTTPS   = false
	defaultCertPath      = "./cert.pem"
	defaultKeyPath       = "./key.pem"
)

// loadConfigFile загружает конфигурацию из JSON файла
func loadConfigFile(filename string) (*Config, error) {
	if filename == "" {
		return &Config{}, nil
	}

	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(file, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// InitConfig инициализирует конфигурацию приложения
func InitConfig() (*Config, error) {
	var configPath string

	// Определяем путь к конфиг файлу
	flag.StringVar(&configPath, "c", "", "Path to config file")
	flag.StringVar(&configPath, "config", "", "Path to config file")

	// Если флаг не установлен, проверяем переменную окружения
	if configPath == "" {
		configPath = os.Getenv("CONFIG")
	}

	// Загружаем конфигурацию из файла
	fileCfg, err := loadConfigFile(configPath)
	if err != nil {
		return nil, err
	}

	// Создаем итоговую конфигурацию
	cfg := &Config{}

	// Получаем значения из переменных окружения
	envServerAddress := os.Getenv("SERVER_ADDRESS")
	envBaseURL := os.Getenv("BASE_URL")
	envPath := os.Getenv("FILE_STORAGE_PATH")
	envFileStorageName := os.Getenv("FILE_STORAGE_NAME")
	envDatabaseDSN := os.Getenv("DATABASE_DSN")
	envBatchSize := os.Getenv("BATCH_SIZE")
	envDebug := os.Getenv("DEBUG")
	envEnableHTTPS := os.Getenv("ENABLE_HTTPS")
	envCertPath := os.Getenv("CERT_PATH")
	envKeyPath := os.Getenv("KEY_PATH")

	// Устанавливаем значения из переменных окружения
	debug := defaultDebug
	if envDebug != "" {
		debug = envDebug == "true"
	}

	enableHTTPS := defaultEnableHTTPS
	if envEnableHTTPS != "" {
		enableHTTPS = envEnableHTTPS == "true"
	}

	// Определяем флаги
	flag.StringVar(&cfg.ServerAddress, "a", "", "HTTP server address")
	flag.StringVar(&cfg.BaseURL, "b", "", "Base URL for shortened links")
	flag.StringVar(&cfg.FileStoragePath, "f", "", "Path to file storage")
	flag.StringVar(&cfg.DatabaseDSN, "d", envDatabaseDSN, "Database connection string")
	flag.IntVar(&cfg.BatchSize, "batch", defaultBatchSize, "Batch size for bulk operations")
	flag.BoolVar(&cfg.Debug, "debug", debug, "Enable debug mode")
	flag.BoolVar(&cfg.EnableHTTPS, "s", enableHTTPS, "Enable HTTPS")
	flag.StringVar(&cfg.CertPath, "cert", "", "Path to SSL certificate")
	flag.StringVar(&cfg.KeyPath, "key", "", "Path to SSL key")

	// Обрабатываем флаги
	flag.Parse()

	// Приоритет: флаги > переменные окружения > файл конфигурации > значения по умолчанию

	// ServerAddress
	if cfg.ServerAddress == "" {
		if envServerAddress != "" {
			cfg.ServerAddress = envServerAddress
		} else if fileCfg.ServerAddress != "" {
			cfg.ServerAddress = fileCfg.ServerAddress
		} else {
			cfg.ServerAddress = defaultServerAddress
		}
	}

	// Проверка формата host:port
	if err := validator.ValidateServerAddress(cfg.ServerAddress); err != nil {
		return nil, err
	}

	// BaseURL
	if cfg.BaseURL == "" {
		if envBaseURL != "" {
			cfg.BaseURL = envBaseURL
		} else if fileCfg.BaseURL != "" {
			cfg.BaseURL = fileCfg.BaseURL
		} else {
			cfg.BaseURL = defaultBaseURL
		}
	}

	// FileStoragePath
	if cfg.FileStoragePath == "" {
		if envPath != "" {
			cfg.FileStoragePath = filepath.Join(envPath, envFileStorageName)
		} else if fileCfg.FileStoragePath != "" {
			cfg.FileStoragePath = fileCfg.FileStoragePath
		} else {
			cfg.FileStoragePath = defaultStoragePath
		}
	}

	// DatabaseDSN
	if cfg.DatabaseDSN == "" {
		if envDatabaseDSN != "" {
			cfg.DatabaseDSN = envDatabaseDSN
		} else if fileCfg.DatabaseDSN != "" {
			cfg.DatabaseDSN = fileCfg.DatabaseDSN
		} else {
			cfg.DatabaseDSN = defaultDatabaseDSN
		}
	}

	// BatchSize
	if cfg.BatchSize <= 0 {
		if envBatchSize != "" {
			if size, err := strconv.Atoi(envBatchSize); err == nil {
				cfg.BatchSize = size
			}
		} else if fileCfg.BatchSize > 0 {
			cfg.BatchSize = fileCfg.BatchSize
		} else {
			cfg.BatchSize = defaultBatchSize
		}
	}

	// SSL Certificates
	if cfg.CertPath == "" {
		if envCertPath != "" {
			cfg.CertPath = envCertPath
		} else if fileCfg.CertPath != "" {
			cfg.CertPath = fileCfg.CertPath
		} else {
			cfg.CertPath = defaultCertPath
		}
	}

	if cfg.KeyPath == "" {
		if envKeyPath != "" {
			cfg.KeyPath = envKeyPath
		} else if fileCfg.KeyPath != "" {
			cfg.KeyPath = fileCfg.KeyPath
		} else {
			cfg.KeyPath = defaultKeyPath
		}
	}

	// Проверка корректности URL
	if err := validator.ValidateBaseURL(cfg.BaseURL); err != nil {
		return nil, err
	}

	return cfg, nil
}
