package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/YHQZ1/esx/packages/logger"
	"github.com/segmentio/kafka-go"
)

type HandlerFunc func(ctx context.Context, msg Message) error

type Message struct {
	Topic     string
	Key       string
	Value     []byte
	Partition int
	Offset    int64
}

type Consumer struct {
	reader  *kafka.Reader
	log     *logger.Logger
	handler HandlerFunc
}

func NewConsumer(brokers []string, topic, groupID string, log *logger.Logger) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: 0,
	})

	return &Consumer{reader: r, log: log}
}

func (c *Consumer) RegisterHandler(fn HandlerFunc) {
	c.handler = fn
}

func (c *Consumer) Start(ctx context.Context) error {
	if c.handler == nil {
		return fmt.Errorf("no handler registered")
	}

	c.log.Info("consumer started",
		logger.Str("topic", c.reader.Config().Topic),
		logger.Str("group", c.reader.Config().GroupID),
	)

	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			c.log.Error("failed to fetch message", err,
				logger.Str("topic", c.reader.Config().Topic),
			)
			continue
		}

		m := Message{
			Topic:     msg.Topic,
			Key:       string(msg.Key),
			Value:     msg.Value,
			Partition: msg.Partition,
			Offset:    msg.Offset,
		}

		const maxRetries = 3
		var handlerErr error
		for i := range maxRetries {
			handlerErr = c.handler(ctx, m)
			if handlerErr == nil {
				break
			}
			c.log.Error("handler failed, retrying", handlerErr,
				logger.Str("topic", msg.Topic),
				logger.Str("key", string(msg.Key)),
				logger.Int("attempt", i+1),
			)
		}

		if handlerErr != nil {
			c.log.Error("handler permanently failed, skipping", handlerErr,
				logger.Str("topic", msg.Topic),
				logger.Int64("offset", msg.Offset),
			)
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				c.log.Error("failed to commit poison pill", err,
					logger.Str("topic", msg.Topic),
					logger.Int64("offset", msg.Offset),
				)
			}
			continue
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			c.log.Error("failed to commit message", err,
				logger.Str("topic", msg.Topic),
				logger.Int64("offset", msg.Offset),
			)
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}

func Decode[T any](msg Message) (T, error) {
	var v T
	err := json.Unmarshal(msg.Value, &v)
	return v, err
}
