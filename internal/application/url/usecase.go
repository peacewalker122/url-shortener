package url

import (
	"context"

	domain "url-shortener/internal/domain/url"

	"github.com/google/uuid"
)

type UseCase struct {
	repository domain.Repository
	service    domain.ShortCodeService
	cache      URLCache
}

type URLUseCase interface {
	CreateShortURL(ctx context.Context, longURL string) (string, error)
	ResolveLongURL(ctx context.Context, shortCode string) (string, error)
}

type URLCache interface {
	Get(ctx context.Context, shortCode string) (string, bool, error)
	Set(ctx context.Context, shortCode string, longURL string) error
}

func NewUseCase(repository domain.Repository, service domain.ShortCodeService, cache URLCache) UseCase {
	return UseCase{
		repository: repository,
		service:    service,
		cache:      cache,
	}
}

func (u UseCase) CreateShortURL(ctx context.Context, longURL string) (string, error) {
	if longURL == "" {
		return "", domain.ErrMissingURL
	}

	if !u.service.ValidateLongURL(longURL) {
		return "", domain.ErrInvalidURL
	}

	for range 5 {
		id := u.service.GenerateID()
		shortCode := u.service.EncodeBase62(id)
		mapping := domain.Mapping{ID: uuid.New(), LongURL: longURL, ShortURL: shortCode}

		err := u.repository.Save(ctx, mapping)
		if err == nil {
			if u.cache != nil {
				_ = u.cache.Set(ctx, shortCode, longURL)
			}
			return shortCode, nil
		}

		if err != domain.ErrDuplicateCode {
			return "", err
		}
	}

	return "", domain.ErrDuplicateCode
}

func (u UseCase) ResolveLongURL(ctx context.Context, shortCode string) (string, error) {
	if u.cache != nil {
		cachedURL, found, err := u.cache.Get(ctx, shortCode)
		if err == nil && found {
			return cachedURL, nil
		}
	}

	mapping, err := u.repository.FindByShortCode(ctx, shortCode)
	if err != nil {
		return "", err
	}

	if u.cache != nil {
		_ = u.cache.Set(ctx, shortCode, mapping.LongURL)
	}

	return mapping.LongURL, nil
}
