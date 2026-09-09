package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/sinha-shaurya/agentic-bank/internal/config"
)

// Server owns the HTTP listener and the lifecycle around it.
type Server struct {
	cfg    config.Config
	logger *zap.Logger
	engine *gin.Engine
	http   *http.Server
}

// New builds the router and the underlying http.Server. Dependencies are
// passed in rather than reached for: one config and one logger for the life of
// the process.
func New(cfg config.Config, logger *zap.Logger, health Health) *Server {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// gin.New rather than gin.Default: the default text-mode request logger is
	// replaced by structured logging middleware, so we only keep Recovery.
	engine := gin.New()
	engine.Use(gin.Recovery())

	// Off by default in gin, which otherwise answers a wrong-verb request with
	// 404 and hides the real mistake from the caller.
	engine.HandleMethodNotAllowed = true

	engine.NoRoute(func(ctx *gin.Context) {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "route not found"})
	})
	engine.NoMethod(func(ctx *gin.Context) {
		ctx.JSON(http.StatusMethodNotAllowed, gin.H{"message": "method not allowed"})
	})

	registerHealthRoutes(engine, health)

	srv := &Server{
		cfg:    cfg,
		logger: logger,
		engine: engine,
	}

	srv.http = &http.Server{
		Addr:              cfg.Address(),
		Handler:           engine,
		ReadHeaderTimeout: 10 * time.Second,
	}

	return srv
}

// Engine exposes the router so later phases can mount their own route groups.
func (s *Server) Engine() *gin.Engine {
	return s.engine
}

// Run serves until ctx is cancelled, then drains in-flight requests within the
// configured shutdown timeout. A container gets SIGTERM before it is killed;
// draining means a transfer that is mid-commit is not cut off at the socket.
func (s *Server) Run(ctx context.Context) error {
	listenErr := make(chan error, 1)

	go func() {
		s.logger.Info("http server listening", zap.String("address", s.cfg.Address()))
		if err := s.http.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			listenErr <- err
			return
		}
		listenErr <- nil
	}()

	select {
	case err := <-listenErr:
		if err != nil {
			return fmt.Errorf("server: listen: %w", err)
		}
		return nil
	case <-ctx.Done():
	}

	s.logger.Info("shutdown signal received, draining",
		zap.Duration("timeout", s.cfg.ShutdownTimeout))

	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
	defer cancel()

	if err := s.http.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server: graceful shutdown: %w", err)
	}

	if err := <-listenErr; err != nil {
		return fmt.Errorf("server: listen: %w", err)
	}

	s.logger.Info("http server stopped cleanly")
	return nil
}
