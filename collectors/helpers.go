package collectors

import (
	"strconv"
	"strings"
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

// last returns the last element of s split on sep
func last(s, sep string) string {
	sp := strings.Split(s, sep)
	return sp[len(sp)-1]
}
