package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitConfig(t *testing.T) {
	t.Run("Default Configuration", func(t *testing.T) {
		cfg, err := InitConfig()
		assert.NoError(t, err)
		assert.Equal(t, defaultServerAddress, cfg.ServerAddress)
		assert.Equal(t, defaultBaseURL, cfg.BaseURL)
		assert.Equal(t, defaultStoragePath, cfg.FileStoragePath)
	})
}
