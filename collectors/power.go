package collectors

import (
	"log"

	"github.com/prometheus/client_golang/prometheus"

	"lib.hemtjan.st/feature"
	"lib.hemtjan.st/server"
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
		if ft := s.Feature(feature.CurrentPower.String()); ft.Exists() {
			v := ft.Value()
			if v == "" {
				continue
			}
			vf, err := toFloat(v)
			if err != nil {
				log.Print(err.Error())
				continue
			}
			ch <- prometheus.MustNewConstMetric(c.powerCurrent,
				prometheus.GaugeValue, vf, s.Info().Topic)
		}
		if ft := s.Feature(feature.EnergyUsed.String()); ft.Exists() {
			v := ft.Value()
			if v == "" {
				continue
			}
			vf, err := toFloat(v)
			if err != nil {
				log.Print(err.Error())
				continue
			}
			ch <- prometheus.MustNewConstMetric(c.powerTotal,
				prometheus.CounterValue, vf, s.Info().Topic)
		}
		if ft := s.Feature(feature.CurrentVoltage.String()); ft.Exists() {
			v := ft.Value()
			if v == "" {
				continue
			}
			vf, err := toFloat(v)
			if err != nil {
				log.Print(err.Error())
				continue
			}
			ch <- prometheus.MustNewConstMetric(c.voltageCurrent,
				prometheus.GaugeValue, vf, s.Info().Topic)
		}
		if ft := s.Feature(feature.CurrentAmpere.String()); ft.Exists() {
			v := ft.Value()
			if v == "" {
				continue
			}
			vf, err := toFloat(v)
			if err != nil {
				log.Print(err.Error())
				continue
			}
			ch <- prometheus.MustNewConstMetric(c.ampereCurrent,
				prometheus.GaugeValue, vf, s.Info().Topic)
		}
	}
}
