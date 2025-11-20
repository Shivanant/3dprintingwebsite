package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppEnv      string
	Port        int
	PublicURL   string
	FrontendURL string

	Database struct {
		DSN string
	}

	JWT struct {
		Secret           string
		AccessTokenTTL   time.Duration
		RefreshTokenTTL  time.Duration
		RefreshTokenSize int
	}

	Mailgun struct {
		Domain string
		APIKey string
		From   string
	}

	OAuth struct {
		Google OAuthProvider
		GitHub OAuthProvider
	}

	Storage struct {
		UploadsPath string
	}

	Pricing struct {
		MaterialCostPLA float64
		MachineRate     float64
		SetupFee        float64
		PrintSpeed      float64
	}
}

type OAuthProvider struct {
	ClientID     string
	ClientSecret string
	RedirectPath string
}

func Load() (*Config, error) {
	cfg := &Config{
		AppEnv:      getEnv("APP_ENV", "development"),
		PublicURL:   getEnv("PUBLIC_URL", "http://localhost:8080"),
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:3000"),
	}

	port, err := strconv.Atoi(getEnv("PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid PORT: %w", err)
	}
	cfg.Port = port

	cfg.Database.DSN = getEnv("POSTGRES_DSN", "postgres://postgres:postgres@localhost:5432/print_hub?sslmode=disable")

	cfg.JWT.Secret = getEnv("JWT_SECRET", "")
	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	accessTTL, err := time.ParseDuration(getEnv("JWT_ACCESS_TTL", "15m"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_TTL: %w", err)
	}
	refreshTTL, err := time.ParseDuration(getEnv("JWT_REFRESH_TTL", "336h")) // 14 days
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_TTL: %w", err)
	}
	cfg.JWT.AccessTokenTTL = accessTTL
	cfg.JWT.RefreshTokenTTL = refreshTTL
	refreshSize, err := strconv.Atoi(getEnv("JWT_REFRESH_BYTES", "32"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_BYTES: %w", err)
	}
	cfg.JWT.RefreshTokenSize = refreshSize

	cfg.Mailgun.Domain = getEnv("MAILGUN_DOMAIN", "")
	cfg.Mailgun.APIKey = getEnv("MAILGUN_API_KEY", "")
	cfg.Mailgun.From = getEnv("MAILGUN_FROM", "")

	cfg.OAuth.Google = OAuthProvider{
		ClientID:     os.Getenv("OAUTH_GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("OAUTH_GOOGLE_CLIENT_SECRET"),
		RedirectPath: getEnv("OAUTH_GOOGLE_REDIRECT_PATH", "/api/v1/auth/oauth/google/callback"),
	}
	cfg.OAuth.GitHub = OAuthProvider{
		ClientID:     os.Getenv("OAUTH_GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("OAUTH_GITHUB_CLIENT_SECRET"),
		RedirectPath: getEnv("OAUTH_GITHUB_REDIRECT_PATH", "/api/v1/auth/oauth/github/callback"),
	}

	cfg.Storage.UploadsPath = getEnv("STORAGE_UPLOADS_PATH", "storage/uploads")

	cfg.Pricing.MaterialCostPLA = parseFloat(getEnv("PRICING_MATERIAL_COST_PLA", "0.12"))
	cfg.Pricing.MachineRate = parseFloat(getEnv("PRICING_MACHINE_RATE", "12.5"))
	cfg.Pricing.SetupFee = parseFloat(getEnv("PRICING_SETUP_FEE", "4.5"))
	cfg.Pricing.PrintSpeed = parseFloat(getEnv("PRICING_PRINT_SPEED", "5500"))

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseFloat(v string) float64 {
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0
	}
	return f
}
