package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfigFile(t *testing.T) {
	content := `{
        "server_address": "localhost:9000",
        "base_url": "http://test.com",
        "file_storage_path": "/test/path.db",
        "database_dsn": "test-dsn",
        "enable_https": false,
        "trusted_subnet": "192.168.1.0/24"
    }`

	tmpfile, err := os.CreateTemp("", "config-*.json")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(content))
	assert.NoError(t, err)
	tmpfile.Close()

	os.Setenv("CONFIG", tmpfile.Name())
	defer os.Unsetenv("CONFIG")

	cfg, err := InitConfig()
	assert.NoError(t, err)
	assert.Equal(t, "localhost:9000", cfg.ServerAddress())
	assert.Equal(t, "http://test.com", cfg.BaseURL())
	assert.Equal(t, "/test/path.db", cfg.FileStoragePath())
	assert.Equal(t, "test-dsn", cfg.DatabaseDSN())
	assert.False(t, cfg.EnableHTTPS())
	assert.Equal(t, "192.168.1.0/24", cfg.TrustedSubnet())
}
