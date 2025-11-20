package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/3dprint-hub/api/internal/cart"
	httpmw "github.com/3dprint-hub/api/internal/http/middleware"
)

type addCartItemRequest struct {
	SKU            string         `json:"sku"`
	DisplayName    string         `json:"displayName"`
	Quantity       int            `json:"quantity"`
	UnitPriceCents int            `json:"unitPriceCents"`
	Metadata       map[string]any `json:"metadata"`
}

func (h *Handler) GetCart(w http.ResponseWriter, r *http.Request) {
	user, ok := httpmw.GetUser(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "login required")
		return
	}
	cartDTO, err := h.App.Cart.GetByUser(r.Context(), user.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, cartDTO)
}

func (h *Handler) AddCartItem(w http.ResponseWriter, r *http.Request) {
	user, ok := httpmw.GetUser(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "login required")
		return
	}
	var req addCartItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	cartDTO, err := h.App.Cart.AddItem(r.Context(), user.UserID, cart.ItemInput{
		SKU:            req.SKU,
		DisplayName:    req.DisplayName,
		Quantity:       req.Quantity,
		UnitPriceCents: req.UnitPriceCents,
		Metadata:       req.Metadata,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, cartDTO)
}

func (h *Handler) RemoveCartItem(w http.ResponseWriter, r *http.Request) {
	user, ok := httpmw.GetUser(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "login required")
		return
	}
	itemID, err := uuid.Parse(chi.URLParam(r, "itemID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid item id")
		return
	}
	cartDTO, err := h.App.Cart.RemoveItem(r.Context(), user.UserID, itemID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, cartDTO)
}
