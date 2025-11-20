package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/3dprint-hub/api/internal/order"
	httpmw "github.com/3dprint-hub/api/internal/http/middleware"
)

type checkoutRequest struct {
	Notes string `json:"notes"`
}

func (h *Handler) Checkout(w http.ResponseWriter, r *http.Request) {
	user, ok := httpmw.GetUser(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "login required")
		return
	}
	var req checkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err.Error() != "EOF" {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	order, err := h.App.Orders.Checkout(r.Context(), user.UserID, order.CheckoutInput{Notes: req.Notes})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, order)
}

func (h *Handler) ListOrders(w http.ResponseWriter, r *http.Request) {
	user, ok := httpmw.GetUser(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "login required")
		return
	}
	orders, err := h.App.Orders.ListByUser(r.Context(), user.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, orders)
}

func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	user, ok := httpmw.GetUser(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "login required")
		return
	}
	orderID, err := uuid.Parse(chi.URLParam(r, "orderID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid order id")
		return
	}
	order, err := h.App.Orders.Get(r.Context(), user.UserID, orderID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, order)
}
