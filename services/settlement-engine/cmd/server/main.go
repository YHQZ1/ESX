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
	"github.com/YHQZ1/esx/services/settlement-engine/internal/db"
	consumer "github.com/YHQZ1/esx/services/settlement-engine/internal/kafka"
	"github.com/YHQZ1/esx/services/settlement-engine/internal/settlement"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func pingWithRetry(database *sql.DB, name string, log *logger.Logger) {
	for i := range 5 {
		if err := database.Ping(); err == nil {
			return
		} else if i == 4 {
			log.Fatal("failed to ping "+name+" after retries", err)
		} else {
			log.Warn("database not ready, retrying",
				logger.Str("db", name),
				logger.Int("attempt", i+1),
			)
			time.Sleep(2 * time.Second)
		}
	}
}

func main() {
	godotenv.Load()

	log := logger.New("settlement-engine")

	settlementDB, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("failed to connect to settlement database", err)
	}
	defer settlementDB.Close()
	pingWithRetry(settlementDB, "settlement-engine", log)

	participantDB, err := sql.Open("postgres", os.Getenv("PARTICIPANT_DB_URL"))
	if err != nil {
		log.Fatal("failed to connect to participant database", err)
	}
	defer participantDB.Close()
	pingWithRetry(participantDB, "participant-registry", log)

	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")

	producer := kafka.NewProducer(brokers, kafka.TopicTradeSettled, log)
	defer producer.Close()

	queries := db.New(settlementDB, participantDB)
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
