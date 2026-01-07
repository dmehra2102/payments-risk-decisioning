package kafka

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	w *kafka.Writer
}

func NewProducer(brokers []string) *Producer {
	return &Producer{
		w: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireAll,
			Async:        false,
			BatchTimeout: 2 * time.Second,
		},
	}
}

func (p *Producer) Publish(ctx context.Context, topic string, key []byte, value []byte, headers []kafka.Header) error {
	return p.w.WriteMessages(ctx, kafka.Message{
		Topic:   topic,
		Key:     key,
		Value:   value,
		Headers: headers,
		Time:    time.Now(),
	})
}

func (p *Producer) Close() error {
	return p.w.Close()
}
