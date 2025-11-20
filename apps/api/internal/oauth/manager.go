package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"log/slog"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"

	"github.com/3dprint-hub/api/internal/config"
)

type Manager struct {
	logger    *slog.Logger
	config    *config.Config
	providers map[string]*oauth2.Config
	stateTTL  time.Duration
	states    sync.Map
}

type stateEntry struct {
	ExpiresAt time.Time
	Redirect  string
}

type Profile struct {
	Email     string
	Name      string
	AvatarURL string
	Provider  string
	Subject   string
}

func NewManager(cfg *config.Config, logger *slog.Logger) *Manager {
	providers := map[string]*oauth2.Config{}
	if cfg.OAuth.Google.ClientID != "" && cfg.OAuth.Google.ClientSecret != "" {
		providers["google"] = &oauth2.Config{
			ClientID:     cfg.OAuth.Google.ClientID,
			ClientSecret: cfg.OAuth.Google.ClientSecret,
			Endpoint:     google.Endpoint,
			RedirectURL:  cfg.PublicURL + cfg.OAuth.Google.RedirectPath,
			Scopes: []string{
				"openid", "profile", "email",
			},
		}
	}
	if cfg.OAuth.GitHub.ClientID != "" && cfg.OAuth.GitHub.ClientSecret != "" {
		providers["github"] = &oauth2.Config{
			ClientID:     cfg.OAuth.GitHub.ClientID,
			ClientSecret: cfg.OAuth.GitHub.ClientSecret,
			Endpoint:     github.Endpoint,
			RedirectURL:  cfg.PublicURL + cfg.OAuth.GitHub.RedirectPath,
			Scopes:       []string{"user:email"},
		}
	}
	return &Manager{
		logger:    logger,
		config:    cfg,
		providers: providers,
		stateTTL:  10 * time.Minute,
	}
}

func (m *Manager) Providers() []string {
	out := make([]string, 0, len(m.providers))
	for k := range m.providers {
		out = append(out, k)
	}
	return out
}

func (m *Manager) GenerateAuthURL(provider string, redirect string) (url string, state string, err error) {
	cfg, ok := m.providers[provider]
	if !ok {
		return "", "", fmt.Errorf("provider %s not configured", provider)
	}
	state = uuid.New().String()
	cfgCopy := *cfg
	if redirect != "" {
		cfgCopy.RedirectURL = redirect
	}
	url = cfgCopy.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	m.states.Store(state, stateEntry{ExpiresAt: time.Now().Add(m.stateTTL), Redirect: cfgCopy.RedirectURL})
	return url, state, nil
}

func (m *Manager) Exchange(ctx context.Context, provider, state, code string) (*oauth2.Token, Profile, error) {
	entryAny, ok := m.states.Load(state)
	if !ok {
		return nil, Profile{}, errors.New("invalid oauth state")
	}
	m.states.Delete(state)
	entry := entryAny.(stateEntry)
	if time.Now().After(entry.ExpiresAt) {
		return nil, Profile{}, errors.New("oauth state expired")
	}
	cfg, ok := m.providers[provider]
	if !ok {
		return nil, Profile{}, fmt.Errorf("provider %s not configured", provider)
	}
	cfgCopy := *cfg
	if entry.Redirect != "" {
		cfgCopy.RedirectURL = entry.Redirect
	}
	token, err := cfgCopy.Exchange(ctx, code)
	if err != nil {
		return nil, Profile{}, err
	}
	profile, err := m.fetchProfile(ctx, provider, token)
	return token, profile, err
}

func (m *Manager) fetchProfile(ctx context.Context, provider string, token *oauth2.Token) (Profile, error) {
	var req *http.Request
	var err error
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	switch provider {
	case "google":
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v3/userinfo", nil)
	case "github":
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	default:
		return Profile{}, fmt.Errorf("unsupported provider %s", provider)
	}
	if err != nil {
		return Profile{}, err
	}
	res, err := client.Do(req)
	if err != nil {
		return Profile{}, err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		return Profile{}, fmt.Errorf("profile request failed: %s", res.Status)
	}
	if provider == "google" {
		var raw struct {
			Email string `json:"email"`
			Name  string `json:"name"`
			Pic   string `json:"picture"`
			Sub   string `json:"sub"`
		}
		if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
			return Profile{}, err
		}
		return Profile{
			Email:     raw.Email,
			Name:      raw.Name,
			AvatarURL: raw.Pic,
			Provider:  provider,
			Subject:   raw.Sub,
		}, nil
	}
	if provider == "github" {
		var raw struct {
			ID     int64  `json:"id"`
			Login  string `json:"login"`
			Name   string `json:"name"`
			Email  string `json:"email"`
			Avatar string `json:"avatar_url"`
		}
		if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
			return Profile{}, err
		}
		email := raw.Email
		if email == "" {
			email, err = m.fetchGitHubEmail(ctx, token)
			if err != nil {
				return Profile{}, err
			}
		}
		name := raw.Name
		if name == "" {
			name = raw.Login
		}
		return Profile{
			Email:     email,
			Name:      name,
			AvatarURL: raw.Avatar,
			Provider:  provider,
			Subject:   fmt.Sprintf("%d", raw.ID),
		}, nil
	}
	return Profile{}, errors.New("unhandled provider")
}

func (m *Manager) fetchGitHubEmail(ctx context.Context, token *oauth2.Token) (string, error) {
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		return "", fmt.Errorf("github email request failed: %s", res.Status)
	}
	var list []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.NewDecoder(res.Body).Decode(&list); err != nil {
		return "", err
	}
	for _, item := range list {
		if item.Primary && item.Verified {
			return item.Email, nil
		}
	}
	if len(list) > 0 {
		return list[0].Email, nil
	}
	return "", errors.New("no github email found")
}
