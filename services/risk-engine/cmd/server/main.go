package main

import (
	"database/sql"
	"net"
	"os"
	"time"

	"github.com/YHQZ1/esx/packages/logger"
	pb "github.com/YHQZ1/esx/packages/proto/risk"
	"github.com/YHQZ1/esx/services/risk-engine/internal/checks"
	"github.com/YHQZ1/esx/services/risk-engine/internal/db"
	grpcserver "github.com/YHQZ1/esx/services/risk-engine/internal/grpc"
	"github.com/YHQZ1/esx/services/risk-engine/internal/locks"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	godotenv.Load()

	log := logger.New("risk-engine")

	riskDB, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("failed to connect to risk database", err)
	}
	defer riskDB.Close()

	if err := riskDB.Ping(); err != nil {
		log.Fatal("failed to ping risk database", err)
	}

	participantDB, err := sql.Open("postgres", os.Getenv("PARTICIPANT_DB_URL"))
	if err != nil {
		log.Fatal("failed to connect to participant database", err)
	}
	defer participantDB.Close()

	if err := participantDB.Ping(); err != nil {
		log.Fatal("failed to ping participant database", err)
	}

	riskDB.SetMaxOpenConns(50)
	riskDB.SetMaxIdleConns(50)
	riskDB.SetConnMaxLifetime(5 * time.Minute)

	participantDB.SetMaxOpenConns(50)
	participantDB.SetMaxIdleConns(50)
	participantDB.SetConnMaxLifetime(5 * time.Minute)

	queries := db.New(riskDB, participantDB)
	checker := checks.New(queries)
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	locker := locks.New(queries, rdb)
	srv := grpcserver.NewServer(checker, locker, log)

	lis, err := net.Listen("tcp", ":9093")
	if err != nil {
		log.Fatal("failed to listen", err)
	}

	s := grpc.NewServer()
	pb.RegisterRiskServiceServer(s, srv)
	reflection.Register(s)

	log.Info("grpc server listening", logger.Str("addr", ":9092"))

	if err := s.Serve(lis); err != nil {
		log.Fatal("grpc server failed", err)
	}
}
