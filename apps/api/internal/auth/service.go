package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"log/slog"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/3dprint-hub/api/internal/database"
	"github.com/3dprint-hub/api/internal/mailer"
	"github.com/3dprint-hub/api/internal/oauth"
	"github.com/3dprint-hub/api/internal/pricing"
	"github.com/3dprint-hub/api/internal/storage"
	"github.com/3dprint-hub/api/internal/token"
)

type Options struct {
	DB         *gorm.DB
	Logger     *slog.Logger
	TokenSvc   *token.Service
	Mailer     mailer.Mailer
	OAuth      *oauth.Manager
	Pricing    *pricing.Service
	Storage    storage.Provider
	SignerIDFn func() uuid.UUID
}

type Service struct {
	db       *gorm.DB
	logger   *slog.Logger
	tokens   *token.Service
	mailer   mailer.Mailer
	oauth    *oauth.Manager
	pricing  *pricing.Service
	storage  storage.Provider
	signerID func() uuid.UUID
}

type AuthResult struct {
	User             *database.User
	AccessToken      string
	RefreshToken     string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
}

type LoginMetadata struct {
	IP        string
	UserAgent string
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrTokenInvalid       = errors.New("token invalid")
)

func NewService(opts Options) *Service {
	if opts.SignerIDFn == nil {
		opts.SignerIDFn = uuid.New
	}
	return &Service{
		db:       opts.DB,
		logger:   opts.Logger,
		tokens:   opts.TokenSvc,
		mailer:   opts.Mailer,
		oauth:    opts.OAuth,
		pricing:  opts.Pricing,
		storage:  opts.Storage,
		signerID: opts.SignerIDFn,
	}
}

