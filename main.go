package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/sinha-shaurya/agentic-bank/internal/config"
	"github.com/sinha-shaurya/agentic-bank/internal/obs"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger, err := obs.NewLogger(cfg.LogLevel, cfg.IsProduction())
	if err != nil {
		return err
	}
	defer func() { _ = logger.Sync() }()

	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.GET("/healthz", healthCheck)

	logger.Info("starting api",
		zap.String("env", cfg.Env),
		zap.String("address", cfg.Address()),
		zap.String("log_level", cfg.LogLevel))

	// Address is 0.0.0.0:$PORT, not localhost: a container-local listener is
	// unreachable from outside the container.
	return router.Run(cfg.Address())
}

func healthCheck(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}
