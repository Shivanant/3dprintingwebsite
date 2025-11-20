package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/3dprint-hub/api/internal/auth"
	"github.com/3dprint-hub/api/internal/database"
	httpmw "github.com/3dprint-hub/api/internal/http/middleware"
)

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	UserID       string `json:"userId"`
	RefreshToken string `json:"refreshToken"`
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

type resetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"newPassword"`
}

type oauthStartResponse struct {
	URL   string `json:"url"`
	State string `json:"state"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	meta := h.loginMeta(r)
	res, err := h.App.Auth.Register(r.Context(), req.Email, req.Password, req.Name, meta)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, h.authResponse(res))
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	meta := h.loginMeta(r)
	res, err := h.App.Auth.Login(r.Context(), req.Email, req.Password, meta)
	switch err {
	case nil:
	case auth.ErrInvalidCredentials:
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	default:
		writeError(w, http.StatusInternalServerError, "login failed")
		return
	}
	writeJSON(w, http.StatusOK, h.authResponse(res))
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	meta := h.loginMeta(r)
	res, err := h.App.Auth.Refresh(r.Context(), userID, req.RefreshToken, meta)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "refresh failed")
		return
	}
	writeJSON(w, http.StatusOK, h.authResponse(res))
}

func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req forgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	if err := h.App.Auth.ForgotPassword(r.Context(), req.Email); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to send email")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	res, err := h.App.Auth.ResetPassword(r.Context(), req.Token, req.NewPassword)
	if err != nil {
		writeError(w, http.StatusBadRequest, "reset failed")
		return
	}
	writeJSON(w, http.StatusOK, h.authResponse(res))
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := httpmw.GetUser(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var user database.User
	if err := h.App.DB.WithContext(r.Context()).Where("id = ?", userCtx.UserID).First(&user).Error; err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	writeJSON(w, http.StatusOK, sanitizeUser(user))
}

func (h *Handler) OAuthStart(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	redirect := r.URL.Query().Get("redirect_uri")
	url, state, err := h.App.OAuth.GenerateAuthURL(provider, redirect)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, oauthStartResponse{URL: url, State: state})
}

func (h *Handler) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")
	if state == "" || code == "" {
		writeError(w, http.StatusBadRequest, "missing state or code")
		return
	}
	meta := h.loginMeta(r)
	res, err := h.App.Auth.HandleOAuthCallback(r.Context(), provider, state, code, meta)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, h.authResponse(res))
}

func (h *Handler) loginMeta(r *http.Request) auth.LoginMetadata {
	return auth.LoginMetadata{
		IP:        clientIP(r),
		UserAgent: r.UserAgent(),
	}
}

func (h *Handler) authResponse(res *auth.AuthResult) map[string]any {
	return map[string]any{
		"user":             sanitizeUser(*res.User),
		"accessToken":      res.AccessToken,
		"accessExpiresAt":  res.AccessExpiresAt,
		"refreshToken":     res.RefreshToken,
		"refreshExpiresAt": res.RefreshExpiresAt,
	}
}

func sanitizeUser(user database.User) map[string]any {
	return map[string]any{
		"id":        user.ID,
		"email":     user.Email,
		"name":      user.Name,
		"role":      user.Role,
		"avatarUrl": valueOrNil(user.AvatarURL),
	}
}

func valueOrNil[T any](p *T) any {
	if p == nil {
		return nil
	}
	return *p
}

func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		parts := strings.Split(fwd, ",")
		return strings.TrimSpace(parts[0])
	}
	return strings.Split(r.RemoteAddr, ":")[0]
}
