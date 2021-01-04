package interceptor

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "bluekaki"
	subsystem = "vv"
)

// MetricsRequestCost  metrics for request cost
var MetricsRequestCost = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: namespace,
	Subsystem: subsystem,
	Name:      "requestcost",
	Help:      "request(s) cost seconds",
	Buckets:   []float64{0.1, 0.3, 0.5, 0.7, 0.9, 1.1},
}, []string{"method", "code"})

func init() {
	prometheus.MustRegister(MetricsRequestCost)
}
