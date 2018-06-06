# Sensorer

Sensorer, aka "sensors", reads sensor data from MQTT and makes it available
for scraping to Prometheus. It's a Hemtjanst exporter.

This package was written with Go 1.10 but should work with Go 1.8 onwards.

## Installation / Development

* `git clone` the repository
* `dep ensure` to fetch the right versions of the dependencies
* `go build -o sensorer cmd/sensorer/main.go`

## Exported metrics

Which metrics are exported depends on the features the device announces.
Every exported metric has a label named `source` which holds the device's
MQTT topic.

* `currentTemperature`: gauge `sensor_temperature_celcius`
* `currentRelativeHumidity`: gauge `sensor_humidity_relative`
* `contactSensorState`: gauge `sensor_contact_state`
* `currentPower`: gauge `sensor_power_current_watts`
* `energyUsed`: counter `sensor_power_total_kwh`

## Options

A number of options can be passed at startup in order to configure the
behaviour. Most importantly are probably `-metrics.address` and
`-mqtt.address`. These let you configure on what `host:port` combination
the metrics are exported and on what `host:port` combination the MQTT
broker can be found.

Issue a `sensorer -help` for all possible options.

## Caveats

Depending on the Prometheus scrape time and how certain contact sensors
report in, it's very possible that you would not be able to see something
like a door opening and closing in fairly rapid succession reflected
in Prometheus. As such you **must not** rely on this data for the purposes
of alarming/alerting.