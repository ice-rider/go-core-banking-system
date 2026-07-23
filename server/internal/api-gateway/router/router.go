package router

import (
	"github.com/gin-gonic/gin"

	"go-core-banking-system/internal/api-gateway/handler"
	"go-core-banking-system/internal/api-gateway/middleware"
	"go-core-banking-system/pkg/observability"
)

func NewRouter(h *handler.Handler) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.LoggingMiddleware())
	r.Use(middleware.CORSMiddleware())
	r.Use(observability.HTTPMiddleware())

	r.GET("/healthz", h.HealthCheck)
	r.GET("/readyz", h.HealthCheck)
	r.GET("/metrics", observability.MetricsHandler())

	r.POST("/accounts", h.CreateAccount)
	r.GET("/accounts/:id", h.GetAccount)
	r.POST("/accounts/:id/block", h.BlockAccount)
	r.POST("/accounts/:id/unblock", h.UnblockAccount)
	r.POST("/accounts/:id/close", h.CloseAccount)

	r.POST("/transfers", h.CreateTransfer)
	r.GET("/transfers/:id", h.GetTransaction)

	return r
}
