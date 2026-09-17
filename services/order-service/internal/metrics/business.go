package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// OrderMetrics are business-level metrics: they describe the final outcome
// of the whole order flow in domain terms, not just this service's own HTTP
// status code (which is all the generic RED metrics in middleware.go see).
type OrderMetrics struct {
	// OrdersTotal status values: confirmed | out_of_stock | sku_not_found |
	// payment_declined | inventory_unavailable | payment_unavailable
	OrdersTotal *prometheus.CounterVec
}

func NewOrderMetrics() *OrderMetrics {
	return &OrderMetrics{
		OrdersTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "orders_total",
			Help: "Orders processed, labeled by final outcome.",
		}, []string{"status"}),
	}
}
