package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/3dprint-hub/api/internal/app"
	"github.com/3dprint-hub/api/internal/config"
	"github.com/3dprint-hub/api/internal/database"
)

func main() {
	ctx := context.Background()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

	db, err := database.New(cfg.Database.DSN)
	if err != nil {
		logger.Error("connect db", "error", err)
		os.Exit(1)
	}

	appInstance, err := app.New(ctx, cfg, logger, db)
	if err != nil {
		logger.Error("init app", "error", err)
		os.Exit(1)
	}

	if err := appInstance.Migrate(ctx); err != nil {
		logger.Error("migrate", "error", err)
		os.Exit(1)
	}
	logger.Info("migrations complete")
}
