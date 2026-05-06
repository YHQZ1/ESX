package grpc

import (
	"context"
	"strings"
	"time"

	"github.com/YHQZ1/esx/packages/logger"
	pb "github.com/YHQZ1/esx/packages/proto/matching"
	"github.com/YHQZ1/esx/services/matching-engine/internal/db"
	"github.com/YHQZ1/esx/services/matching-engine/internal/matching"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedMatchingServiceServer
	queries   db.Querier
	batcher   *db.OrderBatcher
	engine    *matching.Engine
	log       *logger.Logger
	orderChan chan db.Order
}

func NewServer(queries db.Querier, batcher *db.OrderBatcher, engine *matching.Engine, log *logger.Logger) *Server {
	s := &Server{
		queries:   queries,
		batcher:   batcher,
		engine:    engine,
		log:       log,
		orderChan: make(chan db.Order, 500000),
	}
	go s.runMatcher()
	return s
}

func (s *Server) runMatcher() {
	var count int64
	var totalTime time.Duration

	s.log.Info("matcher loop started with microsecond telemetry")

	for order := range s.orderChan {
		start := time.Now()

		// This single call triggers multiple synchronous Redis TCP hops
		s.engine.Submit(context.Background(), order)

		elapsed := time.Since(start)
		totalTime += elapsed
		count++

		// Print telemetry every 500 orders
		if count%10 == 0 {
			avgMicroseconds := totalTime.Microseconds() / 10

			// Prevent division by zero if it's too fast
			maxRps := int64(0)
			if avgMicroseconds > 0 {
				maxRps = 1000000 / avgMicroseconds
			}

			s.log.Warn("MATCHER BOTTLENECK PROOF",
				logger.Int64("orders_processed", count),
				logger.Int64("avg_time_per_order_microseconds", avgMicroseconds),
				logger.Int64("theoretical_max_rps", maxRps),
			)
			totalTime = 0 // reset for next batch
		}
	}
}

func (s *Server) SubmitOrder(ctx context.Context, req *pb.SubmitOrderRequest) (*pb.SubmitOrderResponse, error) {
	participantID, err := uuid.Parse(req.ParticipantId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid participant_id")
	}

	lockID, err := uuid.Parse(req.LockId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid lock_id")
	}

	orderSide := strings.TrimPrefix(req.Side.String(), "ORDER_SIDE_")
	orderType := strings.TrimPrefix(req.Type.String(), "ORDER_TYPE_")
	timeInForce := strings.TrimPrefix(req.TimeInForce.String(), "TIME_IN_FORCE_")

	order := db.Order{
		ID:            uuid.New(),
		ParticipantID: participantID,
		Symbol:        req.Symbol,
		Side:          orderSide,
		OrderType:     orderType,
		TimeInForce:   timeInForce,
		Quantity:      req.Quantity,
		Price:         req.Price,
		LockID:        lockID,
		Status:        "open",
	}

	// Write to Postgres async via batcher
	s.batcher.Add(order)

	select {
	case s.orderChan <- order:
		return &pb.SubmitOrderResponse{
			OrderId: order.ID.String(),
			Status:  pb.OrderStatus_ORDER_STATUS_ACCEPTED,
		}, nil
	default:
		s.log.Warn("order channel buffer full",
			logger.Str("symbol", req.Symbol),
		)
		return nil, status.Error(codes.ResourceExhausted, "order buffer full, retry")
	}
}

func (s *Server) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.CancelOrderResponse, error) {
	orderID, err := uuid.Parse(req.OrderId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid order_id")
	}

	participantID, err := uuid.Parse(req.ParticipantId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid participant_id")
	}

	order, err := s.queries.CancelOrder(ctx, orderID, participantID)
	if err != nil {
		s.log.Warn("order not found or already closed",
			logger.Str("order_id", req.OrderId),
		)
		return &pb.CancelOrderResponse{Cancelled: false, Reason: "order not found or already closed"}, nil
	}

	// Release the lock on the Risk Engine
	if req.LockId != "" {
		lockID, err := uuid.Parse(req.LockId)
		if err == nil {
			if releaseErr := s.engine.ReleaseLock(ctx, lockID, order.FilledQty); releaseErr != nil {
				s.log.Error("failed to release lock on cancel", releaseErr,
					logger.Str("order_id", req.OrderId),
					logger.Str("lock_id", req.LockId),
				)
			}
		}
	}

	s.log.Info("order cancelled",
		logger.Str("order_id", req.OrderId),
		logger.Str("participant_id", req.ParticipantId),
	)
	return &pb.CancelOrderResponse{Cancelled: true}, nil
}
