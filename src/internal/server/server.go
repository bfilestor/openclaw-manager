package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// Server 封装 HTTP 服务，便于后续注入路由与中间件。
type Server struct {
	httpServer *http.Server
}

func (s *Server) Handler() http.Handler {
	if s == nil || s.httpServer == nil {
		return nil
	}
	return s.httpServer.Handler
}

// New 创建带基础中间件的 HTTP 服务。
func New(addr string, staticDir string, registerFns ...func(*http.ServeMux)) *Server {
	mux := http.NewServeMux()

	// 健康检查接口（E1-S2-I4 要求）
	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"status":  "ok",
			"version": "dev",
		})
	})

	for _, fn := range registerFns {
		if fn != nil {
			fn(mux)
		}
	}

	// API 兜底 404（JSON）
	mux.HandleFunc("/api/v1/", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "not found",
			"code":  "NOT_FOUND",
		})
	})

	if staticDir != "" {
		mux.Handle("/", http.FileServer(http.Dir(staticDir)))
	}

	h := recoverMiddleware(corsMiddleware(requestLogMiddleware(mux)))

	return &Server{
		httpServer: &http.Server{
			Addr:    addr,
			Handler: h,
		},
	}
}

func (s *Server) Start() error {
	if s == nil || s.httpServer == nil {
		return errors.New("server not initialized")
	}
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Shutdown 优雅退出，默认 30 秒超时。
func (s *Server) Shutdown(parent context.Context) error {
	if s == nil || s.httpServer == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(parent, 30*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func requestLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("method=%s path=%s cost_ms=%d", r.Method, r.URL.Path, time.Since(start).Milliseconds())
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				_, file, line, ok := runtime.Caller(3)
				where := "unknown"
				if ok {
					where = fmt.Sprintf("%s:%d", filepath.Base(file), line)
				}
				writeJSON(w, http.StatusInternalServerError, map[string]any{
					"error":  "internal server error",
					"code":   "INTERNAL_ERROR",
					"detail": fmt.Sprintf("panic: %v", rec),
					"where":  where,
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// RunWithSignals 生产场景可直接调用，监听 SIGINT/SIGTERM。
func RunWithSignals(s *Server) error {
	errCh := make(chan error, 1)
	go func() { errCh <- s.Start() }()

	sigCh := make(chan os.Signal, 1)
	signalNotify(sigCh)

	select {
	case err := <-errCh:
		return err
	case <-sigCh:
		return s.Shutdown(context.Background())
	}
}
