package http

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"url-shortener/internal/shared/config"
)

func NewRouter(cfg config.Config, logger *slog.Logger, handler Handler, pool *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(cfg.RequestTimeout))
	r.Use(requestLogger(logger))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.CORSAllowOrigin},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(r.Context()); err != nil {
			logger.Error(
				"health check failed",
				"component", "http_router",
				"path", r.URL.Path,
				"method", r.Method,
				"error", err,
			)
			writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{Detail: "database unavailable"})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Post("/shorten", handler.CreateShortURL)
	r.Get("/{key}", handler.GetURL)

	return r
}

func requestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			logger.Info(
				"request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"latency_ms", time.Since(started).Milliseconds(),
			)
		})
	}
}
