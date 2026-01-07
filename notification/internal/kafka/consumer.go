package kafka

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

type Handler func(ctx context.Context, msg kafka.Message) error

type Consumer struct {
	r       *kafka.Reader
	handler Handler
}

func NewConsumer(brokers []string, groupID, topic string, handler Handler) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		GroupID:        groupID,
		Topic:          topic,
		MinBytes:       1,
		MaxBytes:       10e6,
		StartOffset:    kafka.LastOffset,
		CommitInterval: time.Second,
	})
	return &Consumer{r: r, handler: handler}
}

func (c *Consumer) Run(ctx context.Context) error {
	for {
		msg, err := c.r.FetchMessage(ctx)
		if err != nil {
			return err
		}

		if err := c.handler(ctx, msg); err != nil {
			continue
		}

		if err := c.r.CommitMessages(ctx, msg); err != nil {
			return err
		}
	}
}

func (c *Consumer) Close() error {
	return c.r.Close()
}
