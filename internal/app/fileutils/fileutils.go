package fileutils

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/alexuryumtsev/go-shortener/internal/app/models"
)

// EnsureDirExists проверяет, существует ли директория, и создаёт её, если не существует.
func EnsureDirExists(filePath string) error {
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

// FileStorage представляет собой хранилище, использующее файлы для персистентности.
type FileStorage struct {
	filePath string
}

// NewFileStorage создает новое файловое хранилище.
func NewFileStorage(filePath string) *FileStorage {
	return &FileStorage{filePath: filePath}
}

// SaveRecord сохраняет запись в файл.
func (fs *FileStorage) SaveRecord(w io.WriteCloser, urlModel models.URLModel) error {
	defer w.Close()

	bufferedWriter := bufio.NewWriter(w)
	defer bufferedWriter.Flush()

	record := struct {
		UUID        string `json:"uuid"`
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
		DeletedFlag bool   `json:"is_deleted"`
	}{
		UUID:        urlModel.UserID,
		ShortURL:    urlModel.ID,
		OriginalURL: urlModel.URL,
		DeletedFlag: urlModel.Deleted,
	}

	encoder := json.NewEncoder(bufferedWriter)
	if err := encoder.Encode(record); err != nil {
		return err
	}

	return nil
}

// LoadRecords загружает записи из файла.
func (fs *FileStorage) LoadRecords(r io.Reader) (map[string]models.URLModel, error) {
	data := make(map[string]models.URLModel)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		var record struct {
			UUID        string `json:"uuid"`
			ShortURL    string `json:"short_url"`
			OriginalURL string `json:"original_url"`
			Deleted     bool   `json:"is_deleted"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			return nil, err
		}
		data[record.ShortURL] = models.URLModel{
			ID:      record.ShortURL,
			URL:     record.OriginalURL,
			UserID:  record.UUID,
			Deleted: record.Deleted,
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return data, nil
}
