package main

import (
	"context"
	"database/sql"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/YHQZ1/esx/packages/kafka"
	"github.com/YHQZ1/esx/packages/logger"
	"github.com/YHQZ1/esx/services/settlement-engine/internal/db"
	consumer "github.com/YHQZ1/esx/services/settlement-engine/internal/kafka"
	"github.com/YHQZ1/esx/services/settlement-engine/internal/settlement"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()

	log := logger.New("settlement-engine")

	settlementDB, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("failed to connect to settlement database", err)
	}
	defer settlementDB.Close()

	if err := settlementDB.Ping(); err != nil {
		log.Fatal("failed to ping settlement database", err)
	}

	participantDB, err := sql.Open("postgres", os.Getenv("PARTICIPANT_DB_URL"))
	if err != nil {
		log.Fatal("failed to connect to participant database", err)
	}
	defer participantDB.Close()

	if err := participantDB.Ping(); err != nil {
		log.Fatal("failed to ping participant database", err)
	}

	riskDB, err := sql.Open("postgres", os.Getenv("RISK_DB_URL"))
	if err != nil {
		log.Fatal("failed to connect to risk database", err)
	}
	defer riskDB.Close()

	if err := riskDB.Ping(); err != nil {
		log.Fatal("failed to ping risk database", err)
	}

	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")

	producer := kafka.NewProducer(brokers, kafka.TopicTradeSettled, log)
	defer producer.Close()

	queries := db.New(settlementDB, participantDB, riskDB)
	settler := settlement.New(queries, log)
	h := consumer.New(settler, producer, log)

	c := kafka.NewConsumer(brokers, kafka.TopicTradeCleared, "settlement-engine", log)
	c.RegisterHandler(h.Handle)
	defer c.Close()

	log.Info("settlement engine started")

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := c.Start(ctx); err != nil {
		log.Fatal("consumer error", err)
	}
}
