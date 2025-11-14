package shared

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/api/errors"
)

type APIMetrics struct {
	guard                 *sync.Mutex
	invalidSubsystemChars *regexp.Regexp

	namespace            string
	endpointRequestCount map[string]*prometheus.CounterVec
	endpointLatencySec   map[string]*prometheus.HistogramVec
	commonLabels         map[string]string
}

func NewAPIMetrics(namespace string, commonLabels map[string]string) *APIMetrics {
	return &APIMetrics{
		guard:                 &sync.Mutex{},
		invalidSubsystemChars: regexp.MustCompilePOSIX(`[/.-]`),
		endpointRequestCount:  make(map[string]*prometheus.CounterVec),
		endpointLatencySec:    make(map[string]*prometheus.HistogramVec),
		namespace:             namespace,
		commonLabels:          commonLabels,
	}
}

func (a *APIMetrics) endpointToSubsystem(endpoint string) string {
	if endpointURL, err := url.Parse(endpoint); err == nil {
		if hostname := endpointURL.Hostname(); len(hostname) > 0 {
			endpoint = hostname
		}
	}

	// Replace invalid characters with underscores to create a valid subsystem name
	return strings.TrimRight(a.invalidSubsystemChars.ReplaceAllString(endpoint, "_"), "_")
}

// TrackRequest tracks the number of requests to a specific API endpoint.
// It creates a counter for the endpoint if it doesn't already exist.
// The counter is used to count the total number of requests, categorized by status and path.
func (a *APIMetrics) TrackRequest(endpoint, path string, status int) error {
	subsystem := a.endpointToSubsystem(endpoint)
	a.guard.Lock()
	defer a.guard.Unlock()

	counter, ok := a.endpointRequestCount[subsystem]
	if !ok {
		counter = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace:   a.namespace,
			Subsystem:   subsystem,
			Name:        "requests_total",
			Help:        fmt.Sprintf("Total number of requests to the %s API endpoint.", endpoint),
			ConstLabels: a.commonLabels,
		}, []string{"status", "path"})

		if err := RegisterCollectorOrUseExisting(&counter); err != nil {
			log.Warn().Err(err).Msgf("Failed to register request counter for endpoint %s, metrics will not be available", endpoint)
			return err
		}
		a.endpointRequestCount[subsystem] = counter
	}

	log.Debug().Msgf("Registered counter value for %s_%s{path=\"%s\", status=\"%d\"}", a.namespace, subsystem, path, status)
	counter.WithLabelValues(strconv.Itoa(status), path).Inc()
	return nil
}

// TrackDuration tracks the duration of requests to a specific API endpoint.
// It creates a histogram for the endpoint if it doesn't already exist.
// The histogram is used to observe the duration of requests in seconds.
func (a *APIMetrics) TrackDuration(endpoint, path string, d time.Duration) error {
	subsystem := a.endpointToSubsystem(endpoint)
	a.guard.Lock()
	defer a.guard.Unlock()

	histogram, ok := a.endpointLatencySec[subsystem]
	if !ok {
		histogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace:   a.namespace,
			Subsystem:   subsystem,
			Name:        "request_duration_seconds",
			Help:        fmt.Sprintf("Duration of requests to the %s API endpoint.", endpoint),
			ConstLabels: a.commonLabels,
			Buckets:     []float64{0.005, 0.01, 0.025, 0.050, 0.075, 0.1, 0.15, 0.2, 0.25, 0.3, 0.4, 0.5, 0.75, 1, 1.5, 2, 3},
		}, []string{"path"})

		if err := RegisterCollectorOrUseExisting(&histogram); err != nil {
			log.Warn().Err(err).Msgf("Failed to register request duration for endpoint %s, metrics will not be available", endpoint)
			return err
		}

		a.endpointLatencySec[subsystem] = histogram
	}

	histogram.WithLabelValues(path).Observe(d.Seconds())
	log.Debug().Msgf("Registered histogram value for %s_%s{path=\"%s\"}", a.namespace, subsystem, path)
	return nil
}

// TrackCallResponse tracks both the duration and the status code of an API call.
// It extracts the status code from the http.Response if available, or from the error if not.
// If neither is available, it assumes a status code of 200 for no error, or -1 for an unknown error.
func (a *APIMetrics) TrackCallResponse(endpoint, path string, requestStart time.Time, rsp *http.Response, err error) {
	statusCode := 200

	switch {
	case rsp != nil:
		statusCode = rsp.StatusCode
	case err != nil:
		switch typedErr := err.(type) {
		case *ErrorWithStatus:
			statusCode = typedErr.Code
		case *errors.StatusError:
			statusCode = int(typedErr.ErrStatus.Code)
		default:
			statusCode = -1
		}
	}

	_ = a.TrackDuration(endpoint, path, time.Since(requestStart))
	_ = a.TrackRequest(endpoint, path, statusCode)
}

// RegisterCollectorOrUseExisting registers a Prometheus collector.
// If a metric of the same type is already registered, it changes the
// given metric handle to use the existing one and returns nil.
// If the metric is not of the expected type, it returns the AlreadyRegisteredError.
func RegisterCollectorOrUseExisting[T prometheus.Collector](metric *T) error {
	err := prometheus.Register(*metric)
	if err == nil {
		return nil
	}

	if current, alreadyExists := err.(prometheus.AlreadyRegisteredError); alreadyExists {
		if typedCollector, isOfT := current.ExistingCollector.(T); isOfT {
			*metric = typedCollector
			return nil
		}
	}

	return err
}
