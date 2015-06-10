package exceptionprocessor

import (
	"regexp"

	"github.com/cloudcredo/graphite-nozzle/metrics"
	"github.com/cloudfoundry/noaa/events"
)

//ExceptionProcessor searches for the word "Exception" in the log stream. We
//add a counter for every exception.
type ExceptionProcessor struct{}

//NewExceptionProcessor creates a processor
func NewExceptionProcessor() *ExceptionProcessor {
	return &ExceptionProcessor{}
}

var javaException = regexp.MustCompile("(?i)exception")
var rubyException = regexp.MustCompile("in `block in ")

//Process does the work of processing the metric. Returns nil if message has
//no exception
func (processor *ExceptionProcessor) Process(e *events.Envelope) []metrics.Metric {
	processedMetrics := make([]metrics.Metric, 1)
	processedMetrics[0] = processor.processLogMessage(e.GetLogMessage())
	return processedMetrics
}

func (processor *ExceptionProcessor) processLogMessage(l *events.LogMessage) *metrics.CounterMetric {

	hasException := int64(0)
	if javaException.Match(l.GetMessage()) || rubyException.Match(l.GetMessage()) {
		hasException = int64(1)
	}

	stat := l.GetAppId() + "-exceptions"
	metric := metrics.NewCounterMetric(stat, hasException)

	return metric
}
