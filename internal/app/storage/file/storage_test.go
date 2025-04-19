package file

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/alexuryumtsev/go-shortener/internal/app/models"
	"github.com/stretchr/testify/assert"
)

func TestStorage_SaveAndLoad(t *testing.T) {
	filePath := "test_storage.json"
	defer os.Remove(filePath)

	storage := NewFileStorage(filePath)
	ctx := context.Background()
	urlModel1 := models.URLModel{ID: "4rSPg8ap", URL: "http://yandex.ru"}
	urlModel2 := models.URLModel{ID: "edVPg3ks", URL: "http://ya.ru"}
	urlModel3 := models.URLModel{ID: "dG56Hqxm", URL: "http://practicum.yandex.ru"}

	// Test Save
	err := storage.Save(ctx, urlModel1)
	assert.NoError(t, err)

	err = storage.Save(ctx, urlModel2)
	assert.NoError(t, err)

	err = storage.Save(ctx, urlModel3)
	assert.NoError(t, err)

	// Test Get
	result, exists := storage.Get(ctx, "4rSPg8ap")
	assert.True(t, exists)
	assert.Equal(t, urlModel1, result)

	result, exists = storage.Get(ctx, "edVPg3ks")
	assert.True(t, exists)
	assert.Equal(t, urlModel2, result)

	result, exists = storage.Get(ctx, "dG56Hqxm")
	assert.True(t, exists)
	assert.Equal(t, urlModel3, result)

	// Test LoadFromFile
	newStorage := NewFileStorage(filePath)
	err = newStorage.LoadFromFile()
	assert.NoError(t, err)

	result, exists = newStorage.Get(ctx, "4rSPg8ap")
	assert.True(t, exists)
	assert.Equal(t, urlModel1, result)

	result, exists = newStorage.Get(ctx, "edVPg3ks")
	assert.True(t, exists)
	assert.Equal(t, urlModel2, result)

	result, exists = newStorage.Get(ctx, "dG56Hqxm")
	assert.True(t, exists)
	assert.Equal(t, urlModel3, result)
}

func TestStorage_SaveToFileFormat(t *testing.T) {
	filePath := "test_storage_format.json"
	defer os.Remove(filePath)

	storage := NewFileStorage(filePath)
	ctx := context.Background()
	urlModel := models.URLModel{ID: "4rSPg8ap", URL: "http://yandex.ru", UserID: "1", Deleted: false}
	err := storage.Save(ctx, urlModel)
	assert.NoError(t, err)

	file, err := os.Open(filePath)
	assert.NoError(t, err)
	defer file.Close()

	var record struct {
		UUID        string `json:"uuid"`
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
		Deleted     bool   `json:"is_deleted"`
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&record)
	assert.NoError(t, err)

	assert.Equal(t, "1", record.UUID)
	assert.Equal(t, "4rSPg8ap", record.ShortURL)
	assert.Equal(t, "http://yandex.ru", record.OriginalURL)
}
