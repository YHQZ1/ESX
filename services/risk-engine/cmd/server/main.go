package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	riskpb "github.com/YHQZ1/esx/packages/proto/risk"
	"github.com/YHQZ1/esx/services/risk-engine/internal/checks"
	"github.com/YHQZ1/esx/services/risk-engine/internal/db"
	riskgrpc "github.com/YHQZ1/esx/services/risk-engine/internal/grpc"
	"github.com/YHQZ1/esx/services/risk-engine/internal/locks"
)

func main() {
	godotenv.Load()

	log := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("service", "risk-engine").
		Logger()

	level, err := zerolog.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal().Msg("DATABASE_URL is required")
	}

	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open database connection")
	}
	defer conn.Close()

	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(10)
	conn.SetConnMaxLifetime(5 * time.Minute)

	if err := conn.Ping(); err != nil {
		log.Fatal().Err(err).Msg("failed to ping database")
	}

	log.Info().Msg("connected to database")

	database := db.New(conn)
	checker := checks.New(database)
	lockManager := locks.New(database, checker)
	server := riskgrpc.NewServer(lockManager, log)

	port := os.Getenv("PORT")
	if port == "" {
		port = "9092"
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatal().Err(err).Str("port", port).Msg("failed to listen")
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(unaryLoggingInterceptor(log)),
	)
	riskpb.RegisterRiskServiceServer(grpcServer, server)
	reflection.Register(grpcServer)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info().Str("port", port).Msg("risk engine started")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal().Err(err).Msg("grpc server failed")
		}
	}()

	<-quit
	log.Info().Msg("shutting down risk engine")
	grpcServer.GracefulStop()
}

func unaryLoggingInterceptor(log zerolog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		code := codes.OK
		if err != nil {
			if s, ok := status.FromError(err); ok {
				code = s.Code()
			}
		}

		event := log.Info()
		if err != nil {
			event = log.Warn()
		}

		event.
			Str("method", info.FullMethod).
			Str("code", code.String()).
			Dur("duration_ms", duration).
			Msg("grpc request")

		return resp, err
	}
}
