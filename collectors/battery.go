package collectors

import (
	"log"

	"github.com/hemtjanst/bibliotek/server"
	"github.com/prometheus/client_golang/prometheus"
)

// BatteryCollector gets battery status from sensors
type BatteryCollector struct {
	batteryLevel *prometheus.Desc
	m            *server.Manager
}

// NewBatteryCollector returns a collector fetching battery data of sensors
func NewBatteryCollector(m *server.Manager) (prometheus.Collector, error) {
	return &BatteryCollector{
		m: m,
		batteryLevel: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "battery", "level_percent"),
			"Battery level in percent",
			[]string{"source"}, nil,
		),
	}, nil
}

// Describe sends all metrics descriptions into the channel
func (c *BatteryCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.batteryLevel
}

// Collect sends metric updates into the channel
func (c *BatteryCollector) Collect(ch chan<- prometheus.Metric) {
	sensors := c.m.DeviceByType("contactSensor")
	for _, s := range sensors {
		if s.Feature("batteryLevel").Exists() {
			v, err := toFloat(s.Feature("batteryLevel").Value())
			if err != nil {
				log.Print(err.Error())
				continue
			}
			ch <- prometheus.MustNewConstMetric(c.batteryLevel,
				prometheus.GaugeValue, v, s.Info().Topic)
		}
	}
}
