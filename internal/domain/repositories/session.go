package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/vagonaizer/authenitfication-service/internal/domain/entities"
)

type SessionRepository interface {
	Create(ctx context.Context, session *entities.Session) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Session, error)
	GetByRefreshToken(ctx context.Context, refreshToken string) (*entities.Session, error)
	GetActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.Session, error)
	Update(ctx context.Context, session *entities.Session) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}
