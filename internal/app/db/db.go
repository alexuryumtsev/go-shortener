// Package db предоставляет функции для работы с базой данных PostgreSQL
package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	Pool *pgxpool.Pool
}

// NewDatabaseConnection создает новое подключение к PostgreSQL
func NewDatabaseConnection(ctx context.Context, dsn string) (*Database, error) {
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("error parsing DSN: %w", err)
	}

	// Устанавливаем таймауты
	poolConfig.ConnConfig.ConnectTimeout = 5 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("database connection error: %w", err)
	}

	db := &Database{Pool: pool}

	// Создаем таблицы
	if err := db.createTables(ctx); err != nil {
		return nil, fmt.Errorf("error creating tables: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL and ensured tables exist")

	return db, nil
}

// createTables создает необходимые таблицы в базе данных
func (db *Database) createTables(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS urls (
		id SERIAL PRIMARY KEY,
		user_id VARCHAR(255) NOT NULL,
		short_url VARCHAR(255) NOT NULL UNIQUE,
		original_url TEXT NOT NULL,
		is_deleted BOOLEAN DEFAULT FALSE
	);

	CREATE INDEX IF NOT EXISTS idx_user_short_url ON urls (user_id, short_url);
	`
	_, err := db.Pool.Exec(ctx, query)
	return err
}

// Close закрывает соединение с базой данных
func (db *Database) Close() {
	db.Pool.Close()
	log.Println("database connection closed")
}

// Ping проверяет соединение с базой данных
func (db *Database) Ping(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}
