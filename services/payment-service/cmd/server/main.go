package main

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"payment-service/internal/handler"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	failureRate := 0.1 // 10% of charges are declined by default
	if v := os.Getenv("FAILURE_RATE"); v != "" {
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			failureRate = parsed
		}
	}

	h := handler.NewPaymentHandler(log, failureRate)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handler.Healthz)
	mux.HandleFunc("POST /payments", h.Charge)

	log.Info("payment-service starting", "port", port, "failure_rate", failureRate)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
