package handler

import (
	"encoding/json"
	"log/slog"
	"math/rand"
	"net/http"
	"time"
)

type PaymentHandler struct {
	log *slog.Logger
	// failureRate is the probability (0.0-1.0) that a charge is declined.
	// A knob now, an env var later — this is what we'll turn up on the
	// chaos-testing day to prove our alerts actually fire.
	failureRate float64
}

func NewPaymentHandler(log *slog.Logger, failureRate float64) *PaymentHandler {
	return &PaymentHandler{log: log, failureRate: failureRate}
}

type chargeRequest struct {
	OrderID string  `json:"order_id"`
	Amount  float64 `json:"amount"`
}

type chargeResponse struct {
	Status        string `json:"status"`
	TransactionID string `json:"transaction_id,omitempty"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func (h *PaymentHandler) Charge(w http.ResponseWriter, r *http.Request) {
	var req chargeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}
	if req.Amount <= 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "amount must be positive"})
		return
	}

	// Simulate a real payment gateway's variable latency (20-120ms).
	time.Sleep(time.Duration(20+rand.Intn(100)) * time.Millisecond)

	if rand.Float64() < h.failureRate {
		h.log.Warn("payment declined", "order_id", req.OrderID)
		writeJSON(w, http.StatusPaymentRequired, errorResponse{Error: "payment declined"})
		return
	}

	writeJSON(w, http.StatusOK, chargeResponse{
		Status:        "charged",
		TransactionID: "txn_" + req.OrderID,
	})
}

func Healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
