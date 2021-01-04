package server

import (
	"net/http"

	"github.com/bluekaki/vv/internal/interceptor"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/push"
)

// WithPrometheus prometheus metrics exposes on http://addr/metrics
func WithPrometheus(addr string) Option {
	return func(opt *option) {
		opt.prometheusHandler = func() {
			http.Handle("/metrics", promhttp.Handler())
			go func() {
				if err := http.ListenAndServe(addr, nil); err != nil {
					panic(err)
				}
			}()
		}
	}
}

// WithPrometheusPush  push prometheus metrics to the Pushgateway
func WithPrometheusPush(gateway string) Option {
	return func(opt *option) {
		opt.prometheusHandler = func() {
			if err := push.New(gateway, "bluekaiki_vv_metrics").
				Collector(interceptor.MetricsRequestCost).
				Push(); err != nil {
				panic(err)
			}
		}
	}
}
