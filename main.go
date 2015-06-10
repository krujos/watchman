package main

// Inspired by the noaa firehose sample script
// https://github.com/cloudfoundry/noaa/blob/master/firehose_sample/main.go

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync/atomic"

	"github.com/cloudcredo/graphite-nozzle/metrics"
	"github.com/cloudcredo/graphite-nozzle/processors"
	"github.com/cloudfoundry/noaa"
	"github.com/cloudfoundry/noaa/events"
	"github.com/krujos/exceptionprocessor"
	"github.com/krujos/uaaclientcredentials"
	"github.com/quipo/statsd"
)

var dopplerAddress = os.Getenv("DOPPLER_ADDRESS")
var statsdAddress = os.Getenv("STATSD_ADDRESS")
var statsdPrefix = os.Getenv("STATSD_PREFIX")
var firehoseSubscriptionID = os.Getenv("FIREHOSE_SUBSCRIPTION_ID")
var count = uint64(0)

var uaa = os.Getenv("UAA_ADDRESS")
var clientID = os.Getenv("CLIENT_ID")
var clientSecret = os.Getenv("CLIENT_SECRET")

func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(w,
		"Hello!\nWe have processed", atomic.LoadUint64(&count), "events",
		"\nWe're pushing to StatsD at", statsdAddress, "with a prefix of",
		statsdPrefix,
		"\nWe have tapped the firehose at ", dopplerAddress)
}

func setupHTTP() {
	http.HandleFunc("/", hello)

	go func() {
		err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
		if err != nil {
			log.Fatal("ListenAndServe:", err)
		}
	}()
}

func main() {

	setupHTTP()

	uaaURL, err := url.Parse(uaa)

	if nil != err {
		panic("Failed to parse uaa url!")
	}

	creds, err := uaaclientcredentials.New(uaaURL, true, clientID, clientSecret)

	if nil != err {
		panic("Failed to obtain creds!")
	}

	consumer := noaa.NewConsumer(dopplerAddress, &tls.Config{InsecureSkipVerify: true}, nil)

	httpStartStopProcessor := processors.NewHttpStartStopProcessor()
	exceptionProcessor := exceptionprocessor.NewExceptionProcessor()
	sender := statsd.NewStatsdClient(statsdAddress, statsdPrefix)
	sender.CreateSocket()

	var processedMetrics []metrics.Metric

	msgChan := make(chan *events.Envelope)
	go func() {
		defer close(msgChan)
		errorChan := make(chan error)
		token, err := creds.GetBearerToken()
		if nil != err {
			panic(err)
		}
		go consumer.Firehose(firehoseSubscriptionID, token, msgChan, errorChan, nil)

		for err := range errorChan {
			fmt.Fprintf(os.Stderr, "%v\n", err.Error())
		}
	}()

	for msg := range msgChan {
		eventType := msg.GetEventType()

		switch eventType {
		case events.Envelope_HttpStartStop:
			processedMetrics = httpStartStopProcessor.Process(msg)
		case events.Envelope_LogMessage:
			processedMetrics = exceptionProcessor.Process(msg)
		default:
			atomic.AddUint64(&count, 1)
			// do nothing
		}

		if len(processedMetrics) > 0 {
			for _, metric := range processedMetrics {
				metric.Send(sender)
			}
		}
		processedMetrics = nil
	}
}
