package storage

import (
	"context"

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
}
