package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	domain "url-shortener/internal/domain/url"
)

type fakeUseCase struct {
	createFn  func(longURL string) (string, error)
	resolveFn func(shortCode string) (string, error)
}

func (f fakeUseCase) CreateShortURL(_ context.Context, longURL string) (string, error) {
	return f.createFn(longURL)
}

func (f fakeUseCase) ResolveLongURL(_ context.Context, shortCode string) (string, error) {
	return f.resolveFn(shortCode)
}

func TestCreateShortURLHandlerSuccess(t *testing.T) {
	useCase := fakeUseCase{
		createFn: func(longURL string) (string, error) {
			if longURL != "https://example.com/a" {
				t.Fatalf("unexpected longURL: %s", longURL)
			}
			return "abc123X", nil
		},
		resolveFn: func(shortCode string) (string, error) { return "", nil },
	}

	h := Handler{useCase: useCase}
	body, _ := json.Marshal(CreateShortURLRequest{URL: "https://example.com/a"})
	req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBuffer(body))
	res := httptest.NewRecorder()

	h.CreateShortURL(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
}

func TestGetURLHandlerNotFound(t *testing.T) {
	useCase := fakeUseCase{
		createFn:  func(longURL string) (string, error) { return "", nil },
		resolveFn: func(shortCode string) (string, error) { return "", domain.ErrNotFound },
	}

	h := Handler{useCase: useCase}
	req := httptest.NewRequest(http.MethodGet, "/shorten/missing", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("key", "missing")
	req = req.WithContext(contextWithRoute(req, rctx))
	res := httptest.NewRecorder()

	h.GetURL(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", res.Code)
	}
}

func contextWithRoute(req *http.Request, rctx *chi.Context) context.Context {
	ctx := req.Context()
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	return ctx
}

func TestCreateShortURLHandlerBadPayload(t *testing.T) {
	useCase := fakeUseCase{
		createFn:  func(longURL string) (string, error) { return "", errors.New("should not be called") },
		resolveFn: func(shortCode string) (string, error) { return "", nil },
	}

	h := Handler{useCase: useCase}
	req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBufferString("not-json"))
	res := httptest.NewRecorder()

	h.CreateShortURL(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.Code)
	}
}
