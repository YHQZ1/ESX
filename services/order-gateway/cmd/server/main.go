package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/YHQZ1/esx/packages/logger"
	"github.com/YHQZ1/esx/services/order-gateway/internal/client"
	"github.com/YHQZ1/esx/services/order-gateway/internal/handlers"
	"github.com/YHQZ1/esx/services/order-gateway/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	log := logger.New("order-gateway")

	// Initialize gRPC clients with high-performance connection settings
	registryClient, err := client.NewRegistryClient(os.Getenv("PARTICIPANT_REGISTRY_ADDR"))
	if err != nil {
		log.Fatal("failed to connect to participant registry", err)
	}

	riskClient, err := client.NewRiskClient(os.Getenv("RISK_ENGINE_ADDR"))
	if err != nil {
		log.Fatal("failed to connect to risk engine", err)
	}

	matchingClient, err := client.NewMatchingClient(os.Getenv("MATCHING_ENGINE_ADDR"))
	if err != nil {
		log.Fatal("failed to connect to matching engine", err)
	}

	orderHandler := handlers.NewOrderHandler(riskClient, matchingClient, log)
	fixHandler := handlers.NewFIXHandler(registryClient, riskClient, matchingClient, log)

	// Set Gin to release mode for maximum performance metrics
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())

	r.POST("/fix", fixHandler.Handle)

	auth := r.Group("/")
	auth.Use(middleware.Auth(registryClient, log))
	{
		auth.POST("/orders", orderHandler.SubmitOrder)
		auth.DELETE("/orders/:id", orderHandler.CancelOrder)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// ENTERPRISE HTTP SERVER TUNING
	// Standard r.Run() is too slow for 10k RPS. We use a custom http.Server.
	srv := &http.Server{
		Addr:           fmt.Sprintf(":%s", port),
		Handler:        r,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	log.Info("order gateway started with high-concurrency tuning", logger.Str("port", port))

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("http server failed", err)
	}
}
