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

There are two endpoints:

* `/metrics` which contains metrics internal to Sensorer, like HTTP
  request statistics, memory utilisation etc
* `/sensors` with the sensor metrics

Which metrics are exported depends on the features the device announces.
Every exported metric has a label named `source` which holds the device's
MQTT topic.

* `currentTemperature`: gauge `sensors_temperature_celsius`
* `currentRelativeHumidity`: gauge `sensors_humidity_relative_percent`
* `contactSensorState`: gauge `sensors_contact_state`
* `currentPower`: gauge `sensors_power_current_watts`
* `energyUsed`: counter `sensors_power_total_kwh`
* `currentVoltage`: gauge `sensors_power_current_voltage`
* `currentAmpere`: gauge `sensors_power_current_ampere`
* `batteryLevel`: gauge `sensors_battery_level_percent`

### Built-in sensors

A time series is computed for humiture, also known as
the "feels like" temperature: `sensors_humiture_celsius`.

Two time series are computed based on `location.lat` and `location.long`,
respectively `sensors_sunrise_time_seconds` and `sensors_sunset_time_seconds`.
A third time series, `sensors_daylight` returns 1 if the current time
is between sunrise and sunset, and 0 otherwise.

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
