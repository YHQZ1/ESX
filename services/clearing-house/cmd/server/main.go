package main

import (
	"context"
	"database/sql"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/YHQZ1/esx/packages/kafka"
	"github.com/YHQZ1/esx/packages/logger"
	"github.com/YHQZ1/esx/services/clearing-house/internal/db"
	consumer "github.com/YHQZ1/esx/services/clearing-house/internal/kafka"
	"github.com/YHQZ1/esx/services/clearing-house/internal/novation"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()

	log := logger.New("clearing-house")

	clearingDB, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("failed to connect to clearing database", err)
	}
	defer clearingDB.Close()

	for i := range 5 {
		if err := clearingDB.Ping(); err == nil {
			break
		} else if i == 4 {
			log.Fatal("failed to ping clearing database after retries", err)
		} else {
			log.Warn("clearing database not ready, retrying",
				logger.Int("attempt", i+1),
			)
			time.Sleep(2 * time.Second)
		}
	}

	riskDB, err := sql.Open("postgres", os.Getenv("RISK_DB_URL"))

	for i := range 5 {
		if err := riskDB.Ping(); err == nil {
			break
		} else if i == 4 {
			log.Fatal("failed to ping risk database after retries", err)
		} else {
			log.Warn("risk database not ready, retrying",
				logger.Int("attempt", i+1),
			)
			time.Sleep(2 * time.Second)
		}
	}

	defer riskDB.Close()

	if err := riskDB.Ping(); err != nil {
		log.Fatal("failed to ping risk database", err)
	}

	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")

	producer := kafka.NewProducer(brokers, kafka.TopicTradeCleared, log)
	defer producer.Close()

	queries := db.New(clearingDB, riskDB)
	novator := novation.New(queries, log)
	h := consumer.New(novator, producer, log)

	c := kafka.NewConsumer(brokers, kafka.TopicTradeExecuted, "clearing-house", log)
	c.RegisterHandler(h.Handle)
	defer c.Close()

	log.Info("clearing house started")

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := c.Start(ctx); err != nil {
		log.Fatal("consumer error", err)
	}
}
