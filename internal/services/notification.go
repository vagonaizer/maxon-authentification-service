package services

import (
	"context"

	"github.com/vagonaizer/authenitfication-service/internal/infrastructure/messaging/kafka"
	"github.com/vagonaizer/authenitfication-service/pkg/logger"
)

type notificationService struct {
	producer *kafka.Producer
	logger   *logger.Logger
}

func NewNotificationService(producer *kafka.Producer, logger *logger.Logger) *notificationService {
	return &notificationService{
		producer: producer,
		logger:   logger,
	}
}

func (s *notificationService) SendWelcomeEmail(ctx context.Context, userID, email string) error {
	event := map[string]interface{}{
		"type":    "welcome_email",
		"user_id": userID,
		"email":   email,
	}

	return s.producer.PublishMessage(ctx, "notifications.email", userID, event)
}

func (s *notificationService) SendPasswordResetEmail(ctx context.Context, userID, email, resetToken string) error {
	event := map[string]interface{}{
		"type":        "password_reset_email",
		"user_id":     userID,
		"email":       email,
		"reset_token": resetToken,
	}

	return s.producer.PublishMessage(ctx, "notifications.email", userID, event)
}

func (s *notificationService) SendVerificationEmail(ctx context.Context, userID, email, verificationToken string) error {
	event := map[string]interface{}{
		"type":               "verification_email",
		"user_id":            userID,
		"email":              email,
		"verification_token": verificationToken,
	}

	return s.producer.PublishMessage(ctx, "notifications.email", userID, event)
}
