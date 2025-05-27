package config

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strconv"

	"github.com/alexuryumtsev/go-shortener/internal/app/validator"
)

// Config содержит настройки конфигурации приложения
type Config struct {
	serverAddress   string
	baseURL         string
	fileStoragePath string
	databaseDSN     string
	batchSize       int
	debug           bool
	enableHTTPS     bool
	certPath        string
	keyPath         string
}

// Option определяет функциональную опцию для конфигурации
type Option func(*Config)

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

// New создает новую конфигурацию с применением опций
func New(opts ...Option) (*Config, error) {
	cfg := &Config{
		serverAddress:   defaultServerAddress,
		baseURL:         defaultBaseURL,
		fileStoragePath: defaultStoragePath,
		databaseDSN:     defaultDatabaseDSN,
		batchSize:       defaultBatchSize,
		debug:           defaultDebug,
		enableHTTPS:     defaultEnableHTTPS,
		certPath:        defaultCertPath,
		keyPath:         defaultKeyPath,
	}

	// Применяем все опции
	for _, opt := range opts {
		opt(cfg)
	}

	// Валидация конфигурации
	if err := validator.ValidateServerAddress(cfg.serverAddress); err != nil {
		return nil, err
	}
	if err := validator.ValidateBaseURL(cfg.baseURL); err != nil {
		return nil, err
	}

	return cfg, nil
}

// WithServerAddress устанавливает адрес сервера
func WithServerAddress(addr string) Option {
	return func(c *Config) {
		if addr != "" {
			c.serverAddress = addr
		}
	}
}

// WithBaseURL устанавливает базовый URL
func WithBaseURL(url string) Option {
	return func(c *Config) {
		if url != "" {
			c.baseURL = url
		}
	}
}

// WithFileStoragePath устанавливает путь к файлу хранилища
func WithFileStoragePath(path string) Option {
	return func(c *Config) {
		if path != "" {
			c.fileStoragePath = path
		}
	}
}

// WithDatabaseDSN устанавливает строку подключения к БД
func WithDatabaseDSN(dsn string) Option {
	return func(c *Config) {
		if dsn != "" {
			c.databaseDSN = dsn
		}
	}
}

// WithBatchSize устанавливает размер пакета
func WithBatchSize(size int) Option {
	return func(c *Config) {
		if size > 0 {
			c.batchSize = size
		}
	}
}

// WithDebug устанавливает режим отладки
func WithDebug(debug bool) Option {
	return func(c *Config) {
		c.debug = debug
	}
}

// WithHTTPS устанавливает использование HTTPS
func WithHTTPS(enable bool, certPath, keyPath string) Option {
	return func(c *Config) {
		c.enableHTTPS = enable
		if certPath != "" {
			c.certPath = certPath
		}
		if keyPath != "" {
			c.keyPath = keyPath
		}
	}
}

