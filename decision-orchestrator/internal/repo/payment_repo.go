package repo

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type PaymentStatus string

const (
	StatusPending  PaymentStatus = "PENDING"
	StatusApproved PaymentStatus = "APPROVED"
	StatusDeclined PaymentStatus = "DECLINED"
	StatusFailed   PaymentStatus = "FAILED"
)

type Payment struct {
	ID            string        `bson:"_id"`
	UserID        string        `bson:"user_id"`
	Amount        int64         `bson:"amount"`
	Currency      string        `bson:"currency"`
	MerchantID    string        `bson:"merchant_id"`
	CreatedAt     time.Time     `bson:"created_at"`
	Status        PaymentStatus `bson:"status"`
	RiskScore     float64       `bson:"risk_score,omitempty"`
	RiskReason    string        `bson:"risk_reason,omitempty"`
	CorrelationID string        `bson:"correlation_id"`
	Version       int64         `bson:"version"`
}

type PaymentRepo struct {
	col *mongo.Collection
}

func NewPaymentRepo(db *mongo.Database) *PaymentRepo {
	return &PaymentRepo{col: db.Collection("payments")}
}

func (r *PaymentRepo) UpdateDecision(ctx context.Context, id string, status PaymentStatus, score float64, reason string) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"status":      status,
			"risk_score":  score,
			"risk_reason": reason,
		},
		"$inc": bson.M{"version": 1},
	}

	res, err := r.col.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("payment not found")
	}
	return nil
}

func (r *PaymentRepo) Get(ctx context.Context, id string) (*Payment, error) {
	var p Payment
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
