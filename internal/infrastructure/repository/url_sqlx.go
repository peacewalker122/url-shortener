package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	domain "url-shortener/internal/domain/url"
)

type URLRepository struct {
	db *sqlx.DB
}

func NewURLRepository(db *sqlx.DB) URLRepository {
	return URLRepository{db: db}
}

func (r URLRepository) Save(ctx context.Context, mapping domain.Mapping) error {
	query := `
		INSERT INTO urls (long_url, short_url)
		VALUES ($1, $2)
	`

	_, err := r.db.ExecContext(ctx, query, mapping.LongURL, mapping.ShortURL)
	if err == nil {
		return nil
	}

	if isUniqueViolation(err) {
		return domain.ErrDuplicateCode
	}

	return fmt.Errorf("insert url mapping: %w", err)
}

func (r URLRepository) FindByShortCode(ctx context.Context, shortCode string) (domain.Mapping, error) {
	query := `
		SELECT id, long_url, short_url
		FROM urls
		WHERE short_url = $1
	`

	var mapping domain.Mapping
	err := r.db.GetContext(ctx, &mapping, query, shortCode)
	if err == nil {
		return mapping, nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return domain.Mapping{}, domain.ErrNotFound
	}

	if strings.Contains(strings.ToLower(err.Error()), "no rows") {
		return domain.Mapping{}, domain.ErrNotFound
	}

	return domain.Mapping{}, fmt.Errorf("query long url by short code: %w", err)
}

func isUniqueViolation(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate key") || strings.Contains(message, "unique")
}
