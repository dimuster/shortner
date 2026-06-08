package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("link not found")

type LinkRepository struct {
	pool *pgxpool.Pool
}

func NewLinkRepository(pool *pgxpool.Pool) *LinkRepository {
	return &LinkRepository{pool: pool}
}

func (r *LinkRepository) GetOriginalURL(ctx context.Context, shortCode string) (string, error) {
	var originalURL string

	err := r.pool.QueryRow(ctx,
		`SELECT original_url FROM links WHERE short_code = $1`,
		shortCode,
	).Scan(&originalURL)

	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}

	return originalURL, nil
}

func (r *LinkRepository) CreateLink(ctx context.Context, shortCode, originalURL string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO links (short_code, original_url) VALUES ($1, $2)`,
		shortCode, originalURL,
	)
	return err
}
