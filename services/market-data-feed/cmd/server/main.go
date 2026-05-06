package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/YHQZ1/esx/packages/kafka"
	"github.com/YHQZ1/esx/packages/logger"
	consumer "github.com/YHQZ1/esx/services/market-data-feed/internal/kafka"
	"github.com/YHQZ1/esx/services/market-data-feed/internal/ws"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // restrict in production
	},
}

func main() {
	godotenv.Load()

	log := logger.New("market-data-feed")

	hub := ws.NewHub(log)
	go hub.Run()

	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	h := consumer.New(hub, log)

	tradeConsumer := kafka.NewConsumer(brokers, kafka.TopicTradeExecuted, "market-data-feed", log)
	tradeConsumer.RegisterHandler(h.HandleTradeExecuted)
	defer tradeConsumer.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go func() {
		if err := tradeConsumer.Start(ctx); err != nil {
			log.Fatal("consumer error", err)
		}
	}()

	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Error("websocket upgrade failed", err)
			return
		}
		client := hub.Register(conn)
		go client.WritePump(hub)
		client.ReadPump(hub)
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8085"
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Info("market data feed started", logger.Str("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("http server failed", err)
		}
	}()

	<-ctx.Done()
	log.Info("shutting down market data feed")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("server forced to shutdown", err)
	}

	log.Info("market data feed stopped")
}
