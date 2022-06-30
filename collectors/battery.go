package collectors

import (
	"log"

	"github.com/prometheus/client_golang/prometheus"

	"lib.hemtjan.st/feature"
	"lib.hemtjan.st/server"
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
	devices := c.m.Devices()
	for _, s := range devices {
		if ft := s.Feature(feature.BatteryLevel.String()); ft.Exists() {
			v := ft.Value()
			if v == "" {
				continue
			}
			vf, err := toFloat(v)
			if err != nil {
				log.Print(err.Error())
				continue
			}
			ch <- prometheus.MustNewConstMetric(c.batteryLevel,
				prometheus.GaugeValue, vf, s.Info().Topic)
		}
	}
}
