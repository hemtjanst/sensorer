package collectors

import (
	"log"

	"github.com/hemtjanst/bibliotek/server"
	"github.com/prometheus/client_golang/prometheus"
)

// ContactCollector gets contact state from sensors
type ContactCollector struct {
	contactState *prometheus.Desc
	m            *server.Manager
}

// NewContactCollector returns a collector fetching contact sensor data
func NewContactCollector(m *server.Manager) (prometheus.Collector, error) {
	return &ContactCollector{
		m: m,
		contactState: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "contact", "state"),
			"Contact state (open/closed)",
			[]string{"source"}, nil,
		),
	}, nil
}

// Describe sends all metrics descriptions into the channel
func (c *ContactCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.contactState
}

// Collect sends metric updates into the channel
func (c *ContactCollector) Collect(ch chan<- prometheus.Metric) {
	sensors := c.m.DeviceByType("contactSensor")
	for _, s := range sensors {
		if s.Feature("contactSensorState").Exists() {
			v, err := toFloat(s.Feature("contactSensorState").Value())
			if err != nil {
				log.Print(err.Error())
				continue
			}
			ch <- prometheus.MustNewConstMetric(c.contactState,
				prometheus.GaugeValue, v, s.Info().Topic)
		}
	}
}
