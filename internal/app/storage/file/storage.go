// Package file предоставляет реализацию хранилища для работы с файлами.
package file

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
	data        map[string]models.URLModel
	userData    map[string][]models.URLModel
	filePath    string
	counter     int
	fileStorage *fileutils.FileStorage
}

// NewFileStorage создаёт новое файловое хранилище.
func NewFileStorage(filePath string) *FileStorage {
	return &FileStorage{
		data:        make(map[string]models.URLModel),
		userData:    make(map[string][]models.URLModel),
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
		if existingURL.URL == urlModel.URL {
			// Если URL уже существует, ничего не делаем
			return nil
		}
	}

	s.data[urlModel.ID] = urlModel
	userID := urlModel.UserID
	s.userData[userID] = append(s.userData[userID], urlModel)

	file, err := os.OpenFile(s.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	return s.fileStorage.SaveRecord(file, urlModel)
}

// SaveBatch сохраняет множество URL в файл.
func (s *FileStorage) SaveBatch(ctx context.Context, urlModels []models.URLModel) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.OpenFile(s.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, urlModel := range urlModels {
		// Проверяем, существует ли уже оригинальный URL
		for _, existingURL := range s.data {
			if existingURL.URL == urlModel.URL {
				// Если URL уже существует, ничего не делаем
				continue
			}
		}

		userID := urlModel.UserID
		s.data[urlModel.ID] = urlModel
		s.userData[userID] = append(s.userData[userID], urlModel)

		if err := s.fileStorage.SaveRecord(file, urlModel); err != nil {
			return err
		}
	}

	return nil
}

// Get возвращает оригинальный URL по идентификатору.
func (s *FileStorage) Get(ctx context.Context, id string) (models.URLModel, bool) {
	s.LoadFromFile()
	s.mu.RLock()
	defer s.mu.RUnlock()
	urlModel, exists := s.data[id]
	return models.URLModel{ID: id, URL: urlModel.URL, Deleted: urlModel.Deleted}, exists
}

// GetUserURLs возвращает все URL, сокращённые пользователем.
func (s *FileStorage) GetUserURLs(ctx context.Context, userID string) ([]models.URLModel, error) {
	s.LoadFromFile()
	s.mu.RLock()
	defer s.mu.RUnlock()
	urls, exists := s.userData[userID]
	if !exists {
		return nil, nil
	}
	return urls, nil
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
	for shortURL, urlModel := range data {
		if shortURL == "" || urlModel.URL == "" {
			return fmt.Errorf("invalid data format: short_url or original_url is empty")
		}
	}

	s.data = data
	return nil
}

// Ping проверяет соединение с базой данных (для файлового хранилища всегда возвращает nil).
func (s *FileStorage) Ping(ctx context.Context) error {
	return nil
}

// Close закрывает файловое хранилище.
func (s *FileStorage) DeleteUserURLs(ctx context.Context, userID string, shortURLs []string) error {
	// Загружаем все записи из файла
	if err := s.LoadFromFile(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, shortURL := range shortURLs {
		if urlModel, exists := s.data[shortURL]; exists && urlModel.UserID == userID {
			urlModel.Deleted = true
			s.data[shortURL] = urlModel
		}
	}

	// Открываем файл для записи и очищаем его перед записью
	file, err := os.OpenFile(s.filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Записываем все обновленные записи обратно в файл
	for _, urlModel := range s.data {
		if err := s.fileStorage.SaveRecord(file, urlModel); err != nil {
			return err
		}
	}

	return nil
}
