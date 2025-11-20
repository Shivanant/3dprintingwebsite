package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/3dprint-hub/api/internal/app"
)

type Handler struct {
	App *app.Application
}

func New(app *app.Application) *Handler {
	return &Handler{App: app}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{"error": message})
}
