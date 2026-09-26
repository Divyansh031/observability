package main

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"payment-service/internal/handler"
	"payment-service/internal/metrics"
)

func envInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	parsed, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return parsed
}

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

	latencyMinMs := envInt("LATENCY_MIN_MS", 20)
	latencyMaxMs := envInt("LATENCY_MAX_MS", 120)

	bizMetrics := metrics.NewPaymentMetrics()
	h := handler.NewPaymentHandler(log, bizMetrics, failureRate, latencyMinMs, latencyMaxMs)
	mw := metrics.New()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handler.Healthz)
	mux.Handle("GET /metrics", promhttp.Handler())
	mux.HandleFunc("POST /payments", mw.Wrap("/payments", h.Charge))

	log.Info("payment-service starting", "port", port, "failure_rate", failureRate)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