// FromFile загружает конфигурацию из JSON файла
func FromFile(filename string) Option {
	return func(c *Config) {
		if filename == "" {
			return
		}

		file, err := os.ReadFile(filename)
		if err != nil {
			return
		}

		var fileCfg struct {
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

		if err := json.Unmarshal(file, &fileCfg); err != nil {
			return
		}

		WithServerAddress(fileCfg.ServerAddress)(c)
		WithBaseURL(fileCfg.BaseURL)(c)
		WithFileStoragePath(fileCfg.FileStoragePath)(c)
		WithDatabaseDSN(fileCfg.DatabaseDSN)(c)
		WithBatchSize(fileCfg.BatchSize)(c)
		WithDebug(fileCfg.Debug)(c)
		WithHTTPS(fileCfg.EnableHTTPS, fileCfg.CertPath, fileCfg.KeyPath)(c)
	}
}

// FromEnv загружает конфигурацию из переменных окружения
func FromEnv() Option {
	return func(c *Config) {
		// Сетевые настройки
		if addr := os.Getenv("SERVER_ADDRESS"); addr != "" {
			WithServerAddress(addr)(c)
		}
		if url := os.Getenv("BASE_URL"); url != "" {
			WithBaseURL(url)(c)
		}

		// Настройки хранилища
		if path := os.Getenv("FILE_STORAGE_PATH"); path != "" {
			name := os.Getenv("FILE_STORAGE_NAME")
			fullPath := filepath.Join(path, name)
			WithFileStoragePath(fullPath)(c)
		}
		if dsn := os.Getenv("DATABASE_DSN"); dsn != "" {
			WithDatabaseDSN(dsn)(c)
		}

		// Настройки производительности
		if size := os.Getenv("BATCH_SIZE"); size != "" {
			if s, err := strconv.Atoi(size); err == nil {
				WithBatchSize(s)(c)
			}
		}

		// Настройки отладки
		if debug := os.Getenv("DEBUG"); debug == "true" {
			WithDebug(true)(c)
		}

		// Настройки безопасности (HTTPS)
		WithHTTPS(
			os.Getenv("ENABLE_HTTPS") == "true",
			os.Getenv("CERT_PATH"),
			os.Getenv("KEY_PATH"),
		)(c)
	}
}

// FromFlags загружает конфигурацию из флагов командной строки
func FromFlags() Option {
	return func(c *Config) {
		var (
			// Сетевые настройки
			serverAddress string
			baseURL       string

			// Настройки хранилища
			fileStoragePath string
			databaseDSN     string

			// Настройки производительности
			batchSize int

			// Настройки отладки
			debug bool

			// Настройки безопасности (HTTPS)
			enableHTTPS bool
			certPath    string
			keyPath     string
		)

		// Определение флагов
		flag.StringVar(&serverAddress, "a", "", "HTTP server address")
		flag.StringVar(&baseURL, "b", "", "Base URL for shortened links")
		flag.StringVar(&fileStoragePath, "f", "", "Path to file storage")
		flag.StringVar(&databaseDSN, "d", "", "Database connection string")
		flag.IntVar(&batchSize, "batch", 0, "Batch size for bulk operations")
		flag.BoolVar(&debug, "debug", false, "Enable debug mode")
		flag.BoolVar(&enableHTTPS, "s", false, "Enable HTTPS")
		flag.StringVar(&certPath, "cert", "", "Path to SSL certificate")
		flag.StringVar(&keyPath, "key", "", "Path to SSL key")

		flag.Parse()

		// Применение значений
		WithServerAddress(serverAddress)(c)
		WithBaseURL(baseURL)(c)
		WithFileStoragePath(fileStoragePath)(c)
		WithDatabaseDSN(databaseDSN)(c)
		WithBatchSize(batchSize)(c)
		WithDebug(debug)(c)
		WithHTTPS(enableHTTPS, certPath, keyPath)(c)
	}
}

// InitConfig инициализирует конфигурацию приложения
func InitConfig() (*Config, error) {
	var configPath string
	flag.StringVar(&configPath, "c", "", "Path to config file")
	flag.StringVar(&configPath, "config", "", "Path to config file")

	if configPath == "" {
		configPath = os.Getenv("CONFIG")
	}

	// Создаем конфигурацию с приоритетом: флаги > переменные окружения > файл > значения по умолчанию
	return New(
		FromFile(configPath), // Наименьший приоритет
		FromEnv(),            // Средний приоритет
		FromFlags(),          // Наивысший приоритет
	)
}

// Геттеры для доступа к полям конфигурации
func (c *Config) ServerAddress() string   { return c.serverAddress }
func (c *Config) BaseURL() string         { return c.baseURL }
func (c *Config) FileStoragePath() string { return c.fileStoragePath }
func (c *Config) DatabaseDSN() string     { return c.databaseDSN }
func (c *Config) BatchSize() int          { return c.batchSize }
func (c *Config) Debug() bool             { return c.debug }
func (c *Config) EnableHTTPS() bool       { return c.enableHTTPS }
func (c *Config) CertPath() string        { return c.certPath }
func (c *Config) KeyPath() string         { return c.keyPath }
