# Sensorer

Sensorer, aka "sensors", reads sensor data from MQTT and makes it available
for scraping to Prometheus. It's a Hemtjanst exporter.

This package was written with Go 1.10 but should work with Go 1.8 onwards.

## Installation

* `go install github.com/hemtjanst/sensorer/cmd/sensorer`

### Development

* `git clone github.com/hemtjanst/sensorer`
* `go mod download`

## Exported metrics

Which metrics are exported depends on the features the device announces.
Every exported metric has a label named `source` which holds the device's
MQTT topic.

* `currentTemperature`: gauge `sensor_temperature_celsius`
* `currentRelativeHumidity`: gauge `sensor_humidity_relative_percent`
* `contactSensorState`: gauge `sensor_contact_state`
* `currentPower`: gauge `sensor_power_current_watts`
* `energyUsed`: counter `sensor_power_total_kwh`
* `batteryLevel`: gauge `sensor_battery_level_percent`

Additionally a time series is exposed for humiture, also known as
the "feels like" temperature: `sensor_humiture_celsius`.

## Options

A number of options can be passed at startup in order to configure the
behaviour. Most importantly are probably `-exporter.listen-address` and
`-mqtt.address`. These let you configure on what `host:port` combination
the metrics are exported and on what `host:port` combination the MQTT
broker can be found.

Issue a `sensorer -help` for all possible options.

## Caveats

Depending on the Prometheus scrape time and how certain contact sensors
report in, it's very possible that you would not be able to see something
like a door opening and closing in fairly rapid succession reflected
in Prometheus. As such you **must not** rely on this data for the purposes
of home security.
