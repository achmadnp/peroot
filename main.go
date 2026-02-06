package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/achmadnp/peroot/config"
	"github.com/achmadnp/peroot/handler"
	"github.com/achmadnp/peroot/middleware"
	"github.com/achmadnp/peroot/service"
	"github.com/gin-gonic/gin"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg := config.Load()

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	whatsAppService := service.NewWhatsAppService(cfg)
	orderHandler := handler.NewOrderHandler(whatsAppService)

	router := gin.Default()

	router.GET("/health", orderHandler.HealthCheck)

	v1 := router.Group("/api/v1")
	v1.Use(middleware.APIKeyAuth(cfg.APIKey))
	{
		v1.POST("/orders/notify", orderHandler.NotifyOrder)
	}

	slog.Info("Starting peroot server", "port", cfg.AppPort, "env", cfg.AppEnv)

	if err := router.Run(cfg.AppPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
