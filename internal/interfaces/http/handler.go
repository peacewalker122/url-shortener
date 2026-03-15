package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	appurl "url-shortener/internal/application/url"
	domain "url-shortener/internal/domain/url"
)

type Handler struct {
	useCase appurl.URLUseCase
	logger  *slog.Logger
}

func NewHandler(useCase appurl.URLUseCase, logger *slog.Logger) Handler {
	return Handler{useCase: useCase, logger: logger}
}

func (h Handler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	var payload CreateShortURLRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.logRequestError(r, http.StatusBadRequest, "invalid JSON body", err)
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Detail: "invalid JSON body"})
		return
	}

	shortCode, err := h.useCase.CreateShortURL(r.Context(), payload.URL)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrMissingURL):
			h.logRequestError(r, http.StatusBadRequest, "missing url parameter", err)
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Detail: "Missing url parameter"})
		case errors.Is(err, domain.ErrInvalidURL):
			h.logRequestError(r, http.StatusBadRequest, "invalid url parameter", err)
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Detail: "Invalid url parameter"})
		default:
			h.logRequestError(r, http.StatusInternalServerError, "failed to create short url", err)
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Detail: "internal server error"})
		}
		return
	}

	writeJSON(w, http.StatusOK, ShortURLResponse{URL: shortCode})
}

func (h Handler) GetURL(w http.ResponseWriter, r *http.Request) {
	shortCode := chi.URLParam(r, "key")
	if shortCode == "" {
		h.logRequestError(r, http.StatusBadRequest, "missing key parameter", domain.ErrMissingURL)
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Detail: "missing key parameter"})
		return
	}

	longURL, err := h.useCase.ResolveLongURL(r.Context(), shortCode)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			h.logRequestError(r, http.StatusNotFound, "short code not found", err)
			writeJSON(w, http.StatusNotFound, ErrorResponse{Detail: "URL not found"})
		default:
			h.logRequestError(r, http.StatusInternalServerError, "failed to resolve short url", err)
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Detail: "internal server error"})
		}
		return
	}

	http.Redirect(w, r, longURL, http.StatusFound)
}

func (h Handler) logRequestError(r *http.Request, status int, message string, err error) {
	if h.logger == nil {
		return
	}

	requestID := middleware.GetReqID(r.Context())
	h.logger.Error(
		message,
		"component", "http_handler",
		"method", r.Method,
		"path", r.URL.Path,
		"status", status,
		"request_id", requestID,
		"error", err,
	)
}
