package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/dmehra2102/payments-risk-decisioning/notification/internal/app"
	"github.com/dmehra2102/payments-risk-decisioning/notification/internal/config"
	httpHandler "github.com/dmehra2102/payments-risk-decisioning/notification/internal/http"
	"github.com/dmehra2102/payments-risk-decisioning/notification/internal/kafka"
	"github.com/dmehra2102/payments-risk-decisioning/notification/internal/log"
	"github.com/dmehra2102/payments-risk-decisioning/notification/internal/notify"
	"github.com/dmehra2102/payments-risk-decisioning/notification/internal/observability"
	"github.com/dmehra2102/payments-risk-decisioning/notification/internal/store"
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

	sender := notify.New(cfg.NotifyWebhook, cfg.Timeout)
	state := store.New()
	producer := kafka.NewProducer(cfg.KafkaBrokers)
	defer producer.Close()

	app := app.New(logger, sender, state, producer, cfg.DLQTopic)

	consumer := kafka.NewConsumer(cfg.KafkaBrokers, cfg.GroupID, cfg.OutboxTopic, app.Handle)
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
		logger.Info("consumer running", zap.String("topic", cfg.OutboxTopic))
		if err := consumer.Run(ctx); err != nil {
			logger.Fatal("consumer failed", zap.Error(err))
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	_ = srv.Shutdown(context.Background())
	_ = consumer.Close()
}
