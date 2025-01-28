package storage

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/alexuryumtsev/go-shortener/internal/app/models"
	"github.com/stretchr/testify/assert"
)

func TestStorage_SaveAndLoad(t *testing.T) {
	filePath := "test_storage.json"
	defer os.Remove(filePath)

	storage := NewStorage(filePath)

	urlModel1 := models.URLModel{ID: "4rSPg8ap", URL: "http://yandex.ru"}
	urlModel2 := models.URLModel{ID: "edVPg3ks", URL: "http://ya.ru"}
	urlModel3 := models.URLModel{ID: "dG56Hqxm", URL: "http://practicum.yandex.ru"}

	// Test Save
	err := storage.Save(urlModel1)
	assert.NoError(t, err)

	err = storage.Save(urlModel2)
	assert.NoError(t, err)

	err = storage.Save(urlModel3)
	assert.NoError(t, err)

	// Test Get
	result, exists := storage.Get("4rSPg8ap")
	assert.True(t, exists)
	assert.Equal(t, urlModel1, result)

	result, exists = storage.Get("edVPg3ks")
	assert.True(t, exists)
	assert.Equal(t, urlModel2, result)

	result, exists = storage.Get("dG56Hqxm")
	assert.True(t, exists)
	assert.Equal(t, urlModel3, result)

	// Test LoadFromFile
	newStorage := NewStorage(filePath)
	err = newStorage.LoadFromFile()
	assert.NoError(t, err)

	result, exists = newStorage.Get("4rSPg8ap")
	assert.True(t, exists)
	assert.Equal(t, urlModel1, result)

	result, exists = newStorage.Get("edVPg3ks")
	assert.True(t, exists)
	assert.Equal(t, urlModel2, result)

	result, exists = newStorage.Get("dG56Hqxm")
	assert.True(t, exists)
	assert.Equal(t, urlModel3, result)
}

func TestStorage_SaveToFileFormat(t *testing.T) {
	filePath := "test_storage_format.json"
	defer os.Remove(filePath)

	storage := NewStorage(filePath)

	urlModel := models.URLModel{ID: "4rSPg8ap", URL: "http://yandex.ru"}
	err := storage.Save(urlModel)
	assert.NoError(t, err)

	file, err := os.Open(filePath)
	assert.NoError(t, err)
	defer file.Close()

	var record map[string]string
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&record)
	assert.NoError(t, err)

	assert.Equal(t, "1", record["uuid"])
	assert.Equal(t, "4rSPg8ap", record["short_url"])
	assert.Equal(t, "http://yandex.ru", record["original_url"])
}
