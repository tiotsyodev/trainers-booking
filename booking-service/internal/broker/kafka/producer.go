package kafka

import (
	"context"

	"github.com/kelseyhightower/envconfig"
	"github.com/segmentio/kafka-go"
)

type Publisher struct {
	Writer *kafka.Writer
}

type brokerConfig struct {
	Brokers []string `envconfig:"BROKERS" required:"true"`
	Topic   string   `envconfig:"TOPIC" required:"true"`
}

func mustLoadConfig() brokerConfig {
	var cfg brokerConfig
	envconfig.MustProcess("KAFKA", &cfg)

	return cfg
}

func NewPublisher() Publisher {

	cfg := mustLoadConfig()

	wr := &kafka.Writer{
		Addr:  kafka.TCP(cfg.Brokers...),
		Topic: cfg.Topic,
	}

	return Publisher{
		Writer: wr,
	}
}

func (p *Publisher) Close() error { return p.Writer.Close() }

func (p *Publisher) Publish(ctx context.Context, key string, payload []byte) error {
	return p.Writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: payload,
	})
}
