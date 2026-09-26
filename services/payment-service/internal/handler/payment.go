package handler

import (
	"encoding/json"
	"log/slog"
	"math/rand"
	"net/http"
	"time"

	"payment-service/internal/metrics"
)

type PaymentHandler struct {
	log     *slog.Logger
	metrics *metrics.PaymentMetrics
	// failureRate is the probability (0.0-1.0) that a charge is declined.
	// A knob now, an env var later — this is what we'll turn up on the
	// chaos-testing day to prove our alerts actually fire.
	failureRate float64
	// latencyMinMs/latencyMaxMs bound the simulated gateway delay. Also a
	// knob, for the same reason:a way to trigger the p99
	// latency alert on demand.
	latencyMinMs int
	latencyMaxMs int
}

func NewPaymentHandler(log *slog.Logger, m *metrics.PaymentMetrics, failureRate float64, latencyMinMs, latencyMaxMs int) *PaymentHandler {
	return &PaymentHandler{log: log, metrics: m, failureRate: failureRate, latencyMinMs: latencyMinMs, latencyMaxMs: latencyMaxMs}
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

	// Simulate a real payment gateway's variable latency.
	time.Sleep(time.Duration(h.latencyMinMs+rand.Intn(h.latencyMaxMs-h.latencyMinMs+1)) * time.Millisecond)

	if rand.Float64() < h.failureRate {
		h.metrics.TransactionsTotal.WithLabelValues("declined").Inc()
		h.log.Warn("payment declined", "order_id", req.OrderID)
		writeJSON(w, http.StatusPaymentRequired, errorResponse{Error: "payment declined"})
		return
	}

	h.metrics.TransactionsTotal.WithLabelValues("charged").Inc()
	writeJSON(w, http.StatusOK, chargeResponse{
		Status:        "charged",
		TransactionID: "txn_" + req.OrderID,
	})
}

func Healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
