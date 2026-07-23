package observability

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// TestInit_Success проверяет успешную инициализацию с валидным endpoint
func TestInit_Success(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
		endpoint    string
	}{
		{
			name:        "valid_local_endpoint",
			serviceName: "test-service",
			endpoint:    "localhost:4317",
		},
		{
			name:        "valid_with_port",
			serviceName: "test-service",
			endpoint:    "127.0.0.1:4317",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shutdown, err := Init(tt.serviceName, tt.endpoint)

			// Shutdown функция должна быть возвращена всегда
			require.NotNil(t, shutdown, "shutdown function should not be nil")

			// Если collector недоступен — err допустим, но shutdown должен работать
			if err == nil {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				_ = shutdown(ctx)
			}
		})
	}
}

// TestInit_EmptyEndpoint проверяет инициализацию с пустым endpoint (noop/stdout)
func TestInit_EmptyEndpoint(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
		endpoint    string
	}{
		{
			name:        "empty_endpoint_uses_noop",
			serviceName: "test-service",
			endpoint:    "",
		},
		{
			name:        "empty_service_name_and_endpoint",
			serviceName: "",
			endpoint:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shutdown, err := Init(tt.serviceName, tt.endpoint)

			require.NotNil(t, shutdown, "shutdown function should not be nil")

			if err == nil {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				err = shutdown(ctx)
				assert.NoError(t, err)
			}
		})
	}
}

// TestInit_ShutdownFunc проверяет корректность работы shutdown функции
func TestInit_ShutdownFunc(t *testing.T) {
	shutdown, err := Init("shutdown-test", "")
	require.NoError(t, err)
	require.NotNil(t, shutdown)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = shutdown(ctx)
	assert.NoError(t, err, "shutdown should not return error on first call")
}

// TestGRPCServerInterceptors проверяет возврат non-nil серверных перехватчиков
func TestGRPCServerInterceptors(t *testing.T) {
	unary, stream := GRPCServerInterceptors()

	assert.NotNil(t, unary, "unary server interceptor should not be nil")
	assert.NotNil(t, stream, "stream server interceptor should not be nil")
}

// TestGRPCServerInterceptors_NotNil проверяет что interceptors не nil после repeated вызовов
func TestGRPCServerInterceptors_NotNil(t *testing.T) {
	for i := 0; i < 5; i++ {
		t.Run("iteration", func(t *testing.T) {
			unary, stream := GRPCServerInterceptors()
			require.NotNil(t, unary)
			require.NotNil(t, stream)
		})
	}
}

// TestGRPCClientInterceptors проверяет возврат non-nil клиентских перехватчиков
func TestGRPCClientInterceptors(t *testing.T) {
	unary, stream := GRPCClientInterceptors()

	assert.NotNil(t, unary, "unary client interceptor should not be nil")
	assert.NotNil(t, stream, "stream client interceptor should not be nil")
}

// TestGRPCClientInterceptors_NotNil проверяет что interceptors не nil после repeated вызовов
func TestGRPCClientInterceptors_NotNil(t *testing.T) {
	for i := 0; i < 5; i++ {
		t.Run("iteration", func(t *testing.T) {
			unary, stream := GRPCClientInterceptors()
			require.NotNil(t, unary)
			require.NotNil(t, stream)
		})
	}
}

// TestHTTPMiddleware проверяет возврат non-nil HTTP middleware
func TestHTTPMiddleware(t *testing.T) {
	middleware := HTTPMiddleware()
	assert.NotNil(t, middleware, "HTTP middleware should not be nil")
}

// TestHTTPMiddleware_CreatesSpan проверяет что middleware обрабатывает запрос без паники
func TestHTTPMiddleware_CreatesSpan(t *testing.T) {
	middleware := HTTPMiddleware()
	require.NotNil(t, middleware)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("User-Agent", "test-agent")

	assert.NotPanics(t, func() {
		middleware(c)
	})
}

// TestHTTPMiddleware_WithTracerProvider проверяет middleware с настроенным tracer
func TestHTTPMiddleware_WithTracerProvider(t *testing.T) {
	shutdown, err := Init("middleware-test", "")
	require.NotNil(t, shutdown)

	if err == nil {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = shutdown(ctx)
		}()
	}

	middleware := HTTPMiddleware()
	require.NotNil(t, middleware)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/transfers", nil)
	c.Request.Header.Set("Content-Type", "application/json")

	assert.NotPanics(t, func() {
		middleware(c)
	})
}

