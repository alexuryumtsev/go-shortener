package storage

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/alexuryumtsev/go-shortener/internal/app/fileutils"
	"github.com/alexuryumtsev/go-shortener/internal/app/models"
)

// FileStorage управляет сохранением и получением данных в файле.
type FileStorage struct {
	mu          sync.RWMutex
	data        map[string]string
	filePath    string
	counter     int
	fileStorage *fileutils.FileStorage
}

// NewFileStorage создаёт новое файловое хранилище.
func NewFileStorage(filePath string) *FileStorage {
	return &FileStorage{
		data:        make(map[string]string),
		filePath:    filePath,
		counter:     0,
		fileStorage: fileutils.NewFileStorage(filePath),
	}
}

// Save сохраняет URL и записывает данные в файл.
func (s *FileStorage) Save(ctx context.Context, urlModel models.URLModel) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Проверяем, существует ли уже оригинальный URL
	for _, existingURL := range s.data {
		if existingURL == urlModel.URL {
			// Если URL уже существует, ничего не делаем
			return nil
		}
	}

	s.data[urlModel.ID] = urlModel.URL
	s.counter++

	file, err := os.OpenFile(s.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	return s.fileStorage.SaveRecord(file, s.counter, urlModel)
}

// Get возвращает оригинальный URL по идентификатору.
func (s *FileStorage) Get(ctx context.Context, id string) (models.URLModel, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	url, exists := s.data[id]
	return models.URLModel{ID: id, URL: url}, exists
}

// LoadFromFile загружает данные из файла.
func (s *FileStorage) LoadFromFile() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Open(s.filePath)
	if os.IsNotExist(err) {
		// Если файл не существует, создаем его.
		file, err = os.Create(s.filePath)
		if err != nil {
			return err
		}
		defer file.Close()
		return nil
	} else if err != nil {
		return err
	}
	defer file.Close()

	data, err := s.fileStorage.LoadRecords(file)
	if err != nil {
		return err
	}

	// Валидация формата данных
	for shortURL, originalURL := range data {
		if shortURL == "" || originalURL == "" {
			return fmt.Errorf("invalid data format: short_url or original_url is empty")
		}
	}

	s.data = data
	return nil
}
