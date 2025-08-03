package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
	"github.com/vagonaizer/authenitfication-service/internal/config"
	"github.com/vagonaizer/authenitfication-service/pkg/logger"
)

type Producer struct {
	writer *kafka.Writer
	logger *logger.Logger
}

func NewProducer(cfg *config.KafkaConfig, logger *logger.Logger) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    cfg.BatchSize,
		BatchTimeout: cfg.BatchTimeout,
		RequiredAcks: kafka.RequireOne,
		Async:        false,
	}

	return &Producer{
		writer: writer,
		logger: logger,
	}
}

func (p *Producer) PublishMessage(ctx context.Context, topic string, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		p.logger.WithError(err).Error("failed to marshal message")
		return err
	}

	message := kafka.Message{
		Topic:     topic,
		Key:       []byte(key),
		Value:     data,
		Time:      time.Now(),
		Partition: 0,
	}

	err = p.writer.WriteMessages(ctx, message)
	if err != nil {
		p.logger.WithError(err).WithFields(logrus.Fields{
			"topic": topic,
			"key":   key,
		}).Error("failed to publish message")
		return err
	}

	p.logger.WithFields(logger.Fields{
		"topic": topic,
		"key":   key,
	}).Debug("message published successfully")

	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
