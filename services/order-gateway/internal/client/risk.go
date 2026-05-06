package client

import (
	"context"
	"sync/atomic"

	pb "github.com/YHQZ1/esx/packages/proto/risk"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RiskClient struct {
	pool []pb.RiskServiceClient
	next uint64
}

func NewRiskClient(addr string) (*RiskClient, error) {
	poolSize := 10
	clients := make([]pb.RiskServiceClient, poolSize)

	for i := 0; i < poolSize; i++ {
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, err
		}
		clients[i] = pb.NewRiskServiceClient(conn)
	}

	return &RiskClient{pool: clients}, nil
}

func (c *RiskClient) CheckAndLock(ctx context.Context, participantID, symbol string, side pb.OrderSide, quantity, price int64) (string, bool, string, error) {
	idx := atomic.AddUint64(&c.next, 1) % uint64(len(c.pool))
	client := c.pool[idx]

	resp, err := client.CheckAndLock(ctx, &pb.CheckAndLockRequest{
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
