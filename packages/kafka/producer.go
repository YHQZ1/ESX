package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/YHQZ1/esx/packages/logger"
	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
	log    *logger.Logger
}

func NewProducer(brokers []string, topic string, log *logger.Logger) *Producer {
	w := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    1000,
		BatchTimeout: 5 * time.Millisecond,
		RequiredAcks: kafka.RequireAll,
		Async:        true,
	}

	return &Producer{writer: w, log: log}
}

func (p *Producer) Publish(ctx context.Context, key string, payload any) error {
	value, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(key),
		Value: value,
		Time:  time.Now(),
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		p.log.Error("failed to publish message", err,
			logger.Str("topic", p.writer.Topic),
			logger.Str("key", key),
		)
		return err
	}

	p.log.Debug("message published",
		logger.Str("topic", p.writer.Topic),
		logger.Str("key", key),
	)

	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
