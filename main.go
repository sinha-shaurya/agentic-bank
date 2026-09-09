package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/sinha-shaurya/agentic-bank/internal/config"
	"github.com/sinha-shaurya/agentic-bank/internal/obs"
	"github.com/sinha-shaurya/agentic-bank/internal/server"
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

	logger.Info("starting api",
		zap.String("env", cfg.Env),
		zap.String("address", cfg.Address()),
		zap.String("log_level", cfg.LogLevel))

	// SIGTERM is what a container runtime sends before it kills the process;
	// SIGINT is Ctrl-C locally. Both cancel the context and drain the server.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := server.New(cfg, logger, server.Health{})

	return srv.Run(ctx)
}