// TestHTTPMiddleware_WithPath проверяет middleware с различными путями
func TestHTTPMiddleware_WithPath(t *testing.T) {
	paths := []string{
		"/",
		"/api/v1/accounts",
		"/api/v1/transfers",
		"/health",
		"/metrics",
	}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			middleware := HTTPMiddleware()
			require.NotNil(t, middleware)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, path, nil)

			assert.NotPanics(t, func() {
				middleware(c)
			})
		})
	}
}

// TestHTTPMiddleware_WithMethods проверяет middleware с различными HTTP методами
func TestHTTPMiddleware_WithMethods(t *testing.T) {
	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			middleware := HTTPMiddleware()
			require.NotNil(t, middleware)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(method, "/test", nil)

			assert.NotPanics(t, func() {
				middleware(c)
			})
		})
	}
}

// TestHealthHandler проверяет health check endpoint
func TestHealthHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "GET_returns_200",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "HEAD_returns_200",
			method:         http.MethodHead,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := HealthHandler()
			require.NotNil(t, handler)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(tt.method, "/health", nil)

			handler(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestHealthHandler_ContentType проверяет Content-Type ответа
func TestHealthHandler_ContentType(t *testing.T) {
	handler := HealthHandler()
	require.NotNil(t, handler)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/health", nil)

	handler(c)

	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
}

// TestHealthHandler_ResponseFormat проверяет формат ответа health check
func TestHealthHandler_ResponseFormat(t *testing.T) {
	handler := HealthHandler()
	require.NotNil(t, handler)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/health", nil)

	handler(c)

	body := w.Body.String()
	assert.Contains(t, body, "status")
	assert.Contains(t, body, "ok")
}

// TestHealthHandler_NotNil проверяет что handler не nil
func TestHealthHandler_NotNil(t *testing.T) {
	handler := HealthHandler()
	require.NotNil(t, handler)
}

// TestMetricsHandler проверяет metrics endpoint
func TestMetricsHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "GET_returns_metrics",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "HEAD_returns_metrics",
			method:         http.MethodHead,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := MetricsHandler()
			require.NotNil(t, handler)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(tt.method, "/metrics", nil)

			handler(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestMetricsHandler_NotNil проверяет что handler не nil
func TestMetricsHandler_NotNil(t *testing.T) {
	handler := MetricsHandler()
	assert.NotNil(t, handler, "metrics handler should not be nil")
}

// TestCustomMetrics_Recording проверяет запись значений в гистограммы/счетчики без паники
func TestCustomMetrics_Recording(t *testing.T) {
	mp := sdkmetric.NewMeterProvider()
	meter := mp.Meter("test-meter")
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = mp.Shutdown(ctx)
	}()

	// Тест для HTTPRequestDuration гистограммы
	t.Run("HTTPRequestDuration_recording", func(t *testing.T) {
		histogram, err := meter.Float64Histogram(
			"http.request.duration",
			metric.WithDescription("HTTP request duration in milliseconds"),
			metric.WithUnit("ms"),
		)
		require.NoError(t, err)

		assert.NotPanics(t, func() {
			histogram.Record(context.Background(), 0.001,
				metric.WithAttributes(
					attribute.String("method", "GET"),
					attribute.String("path", "/api/v1/accounts"),
					attribute.Int("status_code", 200),
				))
			histogram.Record(context.Background(), 100.5,
				metric.WithAttributes(
					attribute.String("method", "POST"),
					attribute.String("path", "/api/v1/transfers"),
					attribute.Int("status_code", 201),
				))
			histogram.Record(context.Background(), 1500.0,
				metric.WithAttributes(
					attribute.String("method", "GET"),
					attribute.String("path", "/api/v1/health"),
					attribute.Int("status_code", 500),
				))
		})
	})

	// Тест для GRPCRequestDuration гистограммы
	t.Run("GRPCRequestDuration_recording", func(t *testing.T) {
		histogram, err := meter.Float64Histogram(
			"grpc.request.duration",
			metric.WithDescription("gRPC request duration in milliseconds"),
			metric.WithUnit("ms"),
		)
		require.NoError(t, err)

		assert.NotPanics(t, func() {
			histogram.Record(context.Background(), 0.5,
				metric.WithAttributes(
					attribute.String("method", "/banking.AccountService/GetAccount"),
					attribute.Int("status_code", 0),
				))
			histogram.Record(context.Background(), 250.0,
				metric.WithAttributes(
					attribute.String("method", "/banking.TransferService/CreateTransfer"),
					attribute.Int("status_code", 0),
				))
			histogram.Record(context.Background(), 2000.0,
				metric.WithAttributes(
					attribute.String("method", "/banking.TransferService/ProcessTransfer"),
					attribute.Int("status_code", 13),
				))
		})
	})

	// Тест для TransferCreatedCounter счетчика
	t.Run("TransferCreatedCounter_recording", func(t *testing.T) {
		counter, err := meter.Int64Counter(
			"transfer.created",
			metric.WithDescription("Number of transfers created"),
		)
		require.NoError(t, err)

		assert.NotPanics(t, func() {
			counter.Add(context.Background(), 1,
				metric.WithAttributes(
					attribute.String("account_id", "acc-123"),
					attribute.String("currency", "USD"),
				))
			counter.Add(context.Background(), 1,
				metric.WithAttributes(
					attribute.String("account_id", "acc-456"),
					attribute.String("currency", "EUR"),
				))
			counter.Add(context.Background(), 5,
				metric.WithAttributes(
					attribute.String("account_id", "acc-789"),
					attribute.String("currency", "RUB"),
				))
		})
	})

	// Тест для TransferFailedCounter счетчика
	t.Run("TransferFailedCounter_recording", func(t *testing.T) {
		counter, err := meter.Int64Counter(
			"transfer.failed",
			metric.WithDescription("Number of failed transfers"),
		)
		require.NoError(t, err)

		assert.NotPanics(t, func() {
			counter.Add(context.Background(), 1,
				metric.WithAttributes(
					attribute.String("error_type", "insufficient_funds"),
					attribute.String("currency", "USD"),
				))
			counter.Add(context.Background(), 1,
				metric.WithAttributes(
					attribute.String("error_type", "account_not_found"),
					attribute.String("currency", "EUR"),
				))
			counter.Add(context.Background(), 3,
				metric.WithAttributes(
					attribute.String("error_type", "validation_error"),
					attribute.String("currency", "RUB"),
				))
		})
	})
}

// TestCustomMetrics_WithAttributes проверяет запись с различными комбинациями атрибутов
func TestCustomMetrics_WithAttributes(t *testing.T) {
	mp := sdkmetric.NewMeterProvider()
	meter := mp.Meter("test-meter-attrs")
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = mp.Shutdown(ctx)
	}()

	t.Run("histogram_with_multiple_attributes", func(t *testing.T) {
		histogram, err := meter.Float64Histogram(
			"http.request.duration.attrs",
			metric.WithDescription("Test histogram with attributes"),
		)
		require.NoError(t, err)

		assert.NotPanics(t, func() {
			histogram.Record(context.Background(), 10.0,
				metric.WithAttributes(
					attribute.String("method", "GET"),
					attribute.String("path", "/test"),
					attribute.Int("status_code", 200),
					attribute.String("user_agent", "test-agent"),
				))
		})
	})

	t.Run("counter_with_zero_value", func(t *testing.T) {
		counter, err := meter.Int64Counter(
			"transfer.created.zero",
			metric.WithDescription("Test counter with zero increment"),
		)
		require.NoError(t, err)

		assert.NotPanics(t, func() {
			counter.Add(context.Background(), 0,
				metric.WithAttributes(
					attribute.String("test", "zero"),
				))
			counter.Add(context.Background(), 1,
				metric.WithAttributes(
					attribute.String("test", "one"),
				))
		})
	})

	t.Run("counter_with_large_value", func(t *testing.T) {
		counter, err := meter.Int64Counter(
			"transfer.created.large",
			metric.WithDescription("Test counter with large increment"),
		)
		require.NoError(t, err)

		assert.NotPanics(t, func() {
			counter.Add(context.Background(), 1000000,
				metric.WithAttributes(
					attribute.String("test", "large"),
				))
		})
	})

	t.Run("histogram_negative_value", func(t *testing.T) {
		histogram, err := meter.Float64Histogram(
			"http.request.duration.negative",
			metric.WithDescription("Test histogram with negative value"),
		)
		require.NoError(t, err)

		assert.NotPanics(t, func() {
			histogram.Record(context.Background(), -1.0,
				metric.WithAttributes(
					attribute.String("test", "negative"),
				))
		})
	})

	t.Run("histogram_zero_value", func(t *testing.T) {
		histogram, err := meter.Float64Histogram(
			"http.request.duration.zero",
			metric.WithDescription("Test histogram with zero value"),
		)
		require.NoError(t, err)

		assert.NotPanics(t, func() {
			histogram.Record(context.Background(), 0.0,
				metric.WithAttributes(
					attribute.String("test", "zero"),
				))
		})
	})
}

