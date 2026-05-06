package main

import (
	"database/sql"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/YHQZ1/esx/packages/kafka"
	"github.com/YHQZ1/esx/packages/logger"
	pb "github.com/YHQZ1/esx/packages/proto/matching"
	pbrisk "github.com/YHQZ1/esx/packages/proto/risk"
	"github.com/YHQZ1/esx/services/matching-engine/internal/circuit"
	"github.com/YHQZ1/esx/services/matching-engine/internal/db"
	grpcserver "github.com/YHQZ1/esx/services/matching-engine/internal/grpc"
	"github.com/YHQZ1/esx/services/matching-engine/internal/matching"
	"github.com/YHQZ1/esx/services/matching-engine/internal/orderbook"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	godotenv.Load()

	log := logger.New("matching-engine")

	database, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("failed to connect to database", err)
	}
	defer database.Close()

	for i := range 5 {
		if err := database.Ping(); err == nil {
			break
		} else if i == 4 {
			log.Fatal("failed to ping database after retries", err)
		} else {
			log.Warn("database not ready, retrying",
				logger.Int("attempt", i+1),
			)
			time.Sleep(2 * time.Second)
		}
	}

	database.SetMaxOpenConns(1000)
	database.SetMaxIdleConns(1000)
	database.SetConnMaxLifetime(5 * time.Minute)

	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_URL"),
	})

	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	producer := kafka.NewProducer(brokers, kafka.TopicTradeExecuted, log)
	cbProducer := kafka.NewProducer(brokers, kafka.TopicCircuitBreakerTriggered, log)
	defer producer.Close()
	defer cbProducer.Close()

	threshold := 10.0
	if t := os.Getenv("CIRCUIT_BREAKER_THRESHOLD"); t != "" {
		if v, err := strconv.ParseFloat(t, 64); err == nil {
			threshold = v
		}
	}

	riskAddr := os.Getenv("RISK_ENGINE_ADDR")
	if riskAddr == "" {
		riskAddr = "localhost:9093"
	}
	riskConn, err := grpc.NewClient(riskAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("failed to connect to risk engine", err)
	}
	defer riskConn.Close()
	riskClient := pbrisk.NewRiskServiceClient(riskConn)

	queries := db.New(database)
	batcher := db.NewOrderBatcher(database)
	defer batcher.Stop()
	book := orderbook.New(rdb)
	breaker := circuit.New(book, cbProducer, threshold, log)
	engine := matching.New(book, queries, breaker, producer, log, riskClient)
	srv := grpcserver.NewServer(queries, batcher, engine, log)

	lis, err := net.Listen("tcp", ":9094")
	if err != nil {
		log.Fatal("failed to listen", err)
	}

	s := grpc.NewServer()
	pb.RegisterMatchingServiceServer(s, srv)
	reflection.Register(s)

	log.Info("grpc server listening", logger.Str("addr", ":9094"))

	if err := s.Serve(lis); err != nil {
		log.Fatal("grpc server failed", err)
	}
}
