package config

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitConfig(t *testing.T) {
	// Сброс флагов перед каждым тестом
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	t.Run("Default Configuration", func(t *testing.T) {
		cfg, err := InitConfig()
		assert.NoError(t, err)
		assert.Equal(t, defaultServerAddress, cfg.ServerAddress)
		assert.Equal(t, defaultBaseURL, cfg.BaseURL)
		assert.Equal(t, defaultStoragePath, cfg.FileStoragePath)
		assert.Equal(t, defaultDatabaseDSN, cfg.DatabaseDSN)
		assert.Equal(t, defaultBatchSize, cfg.BatchSize)
		assert.Equal(t, defaultDebug, cfg.Debug)
		assert.Equal(t, defaultEnableHTTPS, cfg.EnableHTTPS)
	})

	t.Run("Config from JSON file", func(t *testing.T) {
		// Создаем временный конфигурационный файл
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.json")

		configData := JSONConfig{
			ServerAddress:   "localhost:9090",
			BaseURL:         "https://short.url",
			FileStoragePath: "/custom/path/storage.db",
			DatabaseDSN:     "postgres://user:pass@localhost/db",
			BatchSize:       100,
			Debug:           true,
			EnableHTTPS:     true,
		}

		data, err := json.Marshal(configData)
		require.NoError(t, err)

		err = os.WriteFile(configPath, data, 0644)
		require.NoError(t, err)

		// Устанавливаем переменную окружения
		os.Setenv("CONFIG", configPath)
		defer os.Unsetenv("CONFIG")

		// Сброс флагов
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

		cfg, err := InitConfig()
		assert.NoError(t, err)
		assert.Equal(t, "localhost:9090", cfg.ServerAddress)
		assert.Equal(t, "https://short.url", cfg.BaseURL)
		assert.Equal(t, "/custom/path/storage.db", cfg.FileStoragePath)
		assert.Equal(t, "postgres://user:pass@localhost/db", cfg.DatabaseDSN)
		assert.Equal(t, 100, cfg.BatchSize)
		assert.True(t, cfg.Debug)
		assert.True(t, cfg.EnableHTTPS)
	})

	t.Run("Environment variables override JSON config", func(t *testing.T) {
		// Создаем временный конфигурационный файл
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.json")

		configData := JSONConfig{
			ServerAddress: "localhost:9090",
			BaseURL:       "https://short.url",
		}

		data, err := json.Marshal(configData)
		require.NoError(t, err)

		err = os.WriteFile(configPath, data, 0644)
		require.NoError(t, err)

		// Устанавливаем переменные окружения
		os.Setenv("CONFIG", configPath)
		os.Setenv("SERVER_ADDRESS", ":7070")
		os.Setenv("BASE_URL", "http://env.url")
		defer func() {
			os.Unsetenv("CONFIG")
			os.Unsetenv("SERVER_ADDRESS")
			os.Unsetenv("BASE_URL")
		}()

		// Сброс флагов
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

		cfg, err := InitConfig()
		assert.NoError(t, err)
		assert.Equal(t, ":7070", cfg.ServerAddress)
		assert.Equal(t, "http://env.url", cfg.BaseURL)
	})

	t.Run("Flags override environment and JSON config", func(t *testing.T) {
		// Создаем временный конфигурационный файл
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.json")

		configData := JSONConfig{
			ServerAddress: "localhost:9090",
			BaseURL:       "https://short.url",
		}

		data, err := json.Marshal(configData)
		require.NoError(t, err)

		err = os.WriteFile(configPath, data, 0644)
		require.NoError(t, err)

		// Устанавливаем переменные окружения
		os.Setenv("SERVER_ADDRESS", ":7070")
		os.Setenv("BASE_URL", "http://env.url")
		defer func() {
			os.Unsetenv("SERVER_ADDRESS")
			os.Unsetenv("BASE_URL")
		}()

		// Сброс флагов и установка новых
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		os.Args = []string{"cmd", "-config", configPath, "-a", ":5050", "-b", "http://flag.url"}

		cfg, err := InitConfig()
		assert.NoError(t, err)
		assert.Equal(t, ":7070", cfg.ServerAddress)
		assert.Equal(t, "http://env.url", cfg.BaseURL)
	})

	t.Run("Invalid JSON config file", func(t *testing.T) {
		// Создаем временный файл с невалидным JSON
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "invalid.json")

		err := os.WriteFile(configPath, []byte("invalid json"), 0644)
		require.NoError(t, err)

		// Устанавливаем переменную окружения
		os.Setenv("CONFIG", configPath)
		defer os.Unsetenv("CONFIG")

		// Сброс флагов
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

		_, err = InitConfig()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load config file")
	})

	t.Run("Non-existent config file", func(t *testing.T) {
		// Устанавливаем путь к несуществующему файлу
		os.Setenv("CONFIG", "/non/existent/config.json")
		defer os.Unsetenv("CONFIG")

		// Сброс флагов
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

		_, err := InitConfig()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load config file")
	})

	t.Run("Config file path from -c flag", func(t *testing.T) {
		// Создаем временный конфигурационный файл
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.json")

		configData := JSONConfig{
			ServerAddress: "localhost:8888",
		}

		data, err := json.Marshal(configData)
		require.NoError(t, err)

		err = os.WriteFile(configPath, data, 0644)
		require.NoError(t, err)

		// Сброс флагов и установка новых
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		os.Args = []string{"cmd", "-c", configPath}

		cfg, err := InitConfig()
		assert.NoError(t, err)
		assert.Equal(t, "localhost:8888", cfg.ServerAddress)
		assert.Equal(t, configPath, cfg.ConfigPath)
	})
}
