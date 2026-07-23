package observability

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var Metrics = struct {
	HTTPRequestDuration  prometheus.Histogram
	HTTPRequestsTotal    prometheus.Counter
	GRPCRequestDuration  prometheus.Histogram
	GRPCRequestsTotal    prometheus.Counter
	TransferCreatedTotal prometheus.Counter
	TransferFailedTotal  prometheus.Counter
	DBQueryDuration      prometheus.Histogram
}{
	HTTPRequestDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Duration of HTTP requests in seconds",
		Buckets: prometheus.DefBuckets,
	}),
	HTTPRequestsTotal: prometheus.NewCounter(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	}),
	GRPCRequestDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "grpc_request_duration_seconds",
		Help:    "Duration of gRPC requests in seconds",
		Buckets: prometheus.DefBuckets,
	}),
	GRPCRequestsTotal: prometheus.NewCounter(prometheus.CounterOpts{
		Name: "grpc_requests_total",
		Help: "Total number of gRPC requests",
	}),
	TransferCreatedTotal: prometheus.NewCounter(prometheus.CounterOpts{
		Name: "transfer_created_total",
		Help: "Total number of transfers created",
	}),
	TransferFailedTotal: prometheus.NewCounter(prometheus.CounterOpts{
		Name: "transfer_failed_total",
		Help: "Total number of failed transfers",
	}),
	DBQueryDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "db_query_duration_seconds",
		Help:    "Duration of database queries in seconds",
		Buckets: prometheus.DefBuckets,
	}),
}

var registerOnce sync.Once

func RegisterMetrics() {
	registerOnce.Do(func() {
		prometheus.MustRegister(
			Metrics.HTTPRequestDuration,
			Metrics.HTTPRequestsTotal,
			Metrics.GRPCRequestDuration,
			Metrics.GRPCRequestsTotal,
			Metrics.TransferCreatedTotal,
			Metrics.TransferFailedTotal,
			Metrics.DBQueryDuration,
		)
	})
}
