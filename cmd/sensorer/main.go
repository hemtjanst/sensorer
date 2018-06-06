package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/hemtjanst/hemtjanst/device"
	"github.com/hemtjanst/hemtjanst/messaging"
	"github.com/hemtjanst/hemtjanst/messaging/flagmqtt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// SensorData is a map of sensor values as received from
// MQTT
type SensorData struct {
	Type  string
	Data  float64
	Topic string
}

// Container for all the SensorData
type Container struct {
	sync.RWMutex
	Sensors map[string]*SensorData
}

// Metrics holds the different metric types we want to publish
type Metrics struct {
	Gauge   map[string]*prometheus.GaugeVec
	Counter map[string]*prometheus.CounterVec
}

// Register a new Sensor
func (c *Container) Register(topic, feature string, value *device.Feature, metrics *Metrics, debug *bool) {
	value.OnUpdate(func(ms messaging.Message) {
		v, err := strconv.ParseFloat(string(ms.Payload()), 64)
		if err != nil {
			if *debug {
				log.Printf("Received data does not appear to be a float: %s %v", string(ms.Payload()), err)
			}
			return
		}

		c.Lock()
		defer c.Unlock()
		if sensor, ok := c.Sensors[ms.Topic()]; ok {
			sensor.Data = v
		} else {
			c.Sensors[ms.Topic()] = &SensorData{
				Type:  feature,
				Topic: topic,
				Data:  v,
			}
			log.Printf("Watching %s for updates on %s", topic, feature)
		}

		switch strings.ToLower(feature) {
		case "currenttemperature":
			metrics.Gauge["temperature"].WithLabelValues(topic).Set(v)
			// Attempt to calculate realtive humidity
			t := strings.Split(ms.Topic(), "/")
			i := Position("temperature", t)
			if i == -1 {
				return
			}
			t[i] = "humidity"
			i = Position("currentTemperature", t)
			if i == -1 {
				return
			}
			t[i] = "currentRelativeHumidity"
			nt := strings.Join(t[:], "/")
			if sensor, ok := c.Sensors[nt]; ok {
				metrics.Gauge["humiture"].WithLabelValues(strings.Replace(topic, "/temperature", "", 1)).Set(
					Humiture(v, sensor.Data))
			}
		case "currentrelativehumidity":
			metrics.Gauge["humidity"].WithLabelValues(topic).Set(v)
			// Attempt to calculate realtive humidity
			t := strings.Split(ms.Topic(), "/")
			i := Position("humidity", t)
			if i == -1 {
				return
			}
			t[i] = "temperature"
			i = Position("currentRelativeHumidity", t)
			if i == -1 {
				return
			}
			t[i] = "currentTemperature"
			nt := strings.Join(t[:], "/")
			if sensor, ok := c.Sensors[nt]; ok {
				metrics.Gauge["humiture"].WithLabelValues(strings.Replace(topic, "/humidity", "", 1)).Set(
					Humiture(sensor.Data, v))
			}
		case "contactsensorstate":
			metrics.Gauge["contact"].WithLabelValues(topic).Set(v)
		case "currentpower":
			metrics.Gauge["power"].WithLabelValues(topic).Set(v)
		case "energyused":
			metrics.Counter["power"].WithLabelValues(topic).Set(v)
		}
		if *debug {
			log.Printf("Updated %s on %s to %f", feature, topic, v)
		}
	})
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Parameters:\n\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n")
	}
	addr := flag.String("metrics.address", "localhost:9123", "Address to expose metrics on")
	debug := flag.Bool("debug", false, "Turn on debugging, generates more logs")
	flag.Parse()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sensors := &Container{Sensors: map[string]*SensorData{}}

	id := flagmqtt.NewUniqueIdentifier()
	conn, err := flagmqtt.NewPersistentMqtt(flagmqtt.ClientConfig{
		ClientID:    id,
		WillTopic:   "leave",
		WillPayload: id,
	})
	if err != nil {
		log.Fatal("Could not configure the MQTT client: ", err)
	}
	messenger := messaging.NewMQTTMessenger(conn)

	if token := conn.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal("Failed to establish connection with broker: ", token.Error())
	}

	devices := map[string]*device.Device{}
	md := sync.RWMutex{}
	metrics := NewMetrics()

	messenger.Subscribe("announce/#", 1, func(m messaging.Message) {
		t := m.Topic()[9:]
		if *debug {
			log.Printf("Received announcement for: %s", t)
		}
		md.RLock()
		if _, ok := devices[t]; !ok {
			md.RUnlock()
			d := device.NewDevice(t, messenger)
			err := d.UnmarshalJSON(m.Payload())
			if err != nil {
				if *debug {
					log.Printf("Could not decode device: %s, %v", t, err)
				}
				return
			}

			for name, value := range d.Features {
				switch strings.ToLower(name) {
				case "currenttemperature", "currentrelativehumidity", "currentpower", "energyused", "contactsensorstate":
					if *debug {
						log.Printf("Found feature %s on device %s", name, t)
					}
					md.Lock()
					devices[t] = d
					if *debug {
						log.Printf("Added device %s", t)
					}
					md.Unlock()
					sensors.Register(t, name, value, metrics, debug)
				}
			}
		} else {
			md.RUnlock()
		}
	})

	h := http.Server{Addr: *addr, Handler: promhttp.Handler()}

	go func() {
		if err := h.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	sig := <-quit
	log.Printf("Received signal: %s, proceeding to shutdown", sig)
	conn.Disconnect(250)
	log.Print("Disconnected from broker. Bye!")
	os.Exit(0)
}

