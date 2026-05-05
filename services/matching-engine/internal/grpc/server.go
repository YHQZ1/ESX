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
	queries db.Querier
	engine  *matching.Engine
	log     *logger.Logger
	// --- ADD CHANNEL ---
	orderChan chan db.Order
}

func NewServer(queries db.Querier, engine *matching.Engine, log *logger.Logger) *Server {
	s := &Server{
		queries:   queries,
		engine:    engine,
		log:       log,
		orderChan: make(chan db.Order, 50000), // Buffer for 50k orders
	}

	// Start the single-threaded Matching Worker
	go s.runMatcher()

	return s
}

func (s *Server) runMatcher() {
	for order := range s.orderChan {
		// Process order in RAM without locks
		s.engine.Submit(context.Background(), order)
	}
}

func (s *Server) SubmitOrder(ctx context.Context, req *pb.SubmitOrderRequest) (*pb.SubmitOrderResponse, error) {
	participantID, _ := uuid.Parse(req.ParticipantId)
	lockID, _ := uuid.Parse(req.LockId)
	orderID := uuid.New()

	order := db.Order{
		ID:            orderID,
		ParticipantID: participantID,
		Symbol:        req.Symbol,
		Side:          strings.TrimPrefix(req.Side.String(), "ORDER_SIDE_"),
		Type:          strings.TrimPrefix(req.Type.String(), "ORDER_TYPE_"),
		Quantity:      req.Quantity,
		Price:         req.Price,
		LockID:        lockID,
	}

	// Non-blocking: drop into pipeline and return instantly
	select {
	case s.orderChan <- order:
		return &pb.SubmitOrderResponse{OrderId: orderID.String(), Status: "queued"}, nil
	default:
		return nil, status.Error(codes.ResourceExhausted, "buffer full")
	}
}

// ... keep CancelOrder as is

func (s *Server) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.CancelOrderResponse, error) {
	orderID, err := uuid.Parse(req.OrderId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid order_id")
	}

	participantID, err := uuid.Parse(req.ParticipantId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid participant_id")
	}

	_, err = s.queries.CancelOrder(ctx, orderID, participantID)
	if err != nil {
		return &pb.CancelOrderResponse{Cancelled: false, Reason: "order not found or already closed"}, nil
	}

	s.log.Info("order cancelled", logger.Str("order_id", req.OrderId))
	return &pb.CancelOrderResponse{Cancelled: true}, nil
}