func (s *Service) Register(ctx context.Context, email, password, name string, meta LoginMetadata) (*AuthResult, error) {
	email = normalizeEmail(email)
	if email == "" || password == "" {
		return nil, errors.New("email and password required")
	}
	var existing database.User
	err := s.db.WithContext(ctx).Where("email = ?", email).First(&existing).Error
	if err == nil {
		return nil, errors.New("email already registered")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &database.User{
		Email:        email,
		PasswordHash: ptr(string(hash)),
		Name:         name,
	}
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		cart := database.Cart{
			UserID: user.ID,
		}
		if err := tx.Create(&cart).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	res, err := s.issueTokens(ctx, user, nil, meta)
	if err != nil {
		return nil, err
	}
	go s.safeSendWelcome(user.Email, user.Name)
	return res, nil
}

func (s *Service) Login(ctx context.Context, email, password string, meta LoginMetadata) (*AuthResult, error) {
	email = normalizeEmail(email)
	var user database.User
	if err := s.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if user.PasswordHash == nil {
		return nil, errors.New("account uses social login")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	now := time.Now()
	s.db.Model(&user).Update("last_login_at", &now)
	return s.issueTokens(ctx, &user, nil, meta)
}

func (s *Service) Refresh(ctx context.Context, userID uuid.UUID, refreshToken string, meta LoginMetadata) (*AuthResult, error) {
	var tokenModel database.RefreshToken
	err := s.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at desc").First(&tokenModel).Error
	if err != nil {
		return nil, ErrTokenInvalid
	}
	if tokenModel.RevokedAt != nil || time.Now().After(tokenModel.ExpiresAt) {
		return nil, ErrTokenInvalid
	}
	if err := s.tokens.VerifyRefreshToken(refreshToken, tokenModel.TokenHash); err != nil {
		return nil, ErrTokenInvalid
	}
	var user database.User
	if err := s.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	return s.issueTokens(ctx, &user, &tokenModel.ID, meta)
}

func (s *Service) ForgotPassword(ctx context.Context, email string) error {
	email = normalizeEmail(email)
	var user database.User
	if err := s.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	token := uuid.NewString()
	reset := database.PasswordReset{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}
	if err := s.db.WithContext(ctx).Create(&reset).Error; err != nil {
		return err
	}
	return s.mailer.SendPasswordReset(ctx, user.Email, token)
}

func (s *Service) ResetPassword(ctx context.Context, token, newPassword string) (*AuthResult, error) {
	var reset database.PasswordReset
	if err := s.db.WithContext(ctx).
		Where("token = ?", token).
		Where("used_at IS NULL").
		First(&reset).Error; err != nil {
		return nil, ErrTokenInvalid
	}
	if time.Now().After(reset.ExpiresAt) {
		return nil, ErrTokenInvalid
	}
	var user database.User
	if err := s.db.WithContext(ctx).Where("id = ?", reset.UserID).First(&user).Error; err != nil {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&user).Update("password_hash", string(hash)).Error; err != nil {
			return err
		}
		if err := tx.Model(&reset).Updates(map[string]any{
			"used_at": now,
		}).Error; err != nil {
			return err
		}
		// revoke existing refresh tokens
		if err := tx.Model(&database.RefreshToken{}).Where("user_id = ?", user.ID).Update("revoked_at", now).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return s.issueTokens(ctx, &user, nil, LoginMetadata{})
}

func (s *Service) HandleOAuthCallback(ctx context.Context, provider, state, code string, meta LoginMetadata) (*AuthResult, error) {
	if s.oauth == nil {
		return nil, errors.New("oauth not configured")
	}
	token, profile, err := s.oauth.Exchange(ctx, provider, state, code)
	if err != nil {
		return nil, err
	}
	var account database.OAuthAccount
	err = s.db.WithContext(ctx).
		Where("provider = ? AND provider_user_id = ?", provider, profile.Subject).
		First(&account).Error
	var user database.User
	newUser := false
	switch {
	case err == nil:
		if err := s.db.WithContext(ctx).Where("id = ?", account.UserID).First(&user).Error; err != nil {
			return nil, err
		}
	case errors.Is(err, gorm.ErrRecordNotFound):
		// find user by email
		if err := s.db.WithContext(ctx).Where("email = ?", normalizeEmail(profile.Email)).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				newUser = true
				user = database.User{
					Email:       normalizeEmail(profile.Email),
					Name:        profile.Name,
					AvatarURL:   ptr(profile.AvatarURL),
					Role:        "user",
					PasswordHash: nil,
				}
				now := time.Now()
				user.EmailVerifiedAt = &now
				if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
					if err := tx.Create(&user).Error; err != nil {
						return err
					}
					cart := database.Cart{UserID: user.ID}
					if err := tx.Create(&cart).Error; err != nil {
						return err
					}
					return nil
				}); err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}
		}
	default:
		return nil, err
	}

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		account = database.OAuthAccount{
			UserID:         user.ID,
			Provider:       provider,
			ProviderUserID: profile.Subject,
			AccessToken:    token.AccessToken,
			RefreshToken:   token.RefreshToken,
		}
		return tx.Where("provider = ? AND provider_user_id = ?", provider, profile.Subject).
			Assign(account).
			FirstOrCreate(&account).Error
	}); err != nil {
		return nil, err
	}

	if profile.AvatarURL != "" && (user.AvatarURL == nil || *user.AvatarURL == "") {
		s.db.Model(&user).Update("avatar_url", profile.AvatarURL)
	}

	if newUser {
		go s.safeSendWelcome(user.Email, user.Name)
	}
	return s.issueTokens(ctx, &user, nil, meta)
}

func (s *Service) issueTokens(ctx context.Context, user *database.User, rotatedFrom *uuid.UUID, meta LoginMetadata) (*AuthResult, error) {
	access, accessExp, err := s.tokens.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, err
	}
	refreshPlain, refreshHash, refreshExp, err := s.tokens.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}
	refreshModel := database.RefreshToken{
		UserID:    user.ID,
		TokenHash: refreshHash,
		ExpiresAt: refreshExp,
		LastIP:    meta.IP,
		UserAgent: meta.UserAgent,
	}
	if rotatedFrom != nil {
		refreshModel.RotatedFromID = rotatedFrom
	}
	if err := s.db.WithContext(ctx).Create(&refreshModel).Error; err != nil {
		return nil, err
	}
	if rotatedFrom != nil {
		s.db.WithContext(ctx).Model(&database.RefreshToken{}).
			Where("id = ?", *rotatedFrom).
			Update("revoked_at", time.Now())
	}
	return &AuthResult{
		User:             user,
		AccessToken:      access,
		AccessExpiresAt:  accessExp,
		RefreshToken:     refreshPlain,
		RefreshExpiresAt: refreshExp,
	}, nil
}

func (s *Service) safeSendWelcome(email, name string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.mailer.SendWelcome(ctx, email, name); err != nil {
		s.logger.Warn("failed to send welcome email", "error", err)
	}
}

func normalizeEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}

func ptr[T any](v T) *T {
	return &v
}
