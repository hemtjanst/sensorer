package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hemtjanst/bibliotek/server"
	"github.com/hemtjanst/bibliotek/transport/mqtt"
	"github.com/hemtjanst/sensorer"
)

func main() {
	mqttCfg := mqtt.MustFlags(flag.String, flag.Bool)
	flgAddress := flag.String("exporter.listen-address", "0.0.0.0:0", "address:port the exporter will listen on")
	flag.Parse()

	m, err := mqtt.New(context.Background(), mqttCfg())
	if err != nil {
		log.Fatal(err.Error())
	}

	mg := server.New(m)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	log.Print("starting exporter")
	shutdown, err := sensorer.NewServer(*flgAddress, mg)
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
