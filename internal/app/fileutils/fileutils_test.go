package fileutils

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexuryumtsev/go-shortener/internal/app/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnsureDirExists(t *testing.T) {
	tempDir := os.TempDir()
	testPath := filepath.Join(tempDir, "test_dir", "test.txt")

	err := EnsureDirExists(testPath)
	assert.NoError(t, err)

	// Проверяем, что директория создана
	_, err = os.Stat(filepath.Dir(testPath))
	assert.NoError(t, err)

	// Очистка
	os.RemoveAll(filepath.Dir(testPath))
}

func TestFileStorage_SaveAndLoadRecords(t *testing.T) {
	// Тестовые данные
	urlModel := models.URLModel{
		ID:      "abc123",
		URL:     "https://example.com",
		UserID:  "user123",
		Deleted: false,
	}

	t.Run("Save and Load Record", func(t *testing.T) {
		var buf bytes.Buffer
		fs := NewFileStorage("test.json")

		// Сохраняем запись
		err := fs.SaveRecord(&nopWriteCloser{&buf}, urlModel)
		require.NoError(t, err)

		// Загружаем записи
		data, err := fs.LoadRecords(&buf)
		require.NoError(t, err)

		// Проверяем результат
		loaded, exists := data[urlModel.ID]
		assert.True(t, exists)
		assert.Equal(t, urlModel, loaded)
	})
}

// Вспомогательный тип для тестирования
type nopWriteCloser struct {
	*bytes.Buffer
}

func (nopWriteCloser) Close() error { return nil }

// Бенчмарки
func BenchmarkFileStorage(b *testing.B) {
	urlModels := []models.URLModel{
		{
			ID:      "abc123",
			URL:     "https://example.com",
			UserID:  "user123",
			Deleted: false,
		},
		{
			ID:      "def456",
			URL:     "https://another.com",
			UserID:  "user456",
			Deleted: true,
		},
	}

	b.Run("SaveRecord", func(b *testing.B) {
		fs := NewFileStorage("bench.json")
		var buf bytes.Buffer
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			buf.Reset()
			err := fs.SaveRecord(&nopWriteCloser{&buf}, urlModels[i%2])
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("LoadRecords", func(b *testing.B) {
		fs := NewFileStorage("bench.json")
		var buf bytes.Buffer

		// Подготовка данных для загрузки
		for _, model := range urlModels {
			fs.SaveRecord(&nopWriteCloser{&buf}, model)
		}
		data := buf.Bytes()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			reader := bytes.NewReader(data)
			_, err := fs.LoadRecords(reader)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
