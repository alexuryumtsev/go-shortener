package storage

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/alexuryumtsev/go-shortener/internal/app/models"
)

type URLStorage interface {
	Save(urlModel models.URLModel) error
	Get(id string) (models.URLModel, bool)
	LoadFromFile() error
}

// Storage управляет сохранением и получением данных.
type Storage struct {
	mu       sync.RWMutex
	data     map[string]string
	filePath string
	counter  int
}

// NewStorage создаёт новое хранилище.
func NewStorage(filePath string) *Storage {
	return &Storage{
		data:     make(map[string]string),
		filePath: filePath,
		counter:  0,
	}
}

// Save сохраняет URL и записывает данные в файл.
func (s *Storage) Save(urlModel models.URLModel) error {
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

	return s.saveToFile(urlModel)
}

// Get возвращает оригинальный URL по идентификатору.
func (s *Storage) Get(id string) (models.URLModel, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	url, exists := s.data[id]
	return models.URLModel{ID: id, URL: url}, exists
}

// saveToFile сохраняет данные в файл.
func (s *Storage) saveToFile(urlModel models.URLModel) error {
	// Проверяем, существует ли директория, и создаем её, если не существует.
	if err := s.ensureDirExists(); err != nil {
		return err
	}

	file, err := os.OpenFile(s.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Формируем структуру с нужным порядком полей
	record := struct {
		UUID        string `json:"uuid"`
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}{
		UUID:        strconv.Itoa(s.counter),
		ShortURL:    urlModel.ID,
		OriginalURL: urlModel.URL,
	}

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(record); err != nil {
		return err
	}

	return nil
}

// LoadFromFile загружает данные из файла.
func (s *Storage) LoadFromFile() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Проверяем, существует ли директория, и создаем её, если не существует.
	if err := s.ensureDirExists(); err != nil {
		return err
	}

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

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var record map[string]string
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			return err
		}
		s.data[record["short_url"]] = record["original_url"]
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (s *Storage) ensureDirExists() error {
	dir := filepath.Dir(s.filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}
