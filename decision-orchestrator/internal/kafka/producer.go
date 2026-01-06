package kafka

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	w *kafka.Writer
}

func New(brokers []string, retries int, timeout time.Duration) *Producer {
	return &Producer{
		w: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireAll,
			Async:        false,
			Transport:    &kafka.Transport{},
			BatchTimeout: timeout,
			MaxAttempts:  retries,
		},
	}
}

func (p *Producer) Publish(ctx context.Context, topic string, key []byte, value []byte, headers []kafka.Header) error {
	msg := kafka.Message{
		Topic:   topic,
		Key:     key,
		Value:   value,
		Headers: headers,
		Time:    time.Now(),
	}

	return p.w.WriteMessages(ctx, msg)
}

func (p *Producer) Close() error {
	return p.w.Close()
}
