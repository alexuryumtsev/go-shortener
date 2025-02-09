package storage

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/alexuryumtsev/go-shortener/internal/app/db"
	"github.com/alexuryumtsev/go-shortener/internal/app/fileutils"
	"github.com/alexuryumtsev/go-shortener/internal/app/models"
)

// URLReader определяет методы для чтения URL.
type URLReader interface {
	Get(ctx context.Context, id string) (models.URLModel, bool)
	LoadFromFile() error
}

// URLWriter определяет методы для записи URL.
type URLWriter interface {
	Save(ctx context.Context, urlModel models.URLModel) error
}

// URLStorage объединяет интерфейсы URLReader и URLWriter.
type URLStorage interface {
	URLReader
	URLWriter
	Ping(ctx context.Context) error
}

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

// Ping проверяет соединение с базой данных (для файлового хранилища всегда возвращает nil).
func (s *FileStorage) Ping(ctx context.Context) error {
	return nil
}

// DatabaseStorage управляет сохранением и получением данных в базе данных.
type DatabaseStorage struct {
	db *db.Database
}

// NewDatabaseStorage создаёт новое хранилище для базы данных.
func NewDatabaseStorage(db *db.Database) *DatabaseStorage {
	return &DatabaseStorage{db: db}
}

// Save сохраняет URL в базе данных.
func (s *DatabaseStorage) Save(ctx context.Context, urlModel models.URLModel) error {
	// Реализация сохранения в базе данных
	return nil
}

// Get возвращает оригинальный URL по идентификатору из базы данных.
func (s *DatabaseStorage) Get(ctx context.Context, id string) (models.URLModel, bool) {
	// Реализация получения из базы данных
	return models.URLModel{}, false
}

// LoadFromFile загружает данные из базы данных (не требуется для базы данных).
func (s *DatabaseStorage) LoadFromFile() error {
	return nil
}

// Ping проверяет соединение с базой данных.
func (s *DatabaseStorage) Ping(ctx context.Context) error {
	return s.db.Ping(ctx)
}
