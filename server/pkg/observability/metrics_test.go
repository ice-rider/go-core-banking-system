package observability

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetrics_Initialization(t *testing.T) {
	tests := []struct {
		name   string
		metric interface{}
	}{
		{"HTTPRequestDuration is non-nil", Metrics.HTTPRequestDuration},
		{"HTTPRequestsTotal is non-nil", Metrics.HTTPRequestsTotal},
		{"GRPCRequestDuration is non-nil", Metrics.GRPCRequestDuration},
		{"GRPCRequestsTotal is non-nil", Metrics.GRPCRequestsTotal},
		{"TransferCreatedTotal is non-nil", Metrics.TransferCreatedTotal},
		{"TransferFailedTotal is non-nil", Metrics.TransferFailedTotal},
		{"DBQueryDuration is non-nil", Metrics.DBQueryDuration},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.metric)
		})
	}
}

func TestHTTPRequestDuration_Observe(t *testing.T) {
	tests := []struct {
		name     string
		duration float64
	}{
		{"observe 0.001 seconds", 0.001},
		{"observe 0.01 seconds", 0.01},
		{"observe 0.1 seconds", 0.1},
		{"observe 1.0 seconds", 1.0},
		{"observe 5.0 seconds", 5.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				Metrics.HTTPRequestDuration.Observe(tt.duration)
			})
		})
	}
}

func TestHTTPRequestsTotal_Inc(t *testing.T) {
	tests := []struct {
		name      string
		increment func()
	}{
		{"single inc", func() { Metrics.HTTPRequestsTotal.Inc() }},
		{"add 1", func() { Metrics.HTTPRequestsTotal.Add(1) }},
		{"add 5", func() { Metrics.HTTPRequestsTotal.Add(5) }},
		{"add 100", func() { Metrics.HTTPRequestsTotal.Add(100) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, tt.increment)
		})
	}
}

func TestGRPCRequestDuration_Observe(t *testing.T) {
	tests := []struct {
		name     string
		duration float64
	}{
		{"observe 0.001 seconds", 0.001},
		{"observe 0.01 seconds", 0.01},
		{"observe 0.1 seconds", 0.1},
		{"observe 1.0 seconds", 1.0},
		{"observe 10.0 seconds", 10.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				Metrics.GRPCRequestDuration.Observe(tt.duration)
			})
		})
	}
}

func TestGRPCRequestsTotal_Inc(t *testing.T) {
	tests := []struct {
		name      string
		increment func()
	}{
		{"single inc", func() { Metrics.GRPCRequestsTotal.Inc() }},
		{"add 1", func() { Metrics.GRPCRequestsTotal.Add(1) }},
		{"add 10", func() { Metrics.GRPCRequestsTotal.Add(10) }},
		{"add 50", func() { Metrics.GRPCRequestsTotal.Add(50) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, tt.increment)
		})
	}
}

func TestTransferCreatedTotal_Inc(t *testing.T) {
	tests := []struct {
		name      string
		increment func()
	}{
		{"single inc", func() { Metrics.TransferCreatedTotal.Inc() }},
		{"add 1", func() { Metrics.TransferCreatedTotal.Add(1) }},
		{"add 10", func() { Metrics.TransferCreatedTotal.Add(10) }},
		{"add 100", func() { Metrics.TransferCreatedTotal.Add(100) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, tt.increment)
		})
	}
}

func TestTransferFailedTotal_Inc(t *testing.T) {
	tests := []struct {
		name      string
		increment func()
	}{
		{"single inc", func() { Metrics.TransferFailedTotal.Inc() }},
		{"add 1", func() { Metrics.TransferFailedTotal.Add(1) }},
		{"add 5", func() { Metrics.TransferFailedTotal.Add(5) }},
		{"add 25", func() { Metrics.TransferFailedTotal.Add(25) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, tt.increment)
		})
	}
}

func TestDBQueryDuration_Observe(t *testing.T) {
	tests := []struct {
		name     string
		duration float64
	}{
		{"observe 0.001 seconds", 0.001},
		{"observe 0.01 seconds", 0.01},
		{"observe 0.1 seconds", 0.1},
		{"observe 1.0 seconds", 1.0},
		{"observe 2.5 seconds", 2.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				Metrics.DBQueryDuration.Observe(tt.duration)
			})
		})
	}
}

func TestRegisterMetrics_Success(t *testing.T) {
	assert.NotPanics(t, func() {
		RegisterMetrics()
	})
}

func TestRegisterMetrics_Describe(t *testing.T) {
	RegisterMetrics()

	metrics, err := prometheus.DefaultGatherer.Gather()
	require.NoError(t, err)

	metricNames := make(map[string]bool)
	for _, m := range metrics {
		metricNames[m.GetName()] = true
	}

	assert.True(t, metricNames["http_request_duration_seconds"], "http_request_duration_seconds should be registered")
	assert.True(t, metricNames["http_requests_total"], "http_requests_total should be registered")
	assert.True(t, metricNames["grpc_request_duration_seconds"], "grpc_request_duration_seconds should be registered")
	assert.True(t, metricNames["grpc_requests_total"], "grpc_requests_total should be registered")
	assert.True(t, metricNames["transfer_created_total"], "transfer_created_total should be registered")
	assert.True(t, metricNames["transfer_failed_total"], "transfer_failed_total should be registered")
	assert.True(t, metricNames["db_query_duration_seconds"], "db_query_duration_seconds should be registered")
}