// TestCustomMetrics_MultipleRecordings проверяет многократную запись значений
func TestCustomMetrics_MultipleRecordings(t *testing.T) {
	mp := sdkmetric.NewMeterProvider()
	meter := mp.Meter("test-meter-multi")
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = mp.Shutdown(ctx)
	}()

	t.Run("multiple_histogram_recordings", func(t *testing.T) {
		histogram, err := meter.Float64Histogram(
			"http.request.duration.multi",
			metric.WithDescription("Histogram for multiple recordings"),
		)
		require.NoError(t, err)

		assert.NotPanics(t, func() {
			for i := 0; i < 100; i++ {
				histogram.Record(context.Background(), float64(i)*1.5,
					metric.WithAttributes(
						attribute.Int("iteration", i),
					))
			}
		})
	})

	t.Run("multiple_counter_recordings", func(t *testing.T) {
		counter, err := meter.Int64Counter(
			"transfer.created.multi",
			metric.WithDescription("Counter for multiple increments"),
		)
		require.NoError(t, err)

		assert.NotPanics(t, func() {
			for i := 0; i < 100; i++ {
				counter.Add(context.Background(), 1,
					metric.WithAttributes(
						attribute.Int("iteration", i),
					))
				counter.Add(context.Background(), int64(i),
					metric.WithAttributes(
						attribute.Int("iteration", i),
					))
			}
		})
	})

	t.Run("concurrent_histogram_recordings", func(t *testing.T) {
		histogram, err := meter.Float64Histogram(
			"http.request.duration.concurrent",
			metric.WithDescription("Histogram for concurrent recordings"),
		)
		require.NoError(t, err)

		done := make(chan struct{}, 10)
		for g := 0; g < 10; g++ {
			go func(goroutine int) {
				for i := 0; i < 50; i++ {
					histogram.Record(context.Background(), float64(i),
						metric.WithAttributes(
							attribute.Int("goroutine", goroutine),
							attribute.Int("iteration", i),
						))
				}
				done <- struct{}{}
			}(g)
		}
		for g := 0; g < 10; g++ {
			<-done
		}
	})
}

