package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// PaymentMetrics are business-level metrics: they describe domain outcomes
// (charged vs. declined), not just HTTP status codes like the generic RED
// metrics in middleware.go do.
type PaymentMetrics struct {
	// TransactionsTotal result values: charged | declined
	TransactionsTotal *prometheus.CounterVec
}

func NewPaymentMetrics() *PaymentMetrics {
	return &PaymentMetrics{
		TransactionsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "payment_transactions_total",
			Help: "Payment transactions processed, labeled by outcome.",
		}, []string{"result"}),
	}
}
