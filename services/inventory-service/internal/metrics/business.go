package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// InventoryMetrics are business-level metrics: they describe domain outcomes
// (was stock reserved, declined because empty, etc.), not just HTTP status
// codes like the generic RED metrics in middleware.go do.
type InventoryMetrics struct {
	// ReservationsTotal result values: reserved | out_of_stock | not_found
	ReservationsTotal *prometheus.CounterVec

	// StockLevel is labeled by sku. This is safe cardinality-wise because
	// our SKU catalog is small and fixed (we define it, nobody else can add
	// to it) — very different from labeling by a user- or request-supplied
	// ID, which would be unbounded.
	StockLevel *prometheus.GaugeVec
}

func NewInventoryMetrics() *InventoryMetrics {
	return &InventoryMetrics{
		ReservationsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "inventory_reservations_total",
			Help: "Inventory reservation attempts, labeled by outcome.",
		}, []string{"result"}),

		StockLevel: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "inventory_stock_level",
			Help: "Current stock level per SKU.",
		}, []string{"sku"}),
	}
}
