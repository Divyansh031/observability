// alert-receiver is a throwaway debug tool, not a real notification
// channel: it exists only to prove Alertmanager's routing config actually
// delivers somewhere, by logging whatever payload it receives. It is
// deliberately NOT instrumented with Prometheus metrics -- it's test
// plumbing, not part of the observability target.
package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
)

// Matches Alertmanager's webhook_config payload shape:
// https://prometheus.io/docs/alerting/latest/configuration/#webhook_config
type webhookPayload struct {
	Status string `json:"status"`
	Alerts []struct {
		Status      string            `json:"status"`
		Labels      map[string]string `json:"labels"`
		Annotations map[string]string `json:"annotations"`
	} `json:"alerts"`
}

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error("failed to read webhook body", "err", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var payload webhookPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			log.Warn("received non-Alertmanager payload", "raw", string(body))
			w.WriteHeader(http.StatusOK)
			return
		}

		for _, a := range payload.Alerts {
			log.Info("ALERT RECEIVED",
				"alertname", a.Labels["alertname"],
				"status", a.Status,
				"severity", a.Labels["severity"],
				"job", a.Labels["job"],
				"deployment", a.Labels["deployment"],
				"summary", a.Annotations["summary"],
			)
		}

		w.WriteHeader(http.StatusOK)
	})

	log.Info("alert-receiver starting", "port", "8090")
	if err := http.ListenAndServe(":8090", nil); err != nil {
		log.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
