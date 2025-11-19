package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	namespace = "board_service"
)

// Metrics holds all application metrics
type Metrics struct {
	// HTTP metrics
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec

	// Database metrics
	DBConnectionsOpen        prometheus.Gauge
	DBConnectionsInUse       prometheus.Gauge
	DBConnectionsIdle        prometheus.Gauge
	DBConnectionsMax         prometheus.Gauge
	DBConnectionWaitTotal    prometheus.Counter
	DBConnectionWaitDuration prometheus.Counter
	DBQueryDuration          *prometheus.HistogramVec
	DBQueryErrors            *prometheus.CounterVec

	// External API metrics
	ExternalAPIRequestDuration *prometheus.HistogramVec
	ExternalAPIRequestsTotal   *prometheus.CounterVec
	ExternalAPIErrors          *prometheus.CounterVec

	// Business metrics
	ProjectsTotal       prometheus.Gauge
	BoardsTotal         prometheus.Gauge
	ProjectCreatedTotal prometheus.Counter
	BoardCreatedTotal   prometheus.Counter
}

// New creates and registers all metrics
func New() *Metrics {
	return &Metrics{
		// HTTP metrics
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request duration in seconds",
				Buckets:   []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"method", "endpoint"},
		),

		// Database connection pool metrics
		DBConnectionsOpen: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "db_connections_open",
				Help:      "Current number of open database connections",
			},
		),
		DBConnectionsInUse: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "db_connections_in_use",
				Help:      "Current number of in-use database connections",
			},
		),
		DBConnectionsIdle: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "db_connections_idle",
				Help:      "Current number of idle database connections",
			},
		),
		DBConnectionsMax: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "db_connections_max",
				Help:      "Maximum number of open database connections configured",
			},
		),
		DBConnectionWaitTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "db_connection_wait_total",
				Help:      "Total number of times waited for a database connection",
			},
		),
		DBConnectionWaitDuration: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "db_connection_wait_duration_seconds_total",
				Help:      "Total duration waited for database connections in seconds",
			},
		),

		// Database query metrics
		DBQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "db_query_duration_seconds",
				Help:      "Database query duration in seconds",
				Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5},
			},
			[]string{"operation", "table"},
		),
		DBQueryErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "db_query_errors_total",
				Help:      "Total number of database query errors",
			},
			[]string{"operation", "table"},
		),

		// External API metrics
		ExternalAPIRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "external_api_request_duration_seconds",
				Help:      "External API request duration in seconds",
				Buckets:   []float64{.01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"endpoint", "status"},
		),
		ExternalAPIRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "external_api_requests_total",
				Help:      "Total number of external API requests",
			},
			[]string{"endpoint", "method", "status"},
		),
		ExternalAPIErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "external_api_errors_total",
				Help:      "Total number of external API errors",
			},
			[]string{"endpoint", "error_type"},
		),

		// Business metrics
		ProjectsTotal: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "projects_total",
				Help:      "Total number of active projects",
			},
		),
		BoardsTotal: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "boards_total",
				Help:      "Total number of boards",
			},
		),
		ProjectCreatedTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "project_created_total",
				Help:      "Total number of project creation events",
			},
		),
		BoardCreatedTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "board_created_total",
				Help:      "Total number of board creation events",
			},
		),
	}
}
