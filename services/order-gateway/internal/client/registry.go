package client

import (
	"context"

	pb "github.com/YHQZ1/esx/packages/proto/participant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RegistryClient struct {
	client pb.ParticipantServiceClient
}

func NewRegistryClient(addr string) (*RegistryClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &RegistryClient{client: pb.NewParticipantServiceClient(conn)}, nil
}

func (c *RegistryClient) ValidateAPIKey(ctx context.Context, apiKey string) (string, bool, error) {
	resp, err := c.client.ValidateAPIKey(ctx, &pb.ValidateAPIKeyRequest{ApiKey: apiKey})
	if err != nil {
		return "", false, err
	}
	return resp.ParticipantId, resp.IsActive, nil
}
