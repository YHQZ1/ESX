package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/YHQZ1/esx/packages/kafka"
	"github.com/YHQZ1/esx/packages/logger"
	"github.com/YHQZ1/esx/services/ledger-service/internal/db"
	"github.com/YHQZ1/esx/services/ledger-service/internal/handlers"
	"github.com/YHQZ1/esx/services/ledger-service/internal/journal"
	consumer "github.com/YHQZ1/esx/services/ledger-service/internal/kafka"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()

	log := logger.New("ledger-service")

	ledgerDB, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("failed to connect to ledger database", err)
	}
	defer ledgerDB.Close()

	if err := ledgerDB.Ping(); err != nil {
		log.Fatal("failed to ping ledger database", err)
	}

	participantDB, err := sql.Open("postgres", os.Getenv("PARTICIPANT_DB_URL"))
	if err != nil {
		log.Fatal("failed to connect to participant database", err)
	}
	defer participantDB.Close()

	if err := participantDB.Ping(); err != nil {
		log.Fatal("failed to ping participant database", err)
	}

	ledgerDB.SetMaxOpenConns(50)
	ledgerDB.SetMaxIdleConns(50)
	ledgerDB.SetConnMaxLifetime(5 * time.Minute)

	participantDB.SetMaxOpenConns(50)
	participantDB.SetMaxIdleConns(50)
	participantDB.SetConnMaxLifetime(5 * time.Minute)

	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")

	queries := db.New(ledgerDB, participantDB)
	j := journal.New(queries, log)
	h := consumer.New(j, log)

	c := kafka.NewConsumer(brokers, kafka.TopicTradeSettled, "ledger-service", log)
	c.RegisterHandler(h.Handle)

	go func() {
		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()
		if err := c.Start(ctx); err != nil {
			log.Fatal("consumer error", err)
		}
	}()

	restHandler := handlers.New(queries, log)

	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/ledger/:id/balance", restHandler.GetBalance)
	r.GET("/ledger/:id/positions", restHandler.GetPositions)
	r.GET("/ledger/:id/cash-transactions", restHandler.GetCashTransactions)
	r.GET("/ledger/:id/securities-transactions", restHandler.GetSecuritiesTransactions)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8087"
	}

	log.Info("ledger service started", logger.Str("port", port))

	if err := r.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatal("http server failed", err)
	}
}
