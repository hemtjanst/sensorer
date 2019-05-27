package collectors

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/kelvins/sunrisesunset"
	"github.com/prometheus/client_golang/prometheus"

	"lib.hemtjan.st/feature"
	"lib.hemtjan.st/server"
)

// EnvironmentalCollector collects sensor data from environmental sensors
type EnvironmentalCollector struct {
	humiture         *prometheus.Desc
	relativeHumidity *prometheus.Desc
	temperature      *prometheus.Desc
	daylight         *prometheus.Desc
	sunrise          *prometheus.Desc
	sunset           *prometheus.Desc

	m    *server.Manager
	lat  float64
	long float64
}

// NewEnvironmentalCollector returns a new collector for gather sensor
// metrics from environmental sensors
func NewEnvironmentalCollector(m *server.Manager, lat, long float64) (prometheus.Collector, error) {
	return &EnvironmentalCollector{
		m:    m,
		lat:  lat,
		long: long,
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
		daylight: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "daylight"),
			"Between sunrise and sunset",
			[]string{"source"}, nil,
		),
		sunrise: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "sunrise_time_seconds"),
			"Time the sun will rise today (UTC)",
			[]string{"source"}, nil,
		),
		sunset: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "sunset_time_seconds"),
			"Time the sun will set today (UTC)",
			[]string{"source"}, nil,
		),
	}, nil
}

// Describe sends all metrics descriptions into the channel
func (c *EnvironmentalCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.humiture
	ch <- c.relativeHumidity
	ch <- c.temperature
	ch <- c.daylight
	ch <- c.sunrise
	ch <- c.sunset
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

	t := time.Now().UTC()
	year, month, day := t.Date()
	p := sunrisesunset.Parameters{
		Latitude:  c.lat,
		Longitude: c.long,
		UtcOffset: 0.0,
		Date:      time.Date(year, month, day, 0, 0, 0, 0, time.UTC),
	}
	sunrise, sunset, err := p.GetSunriseSunset()
	if err != nil {
		return
	}

	sunriseT := time.Date(year, month, day, sunrise.Hour(), sunrise.Minute(), 0, 0, time.UTC)
	sunsetT := time.Date(year, month, day, sunset.Hour(), sunset.Minute(), 0, 0, time.UTC)
	ch <- prometheus.MustNewConstMetric(c.sunrise,
		prometheus.GaugeValue, float64(
			sunriseT.Unix(),
		), "sensor/astrotime")

	ch <- prometheus.MustNewConstMetric(c.sunset,
		prometheus.GaugeValue, float64(
			sunsetT.Unix(),
		), "sensor/astrotime")

	if t.After(sunriseT) && t.Before(sunsetT) {
		ch <- prometheus.MustNewConstMetric(c.daylight,
			prometheus.GaugeValue, 1.0, "sensor/astrotime")
	} else {
		ch <- prometheus.MustNewConstMetric(c.daylight,
			prometheus.GaugeValue, 0.0, "sensor/astrotime")
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
