package kafka

import (
	"context"
	"encoding/json"

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

		if err := c.handler(ctx, m); err != nil {
			c.log.Error("handler failed", err,
				logger.Str("topic", msg.Topic),
				logger.Str("key", string(msg.Key)),
				logger.Int64("offset", msg.Offset),
			)
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
