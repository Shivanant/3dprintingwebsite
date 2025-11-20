package app

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/3dprint-hub/api/internal/auth"
	"github.com/3dprint-hub/api/internal/cart"
	"github.com/3dprint-hub/api/internal/config"
	"github.com/3dprint-hub/api/internal/database"
	"github.com/3dprint-hub/api/internal/jobs"
	"github.com/3dprint-hub/api/internal/mailer"
	"github.com/3dprint-hub/api/internal/oauth"
	"github.com/3dprint-hub/api/internal/order"
	"github.com/3dprint-hub/api/internal/pricing"
	"github.com/3dprint-hub/api/internal/storage"
	"github.com/3dprint-hub/api/internal/token"
)

type Application struct {
	Config  *config.Config
	Logger  *slog.Logger
	DB      *gorm.DB
	Tokens  *token.Service
	Auth    *auth.Service
	Mailer  mailer.Mailer
	OAuth   *oauth.Manager
	Pricing *pricing.Service
	Storage storage.Provider
	Cart    *cart.Service
	Orders  *order.Service
	Jobs    *jobs.Service
}

func New(ctx context.Context, cfg *config.Config, logger *slog.Logger, db *gorm.DB) (*Application, error) {
	mailerSvc := mailer.New(cfg, logger)
	storageProvider, err := storage.NewLocal(cfg.Storage.UploadsPath)
	if err != nil {
		return nil, err
	}

	tokenSvc := token.New(cfg.JWT.Secret, cfg.JWT.AccessTokenTTL, cfg.JWT.RefreshTokenTTL, cfg.JWT.RefreshTokenSize, logger)

	pricingSvc := pricing.NewService(pricing.Options{
		MaterialCostPLA: cfg.Pricing.MaterialCostPLA,
		MachineRate:     cfg.Pricing.MachineRate,
		SetupFee:        cfg.Pricing.SetupFee,
		PrintSpeed:      cfg.Pricing.PrintSpeed,
		Logger:          logger,
		StoragePath:     cfg.Storage.UploadsPath,
	})

	oauthMgr := oauth.NewManager(cfg, logger)

	cartSvc := cart.New(db, logger)
	orderSvc := order.New(db, logger)
	jobSvc := jobs.New(db, logger, storageProvider)

	authSvc := auth.NewService(auth.Options{
		DB:         db,
		Logger:     logger,
		TokenSvc:   tokenSvc,
		Mailer:     mailerSvc,
		OAuth:      oauthMgr,
		Pricing:    pricingSvc,
		Storage:    storageProvider,
		SignerIDFn: func() uuid.UUID { return uuid.New() },
	})

	return &Application{
		Config:  cfg,
		Logger:  logger,
		DB:      db,
		Tokens:  tokenSvc,
		Auth:    authSvc,
		Mailer:  mailerSvc,
		OAuth:   oauthMgr,
		Pricing: pricingSvc,
		Storage: storageProvider,
		Cart:    cartSvc,
		Orders:  orderSvc,
		Jobs:    jobSvc,
	}, nil
}

func (a *Application) Migrate(ctx context.Context) error {
	return database.Migrate(ctx, a.DB)
}
