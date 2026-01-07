package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
    AppName       string
    Env           string
    LogLevel      string
    KafkaBrokers  []string
    GroupID       string
    OutboxTopic   string
    DLQTopic      string
    HTTPAddr      string
    OTELEndpoint  string
    NotifyWebhook string
    Timeout       time.Duration
}

func Load() (*Config, error) {
    v := viper.New()
    v.SetEnvPrefix("NOTIF")
    v.AutomaticEnv()

    v.SetDefault("APP_NAME", "notification")
    v.SetDefault("ENV", "dev")
    v.SetDefault("LOG_LEVEL", "info")
    v.SetDefault("KAFKA_BROKERS", []string{"kafka:9092"})
    v.SetDefault("GROUP_ID", "notification")
    v.SetDefault("OUTBOX_TOPIC", "payments.outbox")
    v.SetDefault("DLQ_TOPIC", "payments.outbox.dlq")
    v.SetDefault("HTTP_ADDR", ":8083")
    v.SetDefault("OTEL_ENDPOINT", "otel-collector:4317")
    v.SetDefault("NOTIFY_WEBHOOK", "http://mock-webhook:8080/notify")
    v.SetDefault("TIMEOUT", 5*time.Second)

    return &Config{
        AppName:       v.GetString("APP_NAME"),
        Env:           v.GetString("ENV"),
        LogLevel:      v.GetString("LOG_LEVEL"),
        KafkaBrokers:  v.GetStringSlice("KAFKA_BROKERS"),
        GroupID:       v.GetString("GROUP_ID"),
        OutboxTopic:   v.GetString("OUTBOX_TOPIC"),
        DLQTopic:      v.GetString("DLQ_TOPIC"),
        HTTPAddr:      v.GetString("HTTP_ADDR"),
        OTELEndpoint:  v.GetString("OTEL_ENDPOINT"),
        NotifyWebhook: v.GetString("NOTIFY_WEBHOOK"),
        Timeout:       v.GetDuration("TIMEOUT"),
    }, nil
}
