package collectors

import (
	"log"

	"lib.hemtjan.st/feature"

	"github.com/prometheus/client_golang/prometheus"
	"lib.hemtjan.st/server"
)

// FilterCollector gets filter status from sensors
type FilterCollector struct {
	filterReplacement *prometheus.Desc
	m                 *server.Manager
}

// NewFilterCollector returns a collector fetching filter data of sensors
func NewFilterCollector(m *server.Manager) (prometheus.Collector, error) {
	return &FilterCollector{
		m: m,
		filterReplacement: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "filter", "needs_replacement"),
			"Filter needs replacement",
			[]string{"source"}, nil,
		),
	}, nil
}

// Describe sends all metrics descriptions into the channel
func (c *FilterCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.filterReplacement
}

// Collect sends metric updates into the channel
func (c *FilterCollector) Collect(ch chan<- prometheus.Metric) {
	devices := c.m.Devices()
	for _, s := range devices {
		if s.Feature(feature.FilterChangeIndication.String()).Exists() {
			v, err := toFloat(s.Feature(feature.FilterChangeIndication.String()).Value())
			if err != nil {
				log.Print(err.Error())
				continue
			}
			ch <- prometheus.MustNewConstMetric(c.filterReplacement,
				prometheus.GaugeValue, v, s.Info().Topic)
		}
	}
}
