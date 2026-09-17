package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"inventory-service/internal/metrics"
	"inventory-service/internal/store"
)

type InventoryHandler struct {
	store   *store.InventoryStore
	metrics *metrics.InventoryMetrics
	log     *slog.Logger
}

func NewInventoryHandler(s *store.InventoryStore, m *metrics.InventoryMetrics, log *slog.Logger) *InventoryHandler {
	return &InventoryHandler{store: s, metrics: m, log: log}
}

type stockResponse struct {
	SKU string `json:"sku"`
	Qty int    `json:"qty"`
}

type reserveRequest struct {
	SKU string `json:"sku"`
	Qty int    `json:"qty"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func (h *InventoryHandler) GetStock(w http.ResponseWriter, r *http.Request) {
	sku := r.PathValue("sku")

	qty, err := h.store.Get(sku)
	if err != nil {
		if errors.Is(err, store.ErrSKUNotFound) {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "sku not found"})
			return
		}
		h.log.Error("get stock failed", "sku", sku, "err", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
		return
	}

	writeJSON(w, http.StatusOK, stockResponse{SKU: sku, Qty: qty})
}

func (h *InventoryHandler) Reserve(w http.ResponseWriter, r *http.Request) {
	var req reserveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}
	if req.Qty <= 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "qty must be positive"})
		return
	}

	remaining, err := h.store.Reserve(req.SKU, req.Qty)
	switch {
	case err == nil:
		h.metrics.ReservationsTotal.WithLabelValues("reserved").Inc()
		h.metrics.StockLevel.WithLabelValues(req.SKU).Set(float64(remaining))
		writeJSON(w, http.StatusOK, map[string]string{"status": "reserved"})
	case errors.Is(err, store.ErrSKUNotFound):
		h.metrics.ReservationsTotal.WithLabelValues("not_found").Inc()
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "sku not found"})
	case errors.Is(err, store.ErrOutOfStock):
		h.metrics.ReservationsTotal.WithLabelValues("out_of_stock").Inc()
		writeJSON(w, http.StatusConflict, errorResponse{Error: "out of stock"})
	default:
		h.log.Error("reserve failed", "sku", req.SKU, "err", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
	}
}

func Healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
