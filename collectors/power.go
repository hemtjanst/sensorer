package collectors

import (
	"fmt"
	"log"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"

	"lib.hemtjan.st/feature"
	"lib.hemtjan.st/server"
)

// PowerCollector gets power data from sensors
type PowerCollector struct {
	powerCurrent         *prometheus.Desc
	powerProducedCurrent *prometheus.Desc
	powerTotal           *prometheus.Desc
	powerProducedTotal   *prometheus.Desc
	voltageCurrent       *prometheus.Desc
	voltageCurrentPhase  *prometheus.Desc
	ampereCurrent        *prometheus.Desc
	ampereCurrentPhase   *prometheus.Desc
	m                    *server.Manager
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
		powerProducedCurrent: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "power", "produced_current_watts"),
			"Current power production in Watts",
			[]string{"source"}, nil,
		),
		powerTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "power", "total_kwh"),
			"Total power usage in kWh",
			[]string{"source"}, nil,
		),
		powerProducedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "power", "produced_total_kwh"),
			"Total power production in kWh",
			[]string{"source"}, nil,
		),
		voltageCurrent: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "power", "current_voltage"),
			"Current voltage",
			[]string{"source"}, nil,
		),
		ampereCurrent: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "power", "current_ampere"),
			"Current power draw in Amperes",
			[]string{"source"}, nil,
		),
		voltageCurrentPhase: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "power", "current_voltage"),
			"Current voltage",
			[]string{"source", "phase"}, nil,
		),
		ampereCurrentPhase: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "power", "current_ampere"),
			"Current power draw in Amperes",
			[]string{"source", "phase"}, nil,
		),
	}, nil
}

// Describe sends all metrics descriptions into the channel
func (c *PowerCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.powerCurrent
	ch <- c.powerProducedCurrent
	ch <- c.powerTotal
	ch <- c.powerProducedTotal
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
		if ft := s.Feature("currentPowerProduced"); ft.Exists() {
			v := ft.Value()
			if v == "" {
				continue
			}
			vf, err := toFloat(v)
			if err != nil {
				log.Print(err.Error())
				continue
			}
			ch <- prometheus.MustNewConstMetric(c.powerProducedCurrent,
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
		if ft := s.Feature("energyProduced"); ft.Exists() {
			v := ft.Value()
			if v == "" {
				continue
			}
			vf, err := toFloat(v)
			if err != nil {
				log.Print(err.Error())
				continue
			}
			ch <- prometheus.MustNewConstMetric(c.powerProducedTotal,
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
		for _, phase := range []int{1, 2, 3} {
			if ft := s.Feature(fmt.Sprintf("phase%dVoltage", phase)); ft.Exists() {
				v := ft.Value()
				if v == "" {
					continue
				}
				vf, err := toFloat(v)
				if err != nil {
					log.Print(err.Error())
					continue
				}
				ch <- prometheus.MustNewConstMetric(c.voltageCurrentPhase,
					prometheus.GaugeValue, vf, s.Info().Topic, strconv.Itoa(phase))
			}
			if ft := s.Feature(fmt.Sprintf("phase%dCurrent", phase)); ft.Exists() {
				v := ft.Value()
				if v == "" {
					continue
				}
				vf, err := toFloat(v)
				if err != nil {
					log.Print(err.Error())
					continue
				}
				ch <- prometheus.MustNewConstMetric(c.ampereCurrentPhase,
					prometheus.GaugeValue, vf, s.Info().Topic, strconv.Itoa(phase))
			}
		}
	}
}
