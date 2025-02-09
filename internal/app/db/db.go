package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	pool *pgxpool.Pool
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

	return &Database{pool: pool}, nil
}

// Close закрывает соединение с базой данных
func (db *Database) Close() {
	db.pool.Close()
	log.Println("database connection closed")
}

// Ping проверяет соединение с базой данных
func (db *Database) Ping(ctx context.Context) error {
	return db.pool.Ping(ctx)
}
