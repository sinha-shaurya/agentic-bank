package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	authRoutes "github.com/sinha-shaurya/agentic-bank/auth/routes"
	"github.com/sinha-shaurya/agentic-bank/internal/config"
	"github.com/sinha-shaurya/agentic-bank/internal/database"
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

	if err := database.New(cfg, logger); err != nil {
		return err
	}
	defer func() { _ = database.Close() }()

	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// perform auto migrate

	router := gin.Default()
	router.GET("/healthz", healthCheck)

	logger.Info("starting api",
		zap.String("env", cfg.Env),
		zap.String("address", cfg.Address()),
		zap.String("log_level", cfg.LogLevel))

	// attach api group
	apiGroup := router.Group("/api")
	attachRoutes(apiGroup)
	return router.Run(cfg.Address())
}

func healthCheck(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func attachRoutes(apiGroup *gin.RouterGroup) {
	authRoutes.AuthRoutes(apiGroup)
}
