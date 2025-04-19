package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/auth-service/internal/models"

	"github.com/auth-service/config"
)

type Postgres struct {
	db *sql.DB
}

func (p *Postgres) DB() *sql.DB {
	return p.db
}

func (p *Postgres) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

func NewPostgres(cfg *config.Config) (*Postgres, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open DB connection: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping DB: %w", err)
	}

	if err := RunMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &Postgres{db: db}, nil
}

func (p *Postgres) SaveRefreshToken(ctx context.Context, userID, tokenHash, ip string) error {
	persistCtx := context.WithoutCancel(ctx)

	_, err := p.db.ExecContext(
		persistCtx,
		`INSERT INTO refresh_tokens (user_id, token_hash, ip, expires_at) 
         VALUES ($1, $2, $3, NOW() + INTERVAL '7 days')`,
		userID,
		tokenHash,
		ip,
	)

	if err != nil {
		return fmt.Errorf("failed to save refresh token for user %s: %w", userID, err)
	}

	log.Printf("Successfully saved refresh token for user %s from IP %s", userID, ip)
	return nil
}

func (p *Postgres) GetRefreshToken(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
	var token models.RefreshToken
	err := p.db.QueryRowContext(ctx,
		`SELECT id, user_id, token_hash, ip, expires_at, created_at 
		FROM refresh_tokens 
		WHERE token_hash = $1`,
		tokenHash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.IP,
		&token.ExpiresAt,
		&token.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("token not found")
		}
		return nil, fmt.Errorf("failed to get token: %w", err)
	}
	return &token, nil
}

func (p *Postgres) GetRefreshTokensByUser(ctx context.Context, userID string) ([]models.RefreshToken, error) {
	rows, err := p.db.QueryContext(ctx,
		`SELECT id, user_id, token_hash, ip, expires_at, created_at 
		FROM refresh_tokens 
		WHERE user_id = $1`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh tokens: %w", err)
	}
	defer rows.Close()

	var tokens []models.RefreshToken
	for rows.Next() {
		var token models.RefreshToken
		if err := rows.Scan(
			&token.ID,
			&token.UserID,
			&token.TokenHash,
			&token.IP,
			&token.ExpiresAt,
			&token.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan token: %w", err)
		}
		tokens = append(tokens, token)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tokens: %w", err)
	}

	return tokens, nil
}

func (p *Postgres) DeleteRefreshToken(ctx context.Context, id string) error {
	_, err := p.db.ExecContext(ctx,
		`DELETE FROM refresh_tokens WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}
	return nil
}

func (p *Postgres) RevokeAllTokens(ctx context.Context, userID string) error {
	_, err := p.db.ExecContext(ctx,
		`DELETE FROM refresh_tokens WHERE user_id = $1`,
		userID)
	if err != nil {
		return fmt.Errorf("failed to revoke tokens: %w", err)
	}
	return nil
}
