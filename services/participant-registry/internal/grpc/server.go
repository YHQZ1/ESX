package grpc

import (
	"context"
	"database/sql"
	"sync"

	"github.com/YHQZ1/esx/packages/logger"
	pb "github.com/YHQZ1/esx/packages/proto/participant"
	"github.com/YHQZ1/esx/services/participant-registry/internal/db"
	"github.com/YHQZ1/esx/services/participant-registry/internal/lib"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedParticipantServiceServer
	db    db.Querier
	log   *logger.Logger
	cache sync.Map
}

func NewServer(database db.Querier, log *logger.Logger) *Server {
	return &Server{db: database, log: log}
}

func (s *Server) ValidateAPIKey(ctx context.Context, req *pb.ValidateAPIKeyRequest) (*pb.ValidateAPIKeyResponse, error) {
	if req.ApiKey == "" {
		return nil, status.Error(codes.InvalidArgument, "api_key is required")
	}

	keyHash := lib.HashAPIKey(req.ApiKey)

	if val, ok := s.cache.Load(keyHash); ok {
		return val.(*pb.ValidateAPIKeyResponse), nil
	}

	apiKey, err := s.db.GetAPIKeyByHash(ctx, keyHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return &pb.ValidateAPIKeyResponse{IsActive: false}, nil
		}
		s.log.Error("failed to look up api key", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	participant, err := s.db.GetParticipantByID(ctx, apiKey.ParticipantID)
	if err != nil {
		s.log.Error("failed to get participant", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	resp := &pb.ValidateAPIKeyResponse{
		ParticipantId: participant.ID.String(),
		Name:          participant.Name,
		IsActive:      participant.Status == "active",
	}

	s.cache.Store(keyHash, resp)

	return resp, nil
}
