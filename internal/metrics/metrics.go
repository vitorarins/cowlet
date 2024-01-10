package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	once   sync.Once
	metric *metrics
)

type Metrics interface {
	OxygenSaturationSet(int)
}

type metrics struct {
	ox prometheus.Gauge
}

func New(reg prometheus.Registerer) Metrics {
	// Prometheus library panics when try to register the same metrics twice
	// for preventing that to happen we made metrics a singleton using once
	once.Do(func() {
		metric = &metrics{
			ox: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "oxygen_saturation_percent",
				Help: "Current Reading Oxygen Saturation.",
			}),
		}

		reg.MustRegister(metric.ox)
	})

	return metric
}

func (m *metrics) OxygenSaturationSet(oxPercent int) {
	m.ox.Set(float64(oxPercent))
}
