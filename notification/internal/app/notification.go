package app

import (
	"context"
	"encoding/json"

	"github.com/dmehra2102/payments-risk-decisioning/notification/internal/kafka"
	"github.com/dmehra2102/payments-risk-decisioning/notification/internal/notify"
	"github.com/dmehra2102/payments-risk-decisioning/notification/internal/store"
	segmentioKafka "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type NotificationApp struct {
	log      *zap.Logger
	sender   *notify.Sender
	state    *store.StateStore
	producer *kafka.Producer
	dlqTopic string
}

func New(log *zap.Logger, sender *notify.Sender, state *store.StateStore, prod *kafka.Producer, dlq string) *NotificationApp {
	return &NotificationApp{
		log:      log,
		sender:   sender,
		state:    state,
		producer: prod,
		dlqTopic: dlq,
	}
}

func (n *NotificationApp) Handle(ctx context.Context, msg segmentioKafka.Message) error {
	var envelope map[string]any
	// I'm doing this for skipping bad message format
	if err := json.Unmarshal(msg.Value, &envelope); err != nil {
		n.log.Error("invalid outbox payload", zap.Error(err))
		return nil
	}

	id := string(msg.Key)
	if n.state.Seen(id) {
		n.log.Info("duplicate event ignored", zap.String("id", id))
		return nil
	}
	if err := n.sender.Send(ctx, msg.Value); err != nil {
		n.log.Warn("notify failed", zap.Error(err))
		// Send to DLQ
		_ = n.producer.Publish(ctx, n.dlqTopic, msg.Key, msg.Value, msg.Headers)
		return err
	}
	n.state.Mark(id)
	return nil
}
