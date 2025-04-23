package pg

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/alexuryumtsev/go-shortener/internal/app/db"
	"github.com/alexuryumtsev/go-shortener/internal/app/models"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultTestDBHost = "localhost"
	defaultTestDBPort = "5432"
	defaultTestDBUser = "postgres"
	defaultTestDBPass = "postgres"
)

func setupTestDB(t *testing.T) (*DatabaseStorage, func()) {
	ctx := context.Background()

	// Connect to default postgres database first
	mainDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=disable",
		defaultTestDBUser, defaultTestDBPass, defaultTestDBHost, defaultTestDBPort)

	mainPool, err := pgxpool.New(ctx, mainDSN)
	require.NoError(t, err, "Failed to connect to postgres database")
	defer mainPool.Close()

	// Create test database
	testDBName := fmt.Sprintf("test_db_%d", os.Getpid())
	_, err = mainPool.Exec(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDBName))
	require.NoError(t, err, "Failed to drop test database")

	_, err = mainPool.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", testDBName))
	require.NoError(t, err, "Failed to create test database")

	// Connect to test database
	testDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		defaultTestDBUser, defaultTestDBPass, defaultTestDBHost, defaultTestDBPort, testDBName)

	pool, err := db.NewDatabaseConnection(ctx, testDSN)
	require.NoError(t, err, "Failed to connect to test database")

	// Create cleanup function
	cleanup := func() {
		pool.Close()
		mainPool, err := pgxpool.New(ctx, mainDSN)
		if err == nil {
			defer mainPool.Close()
			_, _ = mainPool.Exec(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDBName))
		}
	}

	return NewDatabaseStorage(pool), cleanup
}

func TestDatabaseStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database tests in short mode")
	}

	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("Save and Get", func(t *testing.T) {
		urlModel := models.URLModel{
			ID:     "test123",
			URL:    "https://example.com",
			UserID: "user1",
		}

		err := db.Save(ctx, urlModel)
		assert.NoError(t, err)

		retrieved, exists := db.Get(ctx, urlModel.ID)
		assert.True(t, exists)
		assert.Equal(t, urlModel.URL, retrieved.URL)
	})

	t.Run("GetUserURLs", func(t *testing.T) {
		userID := "user1"
		urls, err := db.GetUserURLs(ctx, userID)
		assert.NoError(t, err)
		assert.NotEmpty(t, urls)
	})
}
