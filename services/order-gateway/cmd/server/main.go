package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	srv := &http.Server{
		Addr:           fmt.Sprintf(":%s", port),
		Handler:        r,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		log.Info("order gateway started", logger.Str("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("http server failed", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down order gateway")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("server forced to shutdown", err)
	}

	log.Info("order gateway stopped")
}
