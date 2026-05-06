package grpc

import (
	"context"
	"strings"

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
	for order := range s.orderChan {
		s.engine.Submit(context.Background(), order)
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
