package token

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"log/slog"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	secret           []byte
	accessTTL        time.Duration
	refreshTTL       time.Duration
	refreshTokenSize int
	logger           *slog.Logger
}

type Claims struct {
	UserID uuid.UUID `json:"uid"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

func New(secret string, accessTTL, refreshTTL time.Duration, refreshTokenSize int, logger *slog.Logger) *Service {
	return &Service{
		secret:           []byte(secret),
		accessTTL:        accessTTL,
		refreshTTL:       refreshTTL,
		refreshTokenSize: refreshTokenSize,
		logger:           logger,
	}
}

func (s *Service) GenerateAccessToken(userID uuid.UUID, role string) (token string, expiresAt time.Time, err error) {
	now := time.Now().UTC()
	expiresAt = now.Add(s.accessTTL)
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	j := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err = j.SignedString(s.secret)
	return
}

func (s *Service) ParseAccessToken(token string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := parsed.Claims.(*Claims); ok && parsed.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token claims")
}

func (s *Service) GenerateRefreshToken() (plain string, hash string, expiresAt time.Time, err error) {
	buf := make([]byte, s.refreshTokenSize)
	if _, err = rand.Read(buf); err != nil {
		return
	}
	plain = base64.RawURLEncoding.EncodeToString(buf)
	expiresAt = time.Now().UTC().Add(s.refreshTTL)
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", "", time.Time{}, err
	}
	hash = string(hashed)
	return
}

func (s *Service) VerifyRefreshToken(plain, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}
