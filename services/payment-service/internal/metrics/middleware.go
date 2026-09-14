package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Middleware instruments HTTP handlers with the RED metrics: Rate (the
// counter below), Errors (the status label on that same counter), and
// Duration (the histogram).
//
// route MUST be the registered path pattern (e.g. "/inventory/{sku}"), never
// r.URL.Path. Labeling by the raw path would create a new time series per
// SKU ever requested — an unbounded-cardinality bug that can overload a real
// Prometheus server.
type Middleware struct {
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	inFlight        *prometheus.GaugeVec
}

func New() *Middleware {
	return &Middleware{
		requestsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests processed, labeled by route, method and status.",
		}, []string{"route", "method", "status"}),

		requestDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds, labeled by route and method.",
			Buckets: prometheus.DefBuckets,
		}, []string{"route", "method"}),

		inFlight: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Number of HTTP requests currently being processed, labeled by route.",
		}, []string{"route"}),
	}
}

// statusRecorder captures the status code a handler writes, since the
// stdlib http.ResponseWriter doesn't expose it after the fact.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// Wrap instruments next under the given route label.
func (m *Middleware) Wrap(route string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m.inFlight.WithLabelValues(route).Inc()
		defer m.inFlight.WithLabelValues(route).Dec()

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()

		next(rec, r)

		m.requestDuration.WithLabelValues(route, r.Method).Observe(time.Since(start).Seconds())
		m.requestsTotal.WithLabelValues(route, r.Method, strconv.Itoa(rec.status)).Inc()
	}
}
