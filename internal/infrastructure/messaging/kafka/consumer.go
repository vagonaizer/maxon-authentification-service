package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
	"github.com/vagonaizer/authenitfication-service/internal/config"
	"github.com/vagonaizer/authenitfication-service/pkg/logger"
)

type Consumer struct {
	reader *kafka.Reader
	logger *logger.Logger
}

type MessageHandler func(ctx context.Context, message []byte) error

func NewConsumer(cfg *config.KafkaConfig, topic string, logger *logger.Logger) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		Topic:    topic,
		GroupID:  cfg.GroupID,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	return &Consumer{
		reader: reader,
		logger: logger,
	}
}

func (c *Consumer) Consume(ctx context.Context, handler MessageHandler) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			message, err := c.reader.ReadMessage(ctx)
			if err != nil {
				c.logger.WithError(err).Error("failed to read message")
				continue
			}

			err = handler(ctx, message.Value)
			if err != nil {
				c.logger.WithError(err).WithFields(logrus.Fields{
					"topic":     message.Topic,
					"partition": message.Partition,
					"offset":    message.Offset,
				}).Error("failed to handle message")
				continue
			}

			c.logger.WithFields(logger.Fields{
				"topic":     message.Topic,
				"partition": message.Partition,
				"offset":    message.Offset,
			}).Debug("message processed successfully")
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