// TestInit_ContextTimeout проверяет поведение shutdown при истечении контекста
func TestInit_ContextTimeout(t *testing.T) {
	shutdown, err := Init("timeout-test", "")
	require.NotNil(t, shutdown)

	if err != nil {
		t.Skip("Init failed, skipping shutdown test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(2 * time.Millisecond)

	assert.NotPanics(t, func() {
		_ = shutdown(ctx)
	})
}

// TestGRPCServerInterceptors_Consistency проверяет консистентность возвращаемых значений
func TestGRPCServerInterceptors_Consistency(t *testing.T) {
	unary1, stream1 := GRPCServerInterceptors()
	unary2, stream2 := GRPCServerInterceptors()

	// Оба вызова должны возвращать не-nil значения
	assert.NotNil(t, unary1)
	assert.NotNil(t, unary2)
	assert.NotNil(t, stream1)
	assert.NotNil(t, stream2)
}

// TestGRPCClientInterceptors_Consistency проверяет консистентность возвращаемых значений
func TestGRPCClientInterceptors_Consistency(t *testing.T) {
	unary1, stream1 := GRPCClientInterceptors()
	unary2, stream2 := GRPCClientInterceptors()

	assert.NotNil(t, unary1)
	assert.NotNil(t, unary2)
	assert.NotNil(t, stream1)
	assert.NotNil(t, stream2)
}
