package metrics

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path"},
	)

	activeRequests = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_active_requests",
			Help: "Number of active HTTP requests.",
		},
		[]string{"path"},
	)

	rateLimitExceededTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "rate_limit_exceeded_total",
			Help: "Total number of rate limit exceeded responses.",
		},
	)

	clusterSizeMin = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "cluster_size_min",
			Help: "Minimum cluster size.",
		},
	)

	clusterSizeMax = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "cluster_size_max",
			Help: "Maximum cluster size.",
		},
	)

	clusterSizeAvg = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "cluster_size_avg",
			Help: "Average cluster size.",
		},
	)

	pctClustered = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "pct_clustered",
			Help: "Percent of documents with cluster_id set.",
		},
	)

	registerOnce sync.Once
)

func NewRegistry() {
	registerOnce.Do(func() {
		prometheus.MustRegister(
			httpRequestsTotal,
			httpRequestDuration,
			activeRequests,
			rateLimitExceededTotal,
			clusterSizeMin,
			clusterSizeMax,
			clusterSizeAvg,
			pctClustered,
		)
	})
}

func ObserveRequest(method, path string, statusCode int, duration time.Duration) {
	httpRequestsTotal.WithLabelValues(method, path, strconv.Itoa(statusCode)).Inc()
	httpRequestDuration.WithLabelValues(path).Observe(duration.Seconds())
}

func IncActive(path string) {
	activeRequests.WithLabelValues(path).Inc()
}

func DecActive(path string) {
	activeRequests.WithLabelValues(path).Dec()
}

func IncRateLimitExceeded() {
	rateLimitExceededTotal.Inc()
}

func SetClusterSizeStats(min, max, avg float64) {
	clusterSizeMin.Set(min)
	clusterSizeMax.Set(max)
	clusterSizeAvg.Set(avg)
}

func SetPctClustered(pct float64) {
	pctClustered.Set(pct)
}

func Handler() http.Handler {
	return promhttp.Handler()
}
