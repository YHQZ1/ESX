package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	riskpb "github.com/YHQZ1/esx/packages/proto/risk"
	"github.com/YHQZ1/esx/services/risk-engine/internal/checks"
	"github.com/YHQZ1/esx/services/risk-engine/internal/locks"
	"github.com/rs/zerolog"
)

type Server struct {
	riskpb.UnimplementedRiskServiceServer
	locks *locks.Manager
	log   zerolog.Logger
}

func NewServer(lockManager *locks.Manager, log zerolog.Logger) *Server {
	return &Server{locks: lockManager, log: log}
}

func (s *Server) CheckAndLock(ctx context.Context, req *riskpb.CheckAndLockRequest) (*riskpb.CheckAndLockResponse, error) {
	if req.ParticipantId == "" {
		return nil, status.Error(codes.InvalidArgument, "participant_id is required")
	}
	if req.Symbol == "" {
		return nil, status.Error(codes.InvalidArgument, "symbol is required")
	}
	if req.Quantity <= 0 {
		return nil, status.Error(codes.InvalidArgument, "quantity must be greater than zero")
	}
	if req.Side == riskpb.OrderSide_ORDER_SIDE_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "side is required")
	}

	var result *locks.LockResult
	var err error

	switch req.Side {
	case riskpb.OrderSide_ORDER_SIDE_BUY:
		if req.Price <= 0 {
			return nil, status.Error(codes.InvalidArgument, "price must be greater than zero for buy orders")
		}
		result, err = s.locks.LockForBuy(ctx, req.ParticipantId, req.Symbol, req.Quantity, req.Price)
	case riskpb.OrderSide_ORDER_SIDE_SELL:
		result, err = s.locks.LockForSell(ctx, req.ParticipantId, req.Symbol, req.Quantity, req.Price)
	default:
		return nil, status.Error(codes.InvalidArgument, "unrecognised order side")
	}

	if err != nil {
		s.log.Warn().
			Str("participant_id", req.ParticipantId).
			Str("symbol", req.Symbol).
			Int32("side", int32(req.Side)).
			Int64("quantity", req.Quantity).
			Int64("price", req.Price).
			Err(err).
			Msg("risk check failed")

		if errors.Is(err, checks.ErrInsufficientCash) || errors.Is(err, checks.ErrInsufficientShares) {
			return &riskpb.CheckAndLockResponse{Approved: false, Reason: err.Error()}, nil
		}
		if errors.Is(err, checks.ErrAccountNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		if errors.Is(err, checks.ErrInvalidPrice) || errors.Is(err, checks.ErrInvalidQuantity) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	s.log.Info().
		Str("participant_id", req.ParticipantId).
		Str("symbol", req.Symbol).
		Int32("side", int32(req.Side)).
		Int64("quantity", req.Quantity).
		Int64("price", req.Price).
		Str("lock_id", result.LockID).
		Msg("collateral locked")

	return &riskpb.CheckAndLockResponse{
		Approved: true,
		LockId:   result.LockID,
	}, nil
}

func (s *Server) ReleaseLock(ctx context.Context, req *riskpb.ReleaseLockRequest) (*riskpb.ReleaseLockResponse, error) {
	if req.LockId == "" {
		return nil, status.Error(codes.InvalidArgument, "lock_id is required")
	}

	var err error
	if req.FilledQuantity > 0 {
		err = s.locks.Consume(ctx, req.LockId)
	} else {
		err = s.locks.Release(ctx, req.LockId)
	}

	if err != nil {
		s.log.Error().
			Str("lock_id", req.LockId).
			Int64("filled_quantity", req.FilledQuantity).
			Err(err).
			Msg("lock release failed")
		return nil, status.Error(codes.Internal, err.Error())
	}

	s.log.Info().
		Str("lock_id", req.LockId).
		Int64("filled_quantity", req.FilledQuantity).
		Msg("lock released")

	return &riskpb.ReleaseLockResponse{Released: true}, nil
}
