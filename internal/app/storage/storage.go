// Package storage предоставляет интерфейсы для работы с URL.
package storage

import (
	"context"

	"github.com/alexuryumtsev/go-shortener/internal/app/models"
)

// URLReader определяет методы для чтения URL.
type URLReader interface {
	Get(ctx context.Context, id string) (models.URLModel, bool)
	GetUserURLs(ctx context.Context, userID string) ([]models.URLModel, error)
	LoadFromFile() error
}

// URLWriter определяет методы для записи URL.
type URLWriter interface {
	Save(ctx context.Context, urlModel models.URLModel) error
	SaveBatch(ctx context.Context, urlModels []models.URLModel) error
	DeleteUserURLs(ctx context.Context, userID string, shortURLs []string) error
}

// GracefulCloser определяет метод для корректного закрытия хранилища.
type GracefulCloser interface {
	Close() error
}

// URLStorage объединяет интерфейсы URLReader, URLWriter и GracefulCloser.
type URLStorage interface {
	URLReader
	URLWriter
	GracefulCloser
}
