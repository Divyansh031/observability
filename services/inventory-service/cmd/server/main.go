package main

import (
	"log/slog"
	"net/http"
	"os"

	"inventory-service/internal/handler"
	"inventory-service/internal/store"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	invStore := store.NewInventoryStore()
	h := handler.NewInventoryHandler(invStore, log)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handler.Healthz)
	mux.HandleFunc("GET /inventory/{sku}", h.GetStock)
	mux.HandleFunc("POST /inventory/reserve", h.Reserve)

	log.Info("inventory-service starting", "port", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
