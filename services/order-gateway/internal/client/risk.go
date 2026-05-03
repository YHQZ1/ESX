package client

import (
	"context"

	pb "github.com/YHQZ1/esx/packages/proto/risk"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RiskClient struct {
	client pb.RiskServiceClient
}

func NewRiskClient(addr string) (*RiskClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &RiskClient{client: pb.NewRiskServiceClient(conn)}, nil
}

func (c *RiskClient) CheckAndLock(ctx context.Context, participantID, symbol string, side pb.OrderSide, quantity, price int64) (string, bool, string, error) {
	resp, err := c.client.CheckAndLock(ctx, &pb.CheckAndLockRequest{
		ParticipantId: participantID,
		Symbol:        symbol,
		Side:          side,
		Quantity:      quantity,
		Price:         price,
	})
	if err != nil {
		return "", false, "", err
	}
	return resp.LockId, resp.Approved, resp.Reason, nil
}
