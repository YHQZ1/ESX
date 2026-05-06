package main

import (
	"database/sql"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/YHQZ1/esx/packages/logger"
	pb "github.com/YHQZ1/esx/packages/proto/participant"
	"github.com/YHQZ1/esx/services/participant-registry/internal/db"
	grpcserver "github.com/YHQZ1/esx/services/participant-registry/internal/grpc"
	"github.com/YHQZ1/esx/services/participant-registry/internal/handlers"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

func main() {
	godotenv.Load()

	log := logger.New("participant-registry")

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

	queries := db.New(database)
	h := handlers.New(queries, log)
	grpcSrv := grpcserver.NewServer(queries, log)

	go func() {
		lis, err := net.Listen("tcp", ":9091")
		if err != nil {
			log.Fatal("failed to listen for grpc", err)
		}
		s := grpc.NewServer()
		pb.RegisterParticipantServiceServer(s, grpcSrv)
		log.Info("grpc server listening", logger.Str("addr", ":9091"))
		if err := s.Serve(lis); err != nil {
			log.Fatal("grpc server failed", err)
		}
	}()

	r := gin.New()
	r.Use(gin.Recovery())

	r.POST("/participants/register", h.Register)
	r.POST("/participants/:id/deposit", h.Deposit)
	r.GET("/participants/:id", h.GetAccount)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Info("http server listening", logger.Str("port", port))
	if err := r.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatal("http server failed", err)
	}
}
