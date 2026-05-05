package grpc

import (
	"context"

	"github.com/YHQZ1/esx/packages/logger"
	pb "github.com/YHQZ1/esx/packages/proto/risk"
	"github.com/YHQZ1/esx/services/risk-engine/internal/checks"
	"github.com/YHQZ1/esx/services/risk-engine/internal/locks"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedRiskServiceServer
	checker *checks.Checker
	locker  *locks.Manager
	log     *logger.Logger
}

func NewServer(checker *checks.Checker, locker *locks.Manager, log *logger.Logger) *Server {
	return &Server{checker: checker, locker: locker, log: log}
}

func (s *Server) CheckAndLock(ctx context.Context, req *pb.CheckAndLockRequest) (*pb.CheckAndLockResponse, error) {
	participantID, err := uuid.Parse(req.ParticipantId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid participant_id")
	}

	if req.Side == pb.OrderSide_ORDER_SIDE_BUY {
		// --- OLD POSTGRES CHECK COMMENTED OUT ---
		// if err := s.checker.CheckBuy(ctx, participantID, req.Price, req.Quantity); err != nil {
		// 	s.log.Info("risk check rejected",
		// 		logger.Str("participant_id", req.ParticipantId),
		// 		logger.Str("symbol", req.Symbol),
		// 		logger.Str("reason", err.Error()),
		// 	)
		// 	return &pb.CheckAndLockResponse{Approved: false, Reason: err.Error()}, nil
		// } // <-- Notice this brace is safely commented out now!

		lockID, err := s.locker.LockCash(ctx, participantID, req.Symbol, req.Price, req.Quantity)
		if err != nil {
			// Catch the new Redis rejection gracefully
			if err.Error() == "insufficient funds" {
				s.log.Info("risk check rejected by redis", logger.Str("participant_id", req.ParticipantId), logger.Str("reason", err.Error()))
				return &pb.CheckAndLockResponse{Approved: false, Reason: err.Error()}, nil
			}

			s.log.Error("failed to lock cash", err, logger.Str("participant_id", req.ParticipantId))
			return nil, status.Error(codes.Internal, "failed to lock collateral")
		}

		s.log.Info("cash locked",
			logger.Str("participant_id", req.ParticipantId),
			logger.Str("symbol", req.Symbol),
			logger.Str("lock_id", lockID.String()),
		)

		return &pb.CheckAndLockResponse{Approved: true, LockId: lockID.String()}, nil
	}

	if req.Side == pb.OrderSide_ORDER_SIDE_SELL {
		if err := s.checker.CheckSell(ctx, participantID, req.Symbol, req.Quantity); err != nil {
			s.log.Info("risk check rejected",
				logger.Str("participant_id", req.ParticipantId),
				logger.Str("symbol", req.Symbol),
				logger.Str("reason", err.Error()),
			)
			return &pb.CheckAndLockResponse{Approved: false, Reason: err.Error()}, nil
		}

		lockID, err := s.locker.LockShares(ctx, participantID, req.Symbol, req.Quantity)
		if err != nil {
			s.log.Error("failed to lock shares", err, logger.Str("participant_id", req.ParticipantId))
			return nil, status.Error(codes.Internal, "failed to lock collateral")
		}

		s.log.Info("shares locked",
			logger.Str("participant_id", req.ParticipantId),
			logger.Str("symbol", req.Symbol),
			logger.Str("lock_id", lockID.String()),
		)

		return &pb.CheckAndLockResponse{Approved: true, LockId: lockID.String()}, nil
	}

	return nil, status.Error(codes.InvalidArgument, "invalid order side")
}

func (s *Server) ReleaseLock(ctx context.Context, req *pb.ReleaseLockRequest) (*pb.ReleaseLockResponse, error) {
	lockID, err := uuid.Parse(req.LockId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid lock_id")
	}

	if err := s.locker.Release(ctx, lockID, req.FilledQuantity); err != nil {
		s.log.Error("failed to release lock", err, logger.Str("lock_id", req.LockId))
		return &pb.ReleaseLockResponse{Released: false}, nil
	}

	s.log.Info("lock released", logger.Str("lock_id", req.LockId))
	return &pb.ReleaseLockResponse{Released: true}, nil
}
