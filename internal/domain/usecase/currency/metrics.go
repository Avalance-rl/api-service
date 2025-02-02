package currency

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	priceCounter   prometheus.Counter
	errorCounter   prometheus.Counter
	updateDuration prometheus.Histogram
}

func newMetrics() *Metrics {
	m := &Metrics{
		priceCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "crypto_price_updates_total",
			Help: "Total number of price updates processed",
		}),
		errorCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "crypto_price_errors_total",
			Help: "Total number of price update errors",
		}),
		updateDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "crypto_price_update_duration_seconds",
			Help:    "Time spent updating prices",
			Buckets: prometheus.DefBuckets,
		}),
	}

	prometheus.MustRegister(m.priceCounter)
	prometheus.MustRegister(m.errorCounter)
	prometheus.MustRegister(m.updateDuration)

	return m
}
