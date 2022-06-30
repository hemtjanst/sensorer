package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hemtjan.st/sensorer"
	"lib.hemtjan.st/server"
	"lib.hemtjan.st/transport/mqtt"
)

var (
	version = "unknown"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	mqttCfg := mqtt.MustFlags(flag.String, flag.Bool)
	flgAddress := flag.String("exporter.listen-address", "0.0.0.0:0", "address:port the exporter will listen on")
	flgLatitude := flag.Float64("location.lat", 0.0, "latitude of location for sunrise/sunset")
	flgLongitude := flag.Float64("location.long", 0.0, "longitude of location for sunrise/sunset")
	flgVersion := flag.Bool("version", false, "print version info and exit")
	flag.Parse()

	if *flgVersion {
		fmt.Fprintf(
			os.Stdout,
			`{"version": "%s", "commit": "%s", "date": "%s"}`,
			version, commit, date,
		)
		os.Exit(0)
	}

	m, err := mqtt.New(context.Background(), mqttCfg())
	if err != nil {
		log.Fatal(err.Error())
	}
	go func() {
		for {
			ok, err := m.Start()
			if !ok {
				break
			}
			log.Printf("Error, retrying in 5 seconds: %v", err)
			time.Sleep(5 * time.Second)
		}
		os.Exit(1)
	}()

	mg := server.New(m)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	log.Print("starting exporter")
	shutdown, err := sensorer.NewServer(*flgAddress, *flgLatitude, *flgLongitude, mg)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Print("started exporter")

	<-stop
	log.Print("shutting down exporter")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	shutdown(ctx)
	log.Print("shutdown exporter")
}
