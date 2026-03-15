package profiling

import (
	"context"
	"encoding/json"
	"expvar"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"runtime"
	"time"

	"url-shortener/internal/shared/config"
)

type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
	cfg        config.Config
	stopStats  chan struct{}
}

func NewServer(cfg config.Config, logger *slog.Logger) *Server {
	if !cfg.ProfilingEnabled {
		return nil
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	mux.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	mux.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	mux.Handle("/debug/pprof/block", pprof.Handler("block"))
	mux.Handle("/debug/pprof/mutex", pprof.Handler("mutex"))
	mux.Handle("/debug/vars", expvar.Handler())
	mux.HandleFunc("/debug/stats", runtimeStatsHandler)

	httpServer := &http.Server{
		Addr:              cfg.ProfilingAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &Server{
		httpServer: httpServer,
		logger:     logger,
		cfg:        cfg,
		stopStats:  make(chan struct{}),
	}
}

func (s *Server) Start() {
	if s == nil {
		return
	}

	if s.cfg.ProfilingRuntimeStatsEnabled {
		go s.logRuntimeStats()
	}

	go func() {
		s.logger.Info("profiling server enabled", "addr", s.cfg.ProfilingAddr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("profiling server failed", "error", err)
		}
	}()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s == nil {
		return nil
	}

	close(s.stopStats)
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) logRuntimeStats() {
	ticker := time.NewTicker(s.cfg.ProfilingRuntimeStatsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopStats:
			return
		case <-ticker.C:
			var mem runtime.MemStats
			runtime.ReadMemStats(&mem)
			s.logger.Info(
				"runtime stats",
				"goroutines", runtime.NumGoroutine(),
				"alloc_bytes", mem.Alloc,
				"heap_inuse_bytes", mem.HeapInuse,
				"heap_idle_bytes", mem.HeapIdle,
				"sys_bytes", mem.Sys,
				"gc_cycles", mem.NumGC,
			)
		}
	}
}

type runtimeStats struct {
	Goroutines uint64 `json:"goroutines"`
	Memory     struct {
		Alloc      uint64 `json:"alloc_bytes"`
		HeapInUse  uint64 `json:"heap_inuse_bytes"`
		HeapIdle   uint64 `json:"heap_idle_bytes"`
		System     uint64 `json:"sys_bytes"`
		TotalAlloc uint64 `json:"total_alloc_bytes"`
	} `json:"memory"`
	GC struct {
		Cycles uint32 `json:"cycles"`
	} `json:"gc"`
}

func runtimeStatsHandler(w http.ResponseWriter, _ *http.Request) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	stats := runtimeStats{Goroutines: uint64(runtime.NumGoroutine())}
	stats.Memory.Alloc = mem.Alloc
	stats.Memory.HeapInUse = mem.HeapInuse
	stats.Memory.HeapIdle = mem.HeapIdle
	stats.Memory.System = mem.Sys
	stats.Memory.TotalAlloc = mem.TotalAlloc
	stats.GC.Cycles = mem.NumGC

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(stats)
}
