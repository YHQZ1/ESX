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
}

func NewServer(queries db.Querier, engine *matching.Engine, log *logger.Logger) *Server {
	return &Server{queries: queries, engine: engine, log: log}
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

	orderType := strings.TrimPrefix(req.Type.String(), "ORDER_TYPE_")
	side := strings.TrimPrefix(req.Side.String(), "ORDER_SIDE_")
	tif := strings.TrimPrefix(req.TimeInForce.String(), "TIME_IN_FORCE_")
	if tif == "UNSPECIFIED" {
		tif = "GTC"
	}

	order, err := s.queries.CreateOrder(ctx, db.CreateOrderParams{
		ParticipantID: participantID,
		Symbol:        req.Symbol,
		Side:          side,
		Type:          orderType,
		TimeInForce:   tif,
		Quantity:      req.Quantity,
		Price:         req.Price,
		LockID:        lockID,
	})
	if err != nil {
		s.log.Error("failed to create order", err)
		return nil, status.Error(codes.Internal, "failed to create order")
	}

	orderStatus, err := s.engine.Submit(ctx, order)
	if err != nil {
		s.log.Error("matching failed", err, logger.Str("order_id", order.ID.String()))
		return nil, status.Error(codes.Internal, "matching failed")
	}

	s.log.Info("order submitted",
		logger.Str("order_id", order.ID.String()),
		logger.Str("symbol", req.Symbol),
		logger.Str("side", side),
		logger.Str("status", orderStatus),
	)

	return &pb.SubmitOrderResponse{
		OrderId: order.ID.String(),
		Status:  orderStatus,
	}, nil
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

	_, err = s.queries.CancelOrder(ctx, orderID, participantID)
	if err != nil {
		return &pb.CancelOrderResponse{Cancelled: false, Reason: "order not found or already closed"}, nil
	}

	s.log.Info("order cancelled", logger.Str("order_id", req.OrderId))
	return &pb.CancelOrderResponse{Cancelled: true}, nil
}
