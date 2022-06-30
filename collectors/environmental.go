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
	precipitation    *prometheus.Desc
	airPressure      *prometheus.Desc
	windSpeed        *prometheus.Desc
	windDirection    *prometheus.Desc
	globalRadiation  *prometheus.Desc
	pm25             *prometheus.Desc
	airQuality       *prometheus.Desc
	waterLevel       *prometheus.Desc

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
		precipitation: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "precipitation_mm_per_hour"),
			"Precipitation rate",
			[]string{"source"}, nil,
		),
		airPressure: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "air_pressure_hpa"),
			"Atmospheric pressure",
			[]string{"source"}, nil,
		),
		windSpeed: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "wind_speed_meters_per_second"),
			"Wind Speed",
			[]string{"source"}, nil,
		),
		windDirection: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "wind_direction_degrees"),
			"Wind Direction",
			[]string{"source"}, nil,
		),
		globalRadiation: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "global_radiation_watts_per_square_meter"),
			"Global Radiation",
			[]string{"source"}, nil,
		),
		pm25: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "pm25_microgram_per_square_meter"),
			"Particulate Matter (PM2.5)",
			[]string{"source"}, nil,
		),
		airQuality: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "air_quality"),
			"Air Quality Index",
			[]string{"source"}, nil,
		),
		waterLevel: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "water_level_percent"),
			"Water Level",
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
	ch <- c.precipitation
	ch <- c.airPressure
	ch <- c.windSpeed
	ch <- c.windDirection
	ch <- c.globalRadiation
	ch <- c.pm25
	ch <- c.airQuality
	ch <- c.waterLevel
}

// Collect sends metric updates into the channel
func (c *EnvironmentalCollector) Collect(ch chan<- prometheus.Metric) {
	devices := c.m.Devices()
	humidity := map[string]float64{}
	temperature := map[string]float64{}

	for _, s := range devices {
		if s.Feature(feature.CurrentRelativeHumidity.String()).Exists() {
			v := s.Feature(feature.CurrentRelativeHumidity.String()).Value()
			if v == "" {
				continue
			}
			vf, err := toFloat(v)
			if err != nil {
				log.Print(err.Error())
				continue
			}
			humidity[s.Info().Topic] = vf
			ch <- prometheus.MustNewConstMetric(c.relativeHumidity,
				prometheus.GaugeValue, vf, s.Info().Topic)
		}
		if s.Feature(feature.CurrentTemperature.String()).Exists() {
			v := s.Feature(feature.CurrentTemperature.String()).Value()
			if v == "" {
				continue
			}
			vf, err := toFloat(v)
			if err != nil {
				log.Print(err.Error())
				continue
			}
			temperature[s.Info().Topic] = vf
			ch <- prometheus.MustNewConstMetric(c.temperature,
				prometheus.GaugeValue, vf, s.Info().Topic)
		}
		if ft := s.Feature("precipitation"); ft.Exists() {
			v := ft.Value()
			if v == "" {
				continue
			}
			vf, err := toFloat(v)
			if err != nil {
				log.Print(err.Error())
				continue
			}
			ch <- prometheus.MustNewConstMetric(c.precipitation,
				prometheus.GaugeValue, vf, s.Info().Topic)
		}
		if ft := s.Feature("airPressure"); ft.Exists() {
			v := ft.Value()
			if v == "" {
				continue
			}
			vf, err := toFloat(v)
			if err != nil {
				log.Print(err.Error())
				continue
			}
			ch <- prometheus.MustNewConstMetric(c.airPressure,
				prometheus.GaugeValue, vf, s.Info().Topic)
		}
		if ft := s.Feature("windSpeed"); ft.Exists() {
			v := ft.Value()
			if v == "" {
				continue
			}
			vf, err := toFloat(v)
			if err != nil {
				log.Print(err.Error())
				continue
			}
			ch <- prometheus.MustNewConstMetric(c.windSpeed,
				prometheus.GaugeValue, vf, s.Info().Topic)
		}
		if ft := s.Feature("windDirection"); ft.Exists() {
			v := ft.Value()
			if v == "" {
				continue
			}
			vf, err := toFloat(v)
			if err != nil {
				log.Print(err.Error())
				continue
			}
			ch <- prometheus.MustNewConstMetric(c.windDirection,
				prometheus.GaugeValue, vf, s.Info().Topic)
		}
		if ft := s.Feature("globalRadiation"); ft.Exists() {
			v := ft.Value()
			if v == "" {
				continue
			}
			vf, err := toFloat(v)
			if err != nil {
				log.Print(err.Error())
				continue
			}
			ch <- prometheus.MustNewConstMetric(c.globalRadiation,
				prometheus.GaugeValue, vf, s.Info().Topic)
		}
		if ft := s.Feature("pm2_5Density"); ft.Exists() {
			v := ft.Value()
			if v == "" {
				continue
			}
			vf, err := toFloat(v)
			if err != nil {
				log.Print(err.Error())
				continue
			}
			ch <- prometheus.MustNewConstMetric(c.pm25,
				prometheus.GaugeValue, vf, s.Info().Topic)
		}
		if ft := s.Feature("airQuality"); ft.Exists() {
			v := ft.Value()
			if v == "" {
				continue
			}
			vf, err := toFloat(v)
			if err != nil {
				log.Print(err.Error())
				continue
			}
			ch <- prometheus.MustNewConstMetric(c.airQuality,
				prometheus.GaugeValue, vf, s.Info().Topic)
		}
		if ft := s.Feature("waterLevel"); ft.Exists() {
			v := ft.Value()
			if v == "" {
				continue
			}
			vf, err := toFloat(v)
			if err != nil {
				log.Print(err.Error())
				continue
			}
			ch <- prometheus.MustNewConstMetric(c.waterLevel,
				prometheus.GaugeValue, vf, s.Info().Topic)
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
