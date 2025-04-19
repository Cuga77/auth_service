package repository

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func setupTestDB(t *testing.T) *Postgres {
	connStr := "host=localhost port=5433 user=postgres password=admin dbname=auth_service sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	_, _ = db.Exec("DELETE FROM refresh_tokens")
	return &Postgres{db: db}
}

func TestPostgres_RefreshTokenFlow(t *testing.T) {
	if os.Getenv("CI") == "" {
		t.Skip("Тест требует запущенной тестовой БД (docker-compose up)")
	}
	repo := setupTestDB(t)
	defer repo.Close()
	ctx := context.Background()

	t.Run("Save and Get", func(t *testing.T) {
		err := repo.SaveRefreshToken(ctx, "user1", "hash1", "127.0.0.1")
		assert.NoError(t, err)

		tokens, err := repo.GetRefreshTokensByUser(ctx, "user1")
		assert.NoError(t, err)
		assert.Len(t, tokens, 1)
		assert.Equal(t, "user1", tokens[0].UserID)
	})

	t.Run("Delete", func(t *testing.T) {
		_ = repo.SaveRefreshToken(ctx, "user2", "hash2", "127.0.0.2")
		tokens, _ := repo.GetRefreshTokensByUser(ctx, "user2")
		assert.NotEmpty(t, tokens)

		err := repo.DeleteRefreshToken(ctx, tokens[0].ID)
		assert.NoError(t, err)

		tokens, _ = repo.GetRefreshTokensByUser(ctx, "user2")
		assert.Empty(t, tokens)
	})

	t.Run("RevokeAllTokens", func(t *testing.T) {
		_ = repo.SaveRefreshToken(ctx, "user3", "hash3", "127.0.0.3")
		_ = repo.SaveRefreshToken(ctx, "user3", "hash4", "127.0.0.3")
		tokens, _ := repo.GetRefreshTokensByUser(ctx, "user3")
		assert.Len(t, tokens, 2)

		err := repo.RevokeAllTokens(ctx, "user3")
		assert.NoError(t, err)

		tokens, _ = repo.GetRefreshTokensByUser(ctx, "user3")
		assert.Empty(t, tokens)
	})
}

func TestPostgres_GetRefreshToken(t *testing.T) {
	if os.Getenv("CI") == "" {
		t.Skip("Тест требует запущенной тестовой БД (docker-compose up)")
	}
	repo := setupTestDB(t)
	defer repo.Close()
	ctx := context.Background()

	_ = repo.SaveRefreshToken(ctx, "user4", "hash5", "127.0.0.4")
	tokens, _ := repo.GetRefreshTokensByUser(ctx, "user4")
	assert.NotEmpty(t, tokens)

	token, err := repo.GetRefreshToken(ctx, "hash5")
	assert.NoError(t, err)
	assert.Equal(t, "user4", token.UserID)
	assert.Equal(t, "hash5", token.TokenHash)
}

func TestPostgres_SaveRefreshToken_ExpiresAt(t *testing.T) {
	if os.Getenv("CI") == "" {
		t.Skip("Тест требует запущенной тестовой БД (docker-compose up)")
	}
	repo := setupTestDB(t)
	defer repo.Close()
	ctx := context.Background()

	err := repo.SaveRefreshToken(ctx, "user5", "hash6", "127.0.0.5")
	assert.NoError(t, err)

	tokens, err := repo.GetRefreshTokensByUser(ctx, "user5")
	assert.NoError(t, err)
	assert.WithinDuration(t, time.Now().Add(7*24*time.Hour), tokens[0].ExpiresAt, 2*time.Minute)
}
