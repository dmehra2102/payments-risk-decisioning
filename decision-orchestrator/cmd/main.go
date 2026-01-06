package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dmehra2102/payments-risk-decisioning/decision-orchestrator/internal/app"
	"github.com/dmehra2102/payments-risk-decisioning/decision-orchestrator/internal/config"
	httpHandler "github.com/dmehra2102/payments-risk-decisioning/decision-orchestrator/internal/http"
	"github.com/dmehra2102/payments-risk-decisioning/decision-orchestrator/internal/kafka"
	"github.com/dmehra2102/payments-risk-decisioning/decision-orchestrator/internal/log"
	"github.com/dmehra2102/payments-risk-decisioning/decision-orchestrator/internal/observability"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger, err := log.New(cfg.LogLevel)
	if err != nil {
		panic(err)
	}

	shutdown, err := observability.InitTracer(ctx, cfg.AppName, cfg.OTELEndpoint)
	if err != nil {
		logger.Fatal("otel init failed", zap.Error(err))
	}
	defer shutdown(ctx)

	mongoClient, err := mongo.Connect(options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		logger.Fatal("mongo connect failed", zap.Error(err))
	}

	db := mongoClient.Database(cfg.MongoDB)

	producer := kafka.New(cfg.KafkaBrokers, cfg.ProducerRetries, cfg.ProducerTimeout)
	defer producer.Close()

	orch := app.NewOrchestrator(logger, db, producer, cfg.OutboxTopic)

	consumer := kafka.NewConsumer(cfg.KafkaBrokers, cfg.KafkaGroupID, cfg.RiskDecisionTopic, orch.HandleRiskDecision)

	srv := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: httpHandler.HealthHandler(),
	}

	go func() {
		logger.Info("http server listening", zap.String("addr", cfg.HTTPAddr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("http server failed", zap.Error(err))
		}
	}()

	go func() {
		logger.Info("consumer running", zap.String("topic", cfg.RiskDecisionTopic))
		if err := consumer.Run(ctx); err != nil {
			logger.Fatal("consumer failed", zap.Error(err))
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	logger.Info("shutting down")
	_ = srv.Shutdown(context.Background())
	_ = consumer.Close()
	_ = mongoClient.Disconnect(context.Background())
	time.Sleep(300 * time.Millisecond)
}
