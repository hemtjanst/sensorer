package collectors

import (
	"fmt"
	"log"
	"math"

	"github.com/hemtjanst/bibliotek/feature"

	"github.com/hemtjanst/bibliotek/server"
	"github.com/prometheus/client_golang/prometheus"
)

// EnvironmentalCollector collects sensor data from environmental sensors
type EnvironmentalCollector struct {
	humiture         *prometheus.Desc
	relativeHumidity *prometheus.Desc
	temperature      *prometheus.Desc

	m *server.Manager
}

// NewEnvironmentalCollector returns a new collector for gather sensor
// metrics from environmental sensors
func NewEnvironmentalCollector(m *server.Manager) (prometheus.Collector, error) {
	return &EnvironmentalCollector{
		m: m,
		humiture: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "humiture_celsius"),
			"Heat Index ('feels like temperature') in degrees Celsius",
			[]string{"source"}, nil,
		),
		relativeHumidity: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "humidity", "relative_percent"),
			"Relative Humidity in percent",
			[]string{"source"}, nil,
		),
		temperature: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "temperature_celsius"),
			"Temperature in degrees Celsius",
			[]string{"source"}, nil,
		),
	}, nil
}

// Describe sends all metrics descriptions into the channel
func (c *EnvironmentalCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.humiture
	ch <- c.relativeHumidity
	ch <- c.temperature
}

// Collect sends metric updates into the channel
func (c *EnvironmentalCollector) Collect(ch chan<- prometheus.Metric) {
	devices := c.m.Devices()
	humidity := map[string]float64{}
	temperature := map[string]float64{}

	for _, s := range devices {
		if s.Feature(feature.CurrentRelativeHumidity.String()).Exists() {
			v, err := toFloat(s.Feature(feature.CurrentRelativeHumidity.String()).Value())
			if err != nil {
				log.Print(err.Error())
				continue
			}
			humidity[last(s.Info().Topic, "/")] = v
			ch <- prometheus.MustNewConstMetric(c.relativeHumidity,
				prometheus.GaugeValue, v, s.Info().Topic)
		}
		if s.Feature(feature.CurrentTemperature.String()).Exists() {
			v, err := toFloat(s.Feature(feature.CurrentTemperature.String()).Value())
			if err != nil {
				log.Print(err.Error())
				continue
			}
			temperature[last(s.Info().Topic, "/")] = v
			ch <- prometheus.MustNewConstMetric(c.temperature,
				prometheus.GaugeValue, v, s.Info().Topic)
		}
	}

	for dev, temp := range temperature {
		if hum, ok := humidity[dev]; ok {
			ch <- prometheus.MustNewConstMetric(c.humiture,
				prometheus.GaugeValue, humiture(temp, hum), fmt.Sprintf("sensor/humiture/%s", dev))
		}
	}
}

// humiture returns the Heat Index in degrees Celsius.
// This is also known as the "feels like" temperatue,
// "felt air temperature" or "apparent temperature".
// https://en.wikipedia.org/wiki/Heat_index
func humiture(temp, relativeHumidity float64) float64 {
	c1 := -8.784695
	c2 := 1.61139411
	c3 := 2.33854900
	c4 := -0.14611605
	c5 := -0.01230809
	c6 := -0.01642482
	c7 := 0.00221173
	c8 := 0.00072546
	c9 := -0.00000358

	if temp >= 26.0 {
		return (c1 + (c2 * temp) +
			(c3 * relativeHumidity) +
			(c4 * temp * relativeHumidity) +
			(c5 * math.Pow(temp, 2)) +
			(c6 * math.Pow(relativeHumidity, 2)) +
			(c7 * math.Pow(temp, 2) * relativeHumidity) +
			(c8 * temp * math.Pow(relativeHumidity, 2)) +
			(c9 * math.Pow(temp, 2) * math.Pow(relativeHumidity, 2)))
	}

	return 0.5 * (temp + 16.1 + ((temp - 21) * 1.2) + (relativeHumidity * 0.094))
}
