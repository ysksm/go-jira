package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ysksm/go-jira/core/infrastructure/config"
	"github.com/ysksm/go-jira/core/infrastructure/database"
)

// ServerConfig holds server configuration.
type ServerConfig struct {
	Port        int
	CORSOrigin  string
	ConfigStore *config.FileConfigStore
	ConnMgr     *database.Connection
	Logger      *slog.Logger
}

// Server is the HTTP API server.
type Server struct {
	httpServer *http.Server
	config     ServerConfig
	logger     *slog.Logger
}

// NewServer creates a new API server.
func NewServer(cfg ServerConfig) *Server {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	router := NewRouter(cfg)
	handler := applyMiddleware(router, cfg)

	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.Port),
			Handler:      handler,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 120 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		config: cfg,
		logger: cfg.Logger,
	}
}

// applyMiddleware wraps the handler with middleware.
func applyMiddleware(handler http.Handler, cfg ServerConfig) http.Handler {
	h := recoveryMiddleware(handler)
	h = loggingMiddleware(h)
	if cfg.CORSOrigin != "" {
		h = corsMiddleware(cfg.CORSOrigin)(h)
	}
	return h
}

// Start starts the server and blocks until shutdown signal.
func (s *Server) Start() error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	errCh := make(chan error, 1)
	go func() {
		s.logger.Info("server starting", "addr", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	case sig := <-stop:
		s.logger.Info("shutdown signal received", "signal", sig)
	}

	return s.Shutdown()
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.logger.Info("shutting down server")
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown error: %w", err)
	}
	s.logger.Info("server stopped")
	return nil
}
