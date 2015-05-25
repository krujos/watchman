package main

// Inspired by the noaa firehose sample script
// https://github.com/cloudfoundry/noaa/blob/master/firehose_sample/main.go

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/cloudcredo/graphite-nozzle/metrics"
	"github.com/cloudcredo/graphite-nozzle/processors"
	"github.com/cloudfoundry/noaa"
	"github.com/cloudfoundry/noaa/events"
	"github.com/quipo/statsd"
)

var dopplerAddress = os.Getenv("DOPPLER_ADDRESS")
var statsdAddress = os.Getenv("STATSD_ADDRESS")
var statsdPrefix = os.Getenv("STATSD_PREFIX")
var firehoseSubscriptionID = os.Getenv("FIREHOSE_SUBSCRIPTION_ID")
var authToken = os.Getenv("CF_ACCESS_TOKEN")

func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(w, "hello, world!")
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

/*
func getAccessTokenFromUAA() string {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	url := "https://admin:admin@uaa.10.244.0.34.xip.io/uaa/oauth/authorize?grant_type=password&response_type=token&username=admin&password=admin&client_id=watchman"
	//	var b = strings.NewReader(`{"username":"admin","password":"admin"}`)

	resp, err := client.Post(url, "application/json", nil)

	if nil != err {
		log.Fatal(err)
		panic("Failed to post to UAA!")
	}
	//log.Print("Location: " + resp.Header.Get("Location"))
	for key, value := range resp.Header {
		fmt.Println("Key:", key, "Value:", value)
	}

	log.Print("Status Code: ")
	log.Print(resp.StatusCode)

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	return string(body)
}
*/
func main() {

	setupHTTP()

	//authToken := getAccessTokenFromUAA()
	consumer := noaa.NewConsumer(dopplerAddress, &tls.Config{InsecureSkipVerify: true}, nil)

	httpStartStopProcessor := processors.NewHttpStartStopProcessor()
	sender := statsd.NewStatsdClient(statsdAddress, statsdPrefix)
	sender.CreateSocket()

	var processedMetrics []metrics.Metric

	msgChan := make(chan *events.Envelope)
	go func() {
		defer close(msgChan)
		errorChan := make(chan error)
		go consumer.Firehose(firehoseSubscriptionID, authToken, msgChan, errorChan, nil)

		for err := range errorChan {
			fmt.Fprintf(os.Stderr, "%v\n", err.Error())
		}
	}()

	for msg := range msgChan {
		eventType := msg.GetEventType()

		// graphite-nozzle can handle CounterEvent, ContainerMetric, Heartbeat,
		// HttpStartStop and ValueMetric events
		switch eventType {

		case events.Envelope_HttpStartStop:
			processedMetrics = httpStartStopProcessor.Process(msg)
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