// NewMetrics creates all the Prometheus metrics we want to track
func NewMetrics() *Metrics {
	temperature := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "sensor",
			Name:      "temperature_celsius",
			Help:      "Temperature in degrees Celsius",
		},
		[]string{
			"source",
		},
	)
	humiture := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "sensor",
			Name:      "humiture_celsius",
			Help:      "Heat Index ('feels like temperature') in degrees Celsius",
		},
		[]string{
			"source",
		},
	)
	humidity := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "sensor",
			Name:      "humidity_relative",
			Help:      "Relative Humidity in percent",
		},
		[]string{
			"source",
		},
	)
	contact := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "sensor",
			Name:      "contact_state",
			Help:      "Contact sensor state",
		},
		[]string{
			"source",
		},
	)

	powerUsage := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "sensor",
			Subsystem: "power",
			Name:      "current_watts",
			Help:      "Current power draw",
		},
		[]string{
			"source",
		},
	)

	powerTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "sensor",
			Subsystem: "power",
			Name:      "total_kwh",
			Help:      "Total power usage",
		},
		[]string{
			"source",
		},
	)

	prometheus.MustRegister(temperature)
	prometheus.MustRegister(humiture)
	prometheus.MustRegister(humidity)
	prometheus.MustRegister(contact)
	prometheus.MustRegister(powerUsage)
	prometheus.MustRegister(powerTotal)

	return &Metrics{
		Gauge: map[string]*prometheus.GaugeVec{
			"temperature": temperature,
			"humiture":    humiture,
			"humidity":    humidity,
			"contact":     contact,
			"power":       powerUsage,
		},
		Counter: map[string]*prometheus.CounterVec{
			"power": powerTotal,
		},
	}
}

// Humiture returns the Heat Index in degrees Celsius.
// This is also known as the "feels like" temperatue,
// "felt air temperature" or "apparent temperature".
// https://en.wikipedia.org/wiki/Heat_index
func Humiture(temp, relativeHumidity float64) float64 {
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
	} else {
		return 0.5 * (temp + 16.1 + ((temp - 21) * 1.2) + (relativeHumidity * 0.094))
	}
}

// Position returns the position of a string in a slice, or
// -1 if it is not found
func Position(what string, in []string) int {
	for p, v := range in {
		if v == what {
			return p
		}
	}
	return -1
}
