package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/3dprint-hub/api/internal/token"
)

type contextKey string

const userKey contextKey = "authUser"

type UserContext struct {
	UserID uuid.UUID
	Role   string
}

func WithAuth(next http.Handler, tokens *token.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing Authorization header", http.StatusUnauthorized)
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			http.Error(w, "invalid Authorization header", http.StatusUnauthorized)
			return
		}
		claims, err := tokens.ParseAccessToken(parts[1])
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), userKey, UserContext{
			UserID: claims.UserID,
			Role:   claims.Role,
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUser(ctx context.Context) (UserContext, bool) {
	val := ctx.Value(userKey)
	if val == nil {
		return UserContext{}, false
	}
	user, ok := val.(UserContext)
	return user, ok
}

func RequireRole(role string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := GetUser(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if user.Role != role {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
