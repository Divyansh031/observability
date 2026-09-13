package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"order-service/internal/client"
	"order-service/internal/store"
)

type OrderHandler struct {
	orders    *store.OrderStore
	inventory *client.InventoryClient
	payment   *client.PaymentClient
	log       *slog.Logger
}

func NewOrderHandler(orders *store.OrderStore, inv *client.InventoryClient, pay *client.PaymentClient, log *slog.Logger) *OrderHandler {
	return &OrderHandler{orders: orders, inventory: inv, payment: pay, log: log}
}

type createOrderRequest struct {
	SKU    string  `json:"sku"`
	Qty    int     `json:"qty"`
	Amount float64 `json:"amount"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

// Create runs the order flow: reserve stock, then charge payment.
//
// NOTE (day-1 known simplification): if payment fails after inventory has
// already been reserved, we do NOT roll the reservation back. A real system
// would need a compensating action (saga pattern) or a two-phase reserve/
// commit. We're calling this out deliberately rather than solving it now —
// it's a good example of "obviously missing" production behavior to fix
// later, not an oversight.
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}
	if req.Qty <= 0 || req.Amount <= 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "qty and amount must be positive"})
		return
	}

	if err := h.inventory.Reserve(req.SKU, req.Qty); err != nil {
		switch {
		case errors.Is(err, client.ErrOutOfStock):
			writeJSON(w, http.StatusConflict, errorResponse{Error: "out of stock"})
		case errors.Is(err, client.ErrSKUNotFound):
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "sku not found"})
		default:
			h.log.Error("inventory reservation failed", "sku", req.SKU, "err", err)
			writeJSON(w, http.StatusBadGateway, errorResponse{Error: "inventory service unavailable"})
		}
		return
	}

	order := h.orders.Create(store.Order{
		SKU:    req.SKU,
		Qty:    req.Qty,
		Amount: req.Amount,
		Status: "pending",
	})

	txnID, err := h.payment.Charge(order.ID, req.Amount)
	if err != nil {
		if errors.Is(err, client.ErrPaymentDeclined) {
			writeJSON(w, http.StatusPaymentRequired, errorResponse{Error: "payment declined"})
			return
		}
		h.log.Error("payment charge failed", "order_id", order.ID, "err", err)
		writeJSON(w, http.StatusBadGateway, errorResponse{Error: "payment service unavailable"})
		return
	}

	order.Status = "confirmed"
	order.TransactionID = txnID
	h.orders.Update(order)

	writeJSON(w, http.StatusCreated, order)
}

func (h *OrderHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	order, err := h.orders.Get(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "order not found"})
		return
	}
	writeJSON(w, http.StatusOK, order)
}

func Healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
