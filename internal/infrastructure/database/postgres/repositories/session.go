package repositories

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/vagonaizer/authenitfication-service/internal/domain/entities"
	"github.com/vagonaizer/authenitfication-service/internal/infrastructure/database/postgres"
	"github.com/vagonaizer/authenitfication-service/pkg/errors"
)

type sessionRepository struct {
	db *postgres.DB
}

func NewSessionRepository(db *postgres.DB) *sessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) Create(ctx context.Context, session *entities.Session) error {
	query := `
		INSERT INTO sessions (id, user_id, refresh_token, user_agent, ip_address, is_active, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		session.ID, session.UserID, session.RefreshToken,
		session.UserAgent, session.IPAddress, session.IsActive, session.ExpiresAt,
	).Scan(&session.CreatedAt, &session.UpdatedAt)

	if err != nil {
		return errors.DatabaseError(err)
	}

	return nil
}

func (r *sessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Session, error) {
	session := &entities.Session{}
	query := `
		SELECT id, user_id, refresh_token, user_agent, ip_address, is_active, expires_at, created_at, updated_at
		FROM sessions 
		WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&session.ID, &session.UserID, &session.RefreshToken,
		&session.UserAgent, &session.IPAddress, &session.IsActive,
		&session.ExpiresAt, &session.CreatedAt, &session.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFound("session not found")
		}
		return nil, errors.DatabaseError(err)
	}

	return session, nil
}

func (r *sessionRepository) GetByRefreshToken(ctx context.Context, refreshToken string) (*entities.Session, error) {
	session := &entities.Session{}
	query := `
		SELECT id, user_id, refresh_token, user_agent, ip_address, is_active, expires_at, created_at, updated_at
		FROM sessions 
		WHERE refresh_token = $1`

	err := r.db.QueryRowContext(ctx, query, refreshToken).Scan(
		&session.ID, &session.UserID, &session.RefreshToken,
		&session.UserAgent, &session.IPAddress, &session.IsActive,
		&session.ExpiresAt, &session.CreatedAt, &session.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFound("session not found")
		}
		return nil, errors.DatabaseError(err)
	}

	return session, nil
}

func (r *sessionRepository) GetActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.Session, error) {
	query := `
		SELECT id, user_id, refresh_token, user_agent, ip_address, is_active, expires_at, created_at, updated_at
		FROM sessions 
		WHERE user_id = $1 AND is_active = true AND expires_at > CURRENT_TIMESTAMP
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.DatabaseError(err)
	}
	defer rows.Close()

	var sessions []*entities.Session
	for rows.Next() {
		session := &entities.Session{}
		err := rows.Scan(
			&session.ID, &session.UserID, &session.RefreshToken,
			&session.UserAgent, &session.IPAddress, &session.IsActive,
			&session.ExpiresAt, &session.CreatedAt, &session.UpdatedAt,
		)
		if err != nil {
			return nil, errors.DatabaseError(err)
		}
		sessions = append(sessions, session)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.DatabaseError(err)
	}

	return sessions, nil
}

func (r *sessionRepository) Update(ctx context.Context, session *entities.Session) error {
	query := `
		UPDATE sessions 
		SET user_agent = $2, ip_address = $3, is_active = $4, expires_at = $5
		WHERE id = $1
		RETURNING updated_at`

	err := r.db.QueryRowContext(ctx, query,
		session.ID, session.UserAgent, session.IPAddress,
		session.IsActive, session.ExpiresAt,
	).Scan(&session.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.NotFound("session not found")
		}
		return errors.DatabaseError(err)
	}

	return nil
}

func (r *sessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM sessions WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return errors.DatabaseError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.DatabaseError(err)
	}

	if rowsAffected == 0 {
		return errors.NotFound("session not found")
	}

	return nil
}

func (r *sessionRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM sessions WHERE user_id = $1`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return errors.DatabaseError(err)
	}

	return nil
}

func (r *sessionRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM sessions WHERE expires_at < CURRENT_TIMESTAMP`

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return errors.DatabaseError(err)
	}

	return nil
}
