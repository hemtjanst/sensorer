package sensorer // import "hemtjan.st/sensorer"

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"hemtjan.st/sensorer/collectors"
	"lib.hemtjan.st/server"
)

var (
	latitude  float64
	longitude float64
)

// NewPrometheusMetrics returns a Prometheus registry with metrics that
// instrument the exporter itself
func NewPrometheusMetrics() *prometheus.Registry {
	p := prometheus.NewPedanticRegistry()
	p.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	p.MustRegister(prometheus.NewGoCollector())
	return p
}

// NewSensorMetrics returns a Prometheus registry with sensor related
// collectors
func NewSensorMetrics(mg *server.Manager) (*prometheus.Registry, error) {
	p := prometheus.NewPedanticRegistry()
	c, err := collectors.NewBatteryCollector(mg)
	if err != nil {
		return nil, err
	}
	p.MustRegister(c)

	c, err = collectors.NewContactCollector(mg)
	if err != nil {
		return nil, err
	}
	p.MustRegister(c)

	c, err = collectors.NewPowerCollector(mg)
	if err != nil {
		return nil, err
	}
	p.MustRegister(c)

	c, err = collectors.NewEnvironmentalCollector(mg, latitude, longitude)
	if err != nil {
		return nil, err
	}
	p.MustRegister(c)

	c, err = collectors.NewFilterCollector(mg)
	if err != nil {
		return nil, err
	}
	p.MustRegister(c)
	return p, nil
}

// NewServer starts an HTTP server exposing Prometheus metrics
func NewServer(addr string, lat, long float64, mg *server.Manager) (func(context.Context), error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	latitude = lat
	longitude = long
	sensors, err := NewSensorMetrics(mg)
	if err != nil {
		return nil, err
	}
	promMetrics := NewPrometheusMetrics()
	ctx, ctxCancel := context.WithCancel(context.Background())
	go mg.Start(ctx)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(promMetrics, promhttp.HandlerOpts{}))
	mux.Handle("/sensors", promhttp.InstrumentMetricHandler(promMetrics, promhttp.HandlerFor(sensors, promhttp.HandlerOpts{})))
	h := &http.Server{
		Handler: mux,
	}
	go func() {
		if err := h.Serve(listener); err != http.ErrServerClosed {
			log.Fatal(err.Error())
		}
	}()

	log.Printf("exporter listening on: %s", listener.Addr().String())

	return func(ctx context.Context) {
		ctxCancel()
		h.Shutdown(ctx)
	}, nil
}
