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
	"github.com/krujos/uaaclientcredentials"
	"github.com/quipo/statsd"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	cfPush            = kingpin.Flag("cf-push", "Deploy to Cloud Foundry.").Default("true").Bool()
	domain            = kingpin.Flag("domain", "Domain of your CF installation.").Default("10.244.0.34.xip.io").OverrideDefaultFromEnvar("CF_DOMAIN").String()
	dopplerPort       = kingpin.Flag("doppler-port", "Custom port for doppler / loggregator endpoint").Default("443").Int()
	subscriptionId    = kingpin.Flag("subscription-id", "ID for the firehose subscription.").Default("watchman").OverrideDefaultFromEnvar("FIREHOSE_SUBSCRIPTION_ID").String()
	clientID          = kingpin.Flag("client-id", "CF UAA OAuth client ID with 'doppler.firehose' permissions.").Default("CLIENT_ID").OverrideDefaultFromEnvar("CLIENT_ID").String()
	clientSecret      = kingpin.Flag("client-secret", "CF UAA OAuth client secret of client with 'doppler.firehose' permissions.").Default("CLIENT_SECRET").OverrideDefaultFromEnvar("CLIENT_SECRET").String()
	skipSSLValidation = kingpin.Flag("skip-ssl-validation", "Please don't").Bool()
	statsdAddress     = kingpin.Flag("statsd-address", "IP and port to the statsd endpoint.").Default("STATSD_ADDRESS").OverrideDefaultFromEnvar("STATSD_ADDRESS").String()
	statsdPrefix      = kingpin.Flag("stats-prefix", "The prefix to use for statsd metrics.").Default("cf").OverrideDefaultFromEnvar("STATSD_PREFIX").String()
)

var count = uint64(0)

func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(w,
		"Hello!\nWe have processed", atomic.LoadUint64(&count), "events",
		"\nWe're pushing to StatsD at", statsdAddress, "with a prefix of",
		statsdPrefix,
		"\nWe have tapped the firehose at ", fmt.Sprintf("wss://doppler.%s:%d", *domain, *dopplerPort))
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
	kingpin.Version("0.0.1")
	kingpin.Parse()

	if *cfPush == true {
		setupHTTP()
	}

	uaaURL, err := url.Parse(fmt.Sprintf("https://uaa.%s", *domain))

	if nil != err {
		panic("Failed to parse uaa url!")
	}

	creds, err := uaaclientcredentials.New(uaaURL, true, *clientID, *clientSecret)

	if nil != err {
		panic("Failed to obtain creds!")
	}

	dopplerAddress := fmt.Sprintf("wss://doppler.%s:%d", *domain, *dopplerPort)
	consumer := noaa.NewConsumer(dopplerAddress, &tls.Config{InsecureSkipVerify: true}, nil)

	httpStartStopProcessor := processors.NewHttpStartStopProcessor()
	sender := statsd.NewStatsdClient(*statsdAddress, *statsdPrefix)
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
		go consumer.Firehose(*subscriptionId, token, msgChan, errorChan, nil)

		for err := range errorChan {
			fmt.Fprintf(os.Stderr, "%v\n", err.Error())
		}
	}()

	for msg := range msgChan {
		eventType := msg.GetEventType()

		switch eventType {
		case events.Envelope_HttpStartStop:
			processedMetrics = httpStartStopProcessor.Process(msg)
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
