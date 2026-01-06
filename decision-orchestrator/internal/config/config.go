package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	AppName           string
	Env               string
	LogLevel          string
	MongoURI          string
	MongoDB           string
	KafkaBrokers      []string
	KafkaGroupID      string
	OutboxTopic       string
	RiskDecisionTopic string
	ProducerAcks      string
	ProducerRetries   int
	ProducerTimeout   time.Duration
	OTELExporter      string
	OTELEndpoint      string
	HTTPAddr          string
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetEnvPrefix("ORCH")
	v.AutomaticEnv()

	v.SetDefault("APP_NAME", "decision-orchestrator")
	v.SetDefault("ENV", "dev")
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("MONGO_URI", "mongodb://mongo:27017")
	v.SetDefault("MONGO_DB", "payments")
	v.SetDefault("KAFKA_BROKERS", []string{"kafka:9092"})
	v.SetDefault("KAFKA_GROUP_ID", "decision-orchestrator")
	v.SetDefault("OUTBOX_TOPIC", "payments.outbox")
	v.SetDefault("RISK_DECISION_TOPIC", "risk.decisions")
	v.SetDefault("PRODUCER_ACKS", "all")
	v.SetDefault("PRODUCER_RETRIES", 5)
	v.SetDefault("PRODUCER_TIMEOUT", 5*time.Second)
	v.SetDefault("OTEL_EXPORTER", "otlp")
	v.SetDefault("OTEL_ENDPOINT", "otel-collector:4317")
	v.SetDefault("HTTP_ADDR", ":8082")

	cfg := &Config{
		AppName:           v.GetString("APP_NAME"),
		Env:               v.GetString("ENV"),
		LogLevel:          v.GetString("LOG_LEVEL"),
		MongoURI:          v.GetString("MONGO_URI"),
		MongoDB:           v.GetString("MONGO_DB"),
		KafkaBrokers:      v.GetStringSlice("KAFKA_BROKERS"),
		KafkaGroupID:      v.GetString("KAFKA_GROUP_ID"),
		OutboxTopic:       v.GetString("OUTBOX_TOPIC"),
		RiskDecisionTopic: v.GetString("RISK_DECISION_TOPIC"),
		ProducerAcks:      v.GetString("PRODUCER_ACKS"),
		ProducerRetries:   v.GetInt("PRODUCER_RETRIES"),
		ProducerTimeout:   v.GetDuration("PRODUCER_TIMEOUT"),
		OTELExporter:      v.GetString("OTEL_EXPORTER"),
		OTELEndpoint:      v.GetString("OTEL_ENDPOINT"),
		HTTPAddr:          v.GetString("HTTP_ADDR"),
	}
	return cfg, nil
}
