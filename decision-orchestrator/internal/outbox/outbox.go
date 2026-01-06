package outbox

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type OutboxRepo struct {
	col *mongo.Collection
}

type OutboxEvent struct {
	ID            string    `bson:"_id"`
	AggregateID   string    `bson:"aggregate_id"`
	Type          string    `bson:"type"`
	Payload       []byte    `bson:"payload"`
	Headers       bson.M    `bson:"headers"`
	CreatedAt     time.Time `bson:"created_at"`
	Published     bool      `bson:"published"`
	PublishedAt   time.Time `bson:"published_at,omitempty"`
	CorrelationID string    `bson:"correlation_id"`
}

func NewOutboxRepo(db *mongo.Database) *OutboxRepo {
	return &OutboxRepo{col: db.Collection("outbox")}
}

func (r *OutboxRepo) Insert(ctx context.Context, event OutboxEvent) error {
	_, err := r.col.InsertOne(ctx, event)
	return err
}

func (r *OutboxRepo) MarkPublished(ctx context.Context, id string) error {
	update := bson.M{
		"$set": bson.M{
			"published":     true,
			"$published_at": time.Now(),
		},
	}
	_, err := r.col.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}
