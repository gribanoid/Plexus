package metrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Registry is the process-wide Prometheus registry with Go/process collectors.
var Registry = prometheus.NewRegistry()

var (
	HTTPRequestsTotal = promauto.With(Registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "plexus_http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = promauto.With(Registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "plexus_http_request_duration_seconds",
			Help:    "HTTP request latency in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	WSClients = promauto.With(Registry).NewGauge(
		prometheus.GaugeOpts{
			Name: "plexus_ws_clients",
			Help: "Number of connected WebSocket clients on this process.",
		},
	)

	JobsEnqueued = promauto.With(Registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "plexus_jobs_enqueued_total",
			Help: "Background jobs enqueued.",
		},
		[]string{"task"},
	)

	JobsProcessed = promauto.With(Registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "plexus_jobs_processed_total",
			Help: "Background jobs processed.",
		},
		[]string{"task", "result"},
	)

	Up = promauto.With(Registry).NewGauge(
		prometheus.GaugeOpts{
			Name: "plexus_up",
			Help: "1 if the process is serving.",
		},
	)

	BuildInfo = promauto.With(Registry).NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "plexus_build_info",
			Help: "Build metadata (always 1).",
		},
		[]string{"version", "commit"},
	)
)

func init() {
	Registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	Up.Set(1)
	BuildInfo.WithLabelValues(Version, Commit).Set(1)
}

// Version and Commit are set via -ldflags at build time.
var (
	Version = "dev"
	Commit  = "unknown"
)

// ObserveHTTP records RED metrics for one request.
func ObserveHTTP(method, path string, status int, seconds float64) {
	statusLabel := strconv.Itoa(status)
	HTTPRequestsTotal.WithLabelValues(method, path, statusLabel).Inc()
	HTTPRequestDuration.WithLabelValues(method, path, statusLabel).Observe(seconds)
}
