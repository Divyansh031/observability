package main

import (
	"log/slog"
	"net/http"
	"os"

	"order-service/internal/client"
	"order-service/internal/handler"
	"order-service/internal/store"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}

	inventoryURL := os.Getenv("INVENTORY_SERVICE_URL")
	if inventoryURL == "" {
		inventoryURL = "http://localhost:8082"
	}
	paymentURL := os.Getenv("PAYMENT_SERVICE_URL")
	if paymentURL == "" {
		paymentURL = "http://localhost:8083"
	}

	orderStore := store.NewOrderStore()
	inventoryClient := client.NewInventoryClient(inventoryURL)
	paymentClient := client.NewPaymentClient(paymentURL)
	h := handler.NewOrderHandler(orderStore, inventoryClient, paymentClient, log)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handler.Healthz)
	mux.HandleFunc("POST /orders", h.Create)
	mux.HandleFunc("GET /orders/{id}", h.Get)

	log.Info("order-service starting", "port", port, "inventory_url", inventoryURL, "payment_url", paymentURL)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
