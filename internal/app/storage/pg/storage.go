// Package pg реализует интерфейс Storage для работы с базой данных PostgreSQL.
package pg

import (
	"context"
	"fmt"

	"github.com/alexuryumtsev/go-shortener/internal/app/db"
	"github.com/alexuryumtsev/go-shortener/internal/app/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

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
	query := `INSERT INTO urls (user_id, short_url, original_url) VALUES ($1, $2, $3)`
	_, err := s.db.Pool.Exec(ctx, query, urlModel.UserID, urlModel.ID, urlModel.URL)

	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == pgerrcode.UniqueViolation {
			return fmt.Errorf("conflict: %w", err)
		}
		return fmt.Errorf("failed to save URL: %w", err)
	}
	return nil
}

// SaveBatch сохраняет множество URL в базе данных.
func (s *DatabaseStorage) SaveBatch(ctx context.Context, urlModels []models.URLModel) error {
	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			// Если произошла ошибка при rollback, добавляем её к основной ошибке
			if err != nil {
				err = fmt.Errorf("rollback error: %v: %w", rollbackErr, err)
			} else {
				err = rollbackErr
			}
		}
	}()

	for _, urlModel := range urlModels {
		query := `INSERT INTO urls (user_id, short_url, original_url) VALUES ($1, $2, $3) ON CONFLICT (short_url) DO NOTHING`
		_, err := tx.Exec(ctx, query, urlModel.UserID, urlModel.ID, urlModel.URL)
		if err != nil {
			return fmt.Errorf("failed to save URL: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// Get возвращает оригинальный URL по идентификатору из базы данных.
func (s *DatabaseStorage) Get(ctx context.Context, id string) (models.URLModel, bool) {
	query := `SELECT original_url, is_deleted FROM urls WHERE short_url = $1`
	row := s.db.Pool.QueryRow(ctx, query, id)

	var urlModel models.URLModel
	urlModel.ID = id
	err := row.Scan(&urlModel.URL, &urlModel.Deleted)
	if err != nil {
		return models.URLModel{}, false
	}
	return urlModel, true
}

// GetUserURLs возвращает все URL, сокращённые пользователем.
func (s *DatabaseStorage) GetUserURLs(ctx context.Context, userID string) ([]models.URLModel, error) {
	query := `SELECT short_url, original_url FROM urls WHERE is_deleted IS FALSE AND user_id = $1`
	rows, err := s.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user URLs: %w", err)
	}
	defer rows.Close()

	var urls []models.URLModel
	for rows.Next() {
		var urlModel models.URLModel
		if err := rows.Scan(&urlModel.ID, &urlModel.URL); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		urls = append(urls, urlModel)
	}

	return urls, nil
}

// LoadFromFile загружает данные из базы данных (не требуется для базы данных).
func (s *DatabaseStorage) LoadFromFile() error {
	return nil
}

// Ping проверяет соединение с базой данных.
func (s *DatabaseStorage) Ping(ctx context.Context) error {
	return s.db.Ping(ctx)
}

func (s *DatabaseStorage) DeleteUserURLs(ctx context.Context, userID string, shortURLs []string) error {
	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			// Если произошла ошибка при rollback, добавляем её к основной ошибке
			if err != nil {
				err = fmt.Errorf("rollback error: %v: %w", rollbackErr, err)
			} else {
				err = rollbackErr
			}
		}
	}()

	query := `
        UPDATE urls
        SET is_deleted = true
        WHERE user_id = $1
        AND short_url = ANY($2)`

	_, err = tx.Exec(ctx, query, userID, shortURLs)
	if err != nil {
		return fmt.Errorf("failed to delete URLs: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}
