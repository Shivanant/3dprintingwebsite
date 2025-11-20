package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type updateOrderStatusRequest struct {
	Status string `json:"status"`
}

func (h *Handler) AdminListOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.App.Orders.AdminList(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, orders)
}

func (h *Handler) AdminUpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	orderID, err := uuid.Parse(chi.URLParam(r, "orderID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid order id")
		return
	}
	var req updateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	if req.Status == "" {
		writeError(w, http.StatusBadRequest, "status required")
		return
	}
	if err := h.App.Orders.UpdateStatus(r.Context(), orderID, req.Status); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
