// +build !js

package goa

import (
	"regexp"
	"strings"
	"time"

	"github.com/armon/go-metrics"
)

const (
	allMatcher      string = "*/*"
	allReplacement  string = "all"
	normalizedToken string = "_"
)

var (
	// interface for metrics.Metrics
	metriks Metrics = &metricDiscarder{}

	// used for normalizing names by matching '*' and '/' so they can be replaced.
	invalidCharactersRE = regexp.MustCompile(`[\*/]`)

	// Taken from https://github.com/prometheus/client_golang/blob/66058aac3a83021948e5fb12f1f408ff556b9037/prometheus/desc.go
	metricsNameRE = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_:]*$`)
)

// Metrics generic interface for decoupling Goa from go-metrics.
type Metrics interface {
	SetGauge(key []string, val float32)
	EmitKey(key []string, val float32)
	IncrCounter(key []string, val float32)
	AddSample(key []string, val float32)
	MeasureSince(key []string, start time.Time)
}

// used for dealing with race conditions.
type metricDiscarder struct{}
func (md *metricDiscarder) SetGauge(key []string, val float32) {}
func (md *metricDiscarder) EmitKey(key []string, val float32) {}
func (md *metricDiscarder) IncrCounter(key []string, val float32) {}
func (md *metricDiscarder) AddSample(key []string, val float32) {}
func (md *metricDiscarder) MeasureSince(key []string, start time.Time) {}


// NewMetrics initializes goa's metrics instance with the supplied
// configuration and metrics sink
// This method is deprecated and SetMetrics should be used instead.
func NewMetrics(conf *metrics.Config, sink metrics.MetricSink) (err error) {
	metriks, err = metrics.NewGlobal(conf, sink)
	return
}

// SetMetrics initializes goa's metrics instance with the supplied metrics adapter interface.
func SetMetrics(m Metrics) {
	metriks = m
}

// AddSample adds a sample to an aggregated metric
// reporting count, min, max, mean, and std deviation
// Usage:
//     AddSample([]string{"my","namespace","key"}, 15.0)
func AddSample(key []string, val float32) {
	if metriks != nil {
		normalizeKeys(key)
		metriks.AddSample(key, val)
	}
}

// EmitKey emits a key/value pair
// Usage:
//     EmitKey([]string{"my","namespace","key"}, 15.0)
func EmitKey(key []string, val float32) {
	if metriks != nil {
		normalizeKeys(key)
		metriks.EmitKey(key, val)
	}
}

// IncrCounter increments the counter named by `key`
// Usage:
//     IncrCounter([]key{"my","namespace","counter"}, 1.0)
func IncrCounter(key []string, val float32) {
	if metriks != nil {
		normalizeKeys(key)
		metriks.IncrCounter(key, val)
	}
}

// MeasureSince creates a timing metric that records
// the duration of elapsed time since `start`
// Usage:
//     MeasureSince([]string{"my","namespace","action}, time.Now())
// Frequently used in a defer:
//    defer MeasureSince([]string{"my","namespace","action}, time.Now())
func MeasureSince(key []string, start time.Time) {
	if metriks != nil {
		normalizeKeys(key)
		metriks.MeasureSince(key, start)
	}
}

// SetGauge sets the named gauge to the specified value
// Usage:
//     SetGauge([]string{"my","namespace"}, 2.0)
func SetGauge(key []string, val float32) {
	if metriks != nil {
		normalizeKeys(key)
		metriks.SetGauge(key, val)
	}
}

// This function is used to make metric names safe for all metric services. Specifically, prometheus does
// not support * or / in metric names.
func normalizeKeys(key []string) {
	for i, k := range key {
		if !metricsNameRE.MatchString(k) {
			// first replace */* with all
			k = strings.Replace(k, allMatcher, allReplacement, -1)

			// now replace all other invalid characters with a safe one.
			key[i] = invalidCharactersRE.ReplaceAllString(k, normalizedToken)
		}
	}
}
