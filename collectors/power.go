package collectors

import (
	"log"

	"github.com/hemtjanst/bibliotek/server"
	"github.com/prometheus/client_golang/prometheus"
)

// PowerCollector gets power data from sensors
type PowerCollector struct {
	powerCurrent *prometheus.Desc
	powerTotal   *prometheus.Desc
	m            *server.Manager
}

// NewPowerCollector returns a collector fetching power sensor data
func NewPowerCollector(m *server.Manager) (prometheus.Collector, error) {
	return &PowerCollector{
		m: m,
		powerCurrent: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "power", "current_watts"),
			"Current power draw in Watts",
			[]string{"source"}, nil,
		),
		powerTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "power", "total_kwh"),
			"Total power usage in kWh",
			[]string{"source"}, nil,
		),
	}, nil
}

// Describe sends all metrics descriptions into the channel
func (c *PowerCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.powerCurrent
	ch <- c.powerTotal
}

// Collect sends metric updates into the channel
func (c *PowerCollector) Collect(ch chan<- prometheus.Metric) {
	sensors := c.m.DeviceByType("outlet")
	for _, s := range sensors {
		if s.Feature("currentPower").Exists() {
			v, err := toFloat(s.Feature("currentPower").Value())
			if err != nil {
				log.Print(err.Error())
				continue
			}
			ch <- prometheus.MustNewConstMetric(c.powerCurrent,
				prometheus.GaugeValue, v, s.Info().Topic)
		}
		if s.Feature("energyUsed").Exists() {
			v, err := toFloat(s.Feature("energyUsed").Value())
			if err != nil {
				log.Print(err.Error())
				continue
			}
			ch <- prometheus.MustNewConstMetric(c.powerTotal,
				prometheus.CounterValue, v, s.Info().Topic)
		}
	}
}
