// internal/transport/kafka/consumer.go
package kafka_rating

import (
	"context"
	"encoding/json"

	"github.com/RozmiDan/gameReviewHubRating/internal/config"
	"github.com/RozmiDan/gameReviewHubRating/internal/entity"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type RatingUseCase interface {
	SubmitRating(ctx context.Context, userID string, gameID string, rating int32) error
}

type Consumer struct {
	reader  *kafka.Reader
	handler RatingUseCase
	logger  *zap.Logger
}

func NewConsumer(cfg config.KafkaConfig, handler RatingUseCase, logger *zap.Logger) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		GroupID:  cfg.GroupID,
		Topic:    cfg.TopicRatings,
		MinBytes: cfg.MinBytes,
		MaxBytes: cfg.MaxBytes,
	})
	return &Consumer{
		reader:  r,
		handler: handler,
		logger:  logger.With(zap.String("component", "kafka-consumer")),
	}
}

func (c *Consumer) Start(ctx context.Context) {
	c.logger.Info("starting kafka consumer loop")
	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				c.logger.Info("context done, exiting consumer")
				return
			}
			c.logger.Error("fetch message failed", zap.Error(err))
			continue
		}

		var msg entity.RatingMessage
		if err := json.Unmarshal(m.Value, &msg); err != nil {
			c.logger.Error("invalid message, skipping", zap.Error(err))
			c.reader.CommitMessages(ctx, m)
			continue
		}

		if err := c.reader.CommitMessages(ctx, m); err != nil {
			c.logger.Warn("commit failed", zap.Error(err))
		}

		if err := c.handler.SubmitRating(ctx, msg.UserID, msg.GameID, msg.Rating); err != nil {
			c.logger.Error("handler error", zap.Error(err),
				zap.String("game_id", msg.GameID), zap.String("user_id", msg.UserID))
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
