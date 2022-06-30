package collectors // import "hemtjan.st/sensorer/collectors"

import (
	"strconv"
)

// namespace is the Prometheus namespaces for all collectors in this package
const namespace = "sensors"

// toFloat takes a string and parses it into a float
func toFloat(s string) (float64, error) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0, err
	}
	return f, nil
}
