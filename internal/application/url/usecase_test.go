package url

import (
	"context"
	"errors"
	"testing"

	domain "url-shortener/internal/domain/url"
)

type mockRepository struct {
	saveFn            func(ctx context.Context, mapping domain.Mapping) error
	findByShortCodeFn func(ctx context.Context, shortCode string) (domain.Mapping, error)
}

type mockCache struct {
	getFn func(ctx context.Context, shortCode string) (string, bool, error)
	setFn func(ctx context.Context, shortCode string, longURL string) error
}

func (m mockRepository) Save(ctx context.Context, mapping domain.Mapping) error {
	return m.saveFn(ctx, mapping)
}

func (m mockRepository) FindByShortCode(ctx context.Context, shortCode string) (domain.Mapping, error) {
	return m.findByShortCodeFn(ctx, shortCode)
}

func (m mockCache) Get(ctx context.Context, shortCode string) (string, bool, error) {
	if m.getFn == nil {
		return "", false, nil
	}
	return m.getFn(ctx, shortCode)
}

func (m mockCache) Set(ctx context.Context, shortCode string, longURL string) error {
	if m.setFn == nil {
		return nil
	}
	return m.setFn(ctx, shortCode, longURL)
}

func TestCreateShortURLSuccess(t *testing.T) {
	repo := mockRepository{
		saveFn: func(ctx context.Context, mapping domain.Mapping) error {
			if mapping.LongURL != "https://example.com/ok" {
				t.Fatalf("unexpected long URL: %s", mapping.LongURL)
			}
			return nil
		},
		findByShortCodeFn: func(ctx context.Context, shortCode string) (domain.Mapping, error) {
			return domain.Mapping{}, nil
		},
	}

	service := domain.NewShortCodeService()
	useCase := NewUseCase(repo, service, mockCache{})

	shortCode, err := useCase.CreateShortURL(context.Background(), "https://example.com/ok")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if shortCode == "" {
		t.Fatal("expected non-empty short code")
	}
}

func TestCreateShortURLErrorOnMissingURL(t *testing.T) {
	repo := mockRepository{
		saveFn: func(ctx context.Context, mapping domain.Mapping) error {
			return nil
		},
		findByShortCodeFn: func(ctx context.Context, shortCode string) (domain.Mapping, error) {
			return domain.Mapping{}, nil
		},
	}

	service := domain.NewShortCodeService()
	useCase := NewUseCase(repo, service, mockCache{})

	_, err := useCase.CreateShortURL(context.Background(), "")
	if !errors.Is(err, domain.ErrMissingURL) {
		t.Fatalf("expected ErrMissingURL, got %v", err)
	}
}

func TestResolveLongURLNotFound(t *testing.T) {
	repo := mockRepository{
		saveFn: func(ctx context.Context, mapping domain.Mapping) error {
			return nil
		},
		findByShortCodeFn: func(ctx context.Context, shortCode string) (domain.Mapping, error) {
			return domain.Mapping{}, domain.ErrNotFound
		},
	}

	service := domain.NewShortCodeService()
	useCase := NewUseCase(repo, service, mockCache{})

	_, err := useCase.ResolveLongURL(context.Background(), "missing")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestResolveLongURLReturnsCachedValue(t *testing.T) {
	repoCalled := false
	repo := mockRepository{
		saveFn: func(ctx context.Context, mapping domain.Mapping) error {
			return nil
		},
		findByShortCodeFn: func(ctx context.Context, shortCode string) (domain.Mapping, error) {
			repoCalled = true
			return domain.Mapping{}, nil
		},
	}

	cache := mockCache{
		getFn: func(ctx context.Context, shortCode string) (string, bool, error) {
			return "https://cached.example", true, nil
		},
	}

	service := domain.NewShortCodeService()
	useCase := NewUseCase(repo, service, cache)

	url, err := useCase.ResolveLongURL(context.Background(), "abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if url != "https://cached.example" {
		t.Fatalf("expected cached url, got %s", url)
	}

	if repoCalled {
		t.Fatal("repository should not be called on cache hit")
	}
}

func TestResolveLongURLFallsBackToRepositoryOnCacheError(t *testing.T) {
	repo := mockRepository{
		saveFn: func(ctx context.Context, mapping domain.Mapping) error {
			return nil
		},
		findByShortCodeFn: func(ctx context.Context, shortCode string) (domain.Mapping, error) {
			return domain.Mapping{LongURL: "https://db.example"}, nil
		},
	}

	cacheSetCalled := false
	cache := mockCache{
		getFn: func(ctx context.Context, shortCode string) (string, bool, error) {
			return "", false, errors.New("cache unavailable")
		},
		setFn: func(ctx context.Context, shortCode string, longURL string) error {
			cacheSetCalled = true
			return nil
		},
	}

	service := domain.NewShortCodeService()
	useCase := NewUseCase(repo, service, cache)

	url, err := useCase.ResolveLongURL(context.Background(), "abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if url != "https://db.example" {
		t.Fatalf("expected db url, got %s", url)
	}

	if !cacheSetCalled {
		t.Fatal("expected cache set after db fallback")
	}
}
