package client

import (
	"context"
	"sync/atomic"

	pb "github.com/YHQZ1/esx/packages/proto/matching"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type MatchingClient struct {
	pool []pb.MatchingServiceClient
	next uint64
}

func NewMatchingClient(addr string) (*MatchingClient, error) {
	poolSize := 10
	clients := make([]pb.MatchingServiceClient, poolSize)

	for i := 0; i < poolSize; i++ {
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, err
		}
		clients[i] = pb.NewMatchingServiceClient(conn)
	}

	return &MatchingClient{pool: clients}, nil
}

func (c *MatchingClient) SubmitOrder(ctx context.Context, participantID, symbol, lockID string, side pb.OrderSide, orderType pb.OrderType, tif pb.TimeInForce, quantity, price int64) (string, pb.OrderStatus, error) {
	idx := atomic.AddUint64(&c.next, 1) % uint64(len(c.pool))
	client := c.pool[idx]

	resp, err := client.SubmitOrder(ctx, &pb.SubmitOrderRequest{
		ParticipantId: participantID,
		Symbol:        symbol,
		Side:          side,
		Type:          orderType,
		TimeInForce:   tif,
		Quantity:      quantity,
		Price:         price,
		LockId:        lockID,
	})
	if err != nil {
		return "", pb.OrderStatus_ORDER_STATUS_UNSPECIFIED, err
	}
	return resp.OrderId, resp.Status, nil
}

func (c *MatchingClient) CancelOrder(ctx context.Context, orderID, participantID, lockID string) (bool, string, error) {
	idx := atomic.AddUint64(&c.next, 1) % uint64(len(c.pool))
	client := c.pool[idx]

	resp, err := client.CancelOrder(ctx, &pb.CancelOrderRequest{
		OrderId:       orderID,
		ParticipantId: participantID,
		LockId:        lockID,
	})
	if err != nil {
		return false, "", err
	}
	return resp.Cancelled, resp.Reason, nil
}
