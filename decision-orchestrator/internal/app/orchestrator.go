package app

import (
	"context"
	"encoding/json"
	"time"

	"github.com/dmehra2102/payments-risk-decisioning/decision-orchestrator/internal/kafka"
	"github.com/dmehra2102/payments-risk-decisioning/decision-orchestrator/internal/outbox"
	"github.com/dmehra2102/payments-risk-decisioning/decision-orchestrator/internal/repo"
	segmentioKafka "github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
)

type Orchestrator struct {
	log           *zap.Logger
	payments      *repo.PaymentRepo
	outbox        *outbox.OutboxRepo
	kafkaProducer *kafka.Producer
	outboxTopic   string
}

type RiskDecision struct {
	PaymentID     string  `json:"payment_id"`
	Decision      string  `json:"decision"`
	Score         float64 `json:"score"`
	Reason        string  `json:"reason"`
	CorrelationID string  `json:"correlation_id"`
}

func NewOrchestrator(l *zap.Logger, db *mongo.Database, prod *kafka.Producer, outboxTopic string) *Orchestrator {
	return &Orchestrator{
		log:           l,
		payments:      repo.NewPaymentRepo(db),
		outbox:        outbox.NewOutboxRepo(db),
		kafkaProducer: prod,
		outboxTopic:   outboxTopic,
	}
}

func (o *Orchestrator) HandleRiskDecision(ctx context.Context, msg segmentioKafka.Message) error {
	var rd RiskDecision
	if err := json.Unmarshal(msg.Value, &rd); err != nil {
		o.log.Error("failed to marshal risk decision", zap.Error(err))
		return err
	}

	status := repo.StatusDeclined
	if rd.Decision == string(repo.StatusApproved) {
		status = repo.StatusApproved
	}

	if err := o.payments.UpdateDecision(ctx, rd.PaymentID, status, rd.Score, rd.Reason); err != nil {
		o.log.Error("failed to update payment decision", zap.Error(err), zap.String("payment_id", rd.PaymentID))
		// compensation: mark failed and emit event
		_ = o.payments.UpdateDecision(ctx, rd.PaymentID, repo.StatusFailed, rd.Score, "compensation: update failed")
	}

	eventPayload, _ := json.Marshal(map[string]any{
		"type":        "PaymentDecisionFinalized",
		"payment_id":  rd.PaymentID,
		"status":      status,
		"score":       rd.Score,
		"reason":      rd.Reason,
		"correlation": rd.CorrelationID,
		"ts":          time.Now().UTC(),
	})

	outboxID := rd.PaymentID + ":final:" + rd.CorrelationID
	if err := o.outbox.Insert(ctx, outbox.OutboxEvent{
		ID:            outboxID,
		AggregateID:   rd.PaymentID,
		Type:          "PaymentDecisionFinalized",
		Payload:       eventPayload,
		Headers:       map[string]any{"content-type": "application/json"},
		CreatedAt:     time.Now(),
		Published:     false,
		CorrelationID: rd.CorrelationID,
	}); err != nil {
		o.log.Error("failed to insert outbox event", zap.Error(err))
		return err
	}

	kafkaHeaders := []segmentioKafka.Header{
		{Key: "correlation_id", Value: []byte(rd.CorrelationID)},
		{Key: "event_type", Value: []byte("PaymentDecisionFinalized")},
	}

	if err := o.kafkaProducer.Publish(ctx, o.outboxTopic, []byte(rd.PaymentID), eventPayload, kafkaHeaders); err != nil {
		o.log.Error("failed to publish outbox event", zap.Error(err))
		return err
	}

	if err := o.outbox.MarkPublished(ctx, outboxID); err != nil {
		o.log.Warn("failed to mark outbox published", zap.Error(err))
	}

	return nil
}
