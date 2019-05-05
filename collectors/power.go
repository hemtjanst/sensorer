package collectors

import (
	"log"

	"lib.hemtjan.st/feature"

	"lib.hemtjan.st/server"
	"github.com/prometheus/client_golang/prometheus"
)

// PowerCollector gets power data from sensors
type PowerCollector struct {
	powerCurrent   *prometheus.Desc
	powerTotal     *prometheus.Desc
	voltageCurrent *prometheus.Desc
	ampereCurrent  *prometheus.Desc
	m              *server.Manager
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
		voltageCurrent: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "power", "current_voltage"),
			"Current power draw in Volts",
			[]string{"source"}, nil,
		),
		ampereCurrent: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "power", "current_ampere"),
			"Current power draw in Amperes",
			[]string{"source"}, nil,
		),
	}, nil
}

// Describe sends all metrics descriptions into the channel
func (c *PowerCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.powerCurrent
	ch <- c.powerTotal
	ch <- c.voltageCurrent
	ch <- c.ampereCurrent
}

// Collect sends metric updates into the channel
func (c *PowerCollector) Collect(ch chan<- prometheus.Metric) {
	devices := c.m.Devices()
	for _, s := range devices {
		if s.Feature(feature.CurrentPower.String()).Exists() {
			v, err := toFloat(s.Feature(feature.CurrentPower.String()).Value())
			if err != nil {
				log.Print(err.Error())
				continue
			}
			ch <- prometheus.MustNewConstMetric(c.powerCurrent,
				prometheus.GaugeValue, v, s.Info().Topic)
		}
		if s.Feature(feature.EnergyUsed.String()).Exists() {
			v, err := toFloat(s.Feature(feature.EnergyUsed.String()).Value())
			if err != nil {
				log.Print(err.Error())
				continue
			}
			ch <- prometheus.MustNewConstMetric(c.powerTotal,
				prometheus.CounterValue, v, s.Info().Topic)
		}
		if s.Feature(feature.CurrentVoltage.String()).Exists() {
			v, err := toFloat(s.Feature(feature.CurrentVoltage.String()).Value())
			if err != nil {
				log.Print(err.Error())
				continue
			}
			ch <- prometheus.MustNewConstMetric(c.voltageCurrent,
				prometheus.GaugeValue, v, s.Info().Topic)
		}
		if s.Feature(feature.CurrentAmpere.String()).Exists() {
			v, err := toFloat(s.Feature(feature.CurrentAmpere.String()).Value())
			if err != nil {
				log.Print(err.Error())
				continue
			}
			ch <- prometheus.MustNewConstMetric(c.ampereCurrent,
				prometheus.GaugeValue, v, s.Info().Topic)
		}
	}
}
