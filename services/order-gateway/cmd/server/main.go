package main

import (
	"fmt"
	"os"

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

	log.Info("order gateway started", logger.Str("port", port))

	if err := r.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatal("http server failed", err)
	}
}
