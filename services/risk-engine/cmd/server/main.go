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

	log := logger.New("risk-engine")

	riskDB, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("failed to connect to risk database", err)
	}
	defer riskDB.Close()
	pingWithRetry(riskDB, "risk-engine", log)

	participantDB, err := sql.Open("postgres", os.Getenv("PARTICIPANT_DB_URL"))
	if err != nil {
		log.Fatal("failed to connect to participant database", err)
	}
	defer participantDB.Close()
	pingWithRetry(participantDB, "participant-registry", log)

	riskDB.SetMaxOpenConns(500)
	riskDB.SetMaxIdleConns(500)
	riskDB.SetConnMaxLifetime(5 * time.Minute)

	participantDB.SetMaxOpenConns(500)
	participantDB.SetMaxIdleConns(500)
	participantDB.SetConnMaxLifetime(5 * time.Minute)

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	queries := db.New(riskDB, participantDB)
	checker := checks.New(queries)
	locker := locks.New(queries, rdb, log)
	srv := grpcserver.NewServer(checker, locker, log)

	lis, err := net.Listen("tcp", ":9093")
	if err != nil {
		log.Fatal("failed to listen", err)
	}

	s := grpc.NewServer()
	pb.RegisterRiskServiceServer(s, srv)
	reflection.Register(s)

	log.Info("grpc server listening", logger.Str("addr", ":9093"))

	if err := s.Serve(lis); err != nil {
		log.Fatal("grpc server failed", err)
	}
}
