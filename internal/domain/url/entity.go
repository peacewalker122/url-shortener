package url

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrNotFound      = errors.New("url mapping not found")
	ErrInvalidURL    = errors.New("invalid long url")
	ErrMissingURL    = errors.New("missing url parameter")
	ErrDuplicateCode = errors.New("short code already exists")
)

type Mapping struct {
	ID       uuid.UUID `db:"id"`
	LongURL  string    `db:"long_url"`
	ShortURL string    `db:"short_url"`
}
