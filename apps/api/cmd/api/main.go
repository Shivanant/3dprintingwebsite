package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/3dprint-hub/api/internal/app"
	"github.com/3dprint-hub/api/internal/config"
	"github.com/3dprint-hub/api/internal/database"
	apiserver "github.com/3dprint-hub/api/internal/http"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	db, err := database.New(cfg.Database.DSN)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}

	appInstance, err := app.New(ctx, cfg, logger, db)
	if err != nil {
		logger.Error("failed to initialize application", "error", err)
		os.Exit(1)
	}

	if err := appInstance.Migrate(ctx); err != nil {
		logger.Error("database migration failed", "error", err)
		os.Exit(1)
	}

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           apiserver.New(appInstance),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info("api server listening", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
			cancel()
		}
	}()

	<-ctx.Done()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown failed", "error", err)
	} else {
		logger.Info("server shutdown complete")
	}
}
