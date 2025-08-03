package repositories

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/vagonaizer/authenitfication-service/internal/domain/entities"
	"github.com/vagonaizer/authenitfication-service/internal/infrastructure/database/postgres"
	"github.com/vagonaizer/authenitfication-service/pkg/errors"
)

type roleRepository struct {
	db *postgres.DB
}

func NewRoleRepository(db *postgres.DB) *roleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) Create(ctx context.Context, role *entities.Role) error {
	query := `
		INSERT INTO roles (id, name, description)
		VALUES ($1, $2, $3)
		RETURNING created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		role.ID, role.Name, role.Description,
	).Scan(&role.CreatedAt, &role.UpdatedAt)

	if err != nil {
		return errors.DatabaseError(err)
	}

	return nil
}

func (r *roleRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Role, error) {
	role := &entities.Role{}
	query := `SELECT id, name, description, created_at, updated_at FROM roles WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFound("role not found")
		}
		return nil, errors.DatabaseError(err)
	}

	return role, nil
}

func (r *roleRepository) GetByName(ctx context.Context, name string) (*entities.Role, error) {
	role := &entities.Role{}
	query := `SELECT id, name, description, created_at, updated_at FROM roles WHERE name = $1`

	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFound("role not found")
		}
		return nil, errors.DatabaseError(err)
	}

	return role, nil
}

func (r *roleRepository) List(ctx context.Context) ([]*entities.Role, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM roles ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.DatabaseError(err)
	}
	defer rows.Close()

	var roles []*entities.Role
	for rows.Next() {
		role := &entities.Role{}
		err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt)
		if err != nil {
			return nil, errors.DatabaseError(err)
		}
		roles = append(roles, role)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.DatabaseError(err)
	}

	return roles, nil
}

func (r *roleRepository) Update(ctx context.Context, role *entities.Role) error {
	query := `
		UPDATE roles 
		SET name = $2, description = $3
		WHERE id = $1
		RETURNING updated_at`

	err := r.db.QueryRowContext(ctx, query, role.ID, role.Name, role.Description).Scan(&role.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.NotFound("role not found")
		}
		return errors.DatabaseError(err)
	}

	return nil
}

func (r *roleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM roles WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return errors.DatabaseError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.DatabaseError(err)
	}

	if rowsAffected == 0 {
		return errors.NotFound("role not found")
	}

	return nil
}

func (r *roleRepository) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	query := `INSERT INTO user_roles (id, user_id, role_id) VALUES ($1, $2, $3) ON CONFLICT (user_id, role_id) DO NOTHING`

	_, err := r.db.ExecContext(ctx, query, uuid.New(), userID, roleID)
	if err != nil {
		return errors.DatabaseError(err)
	}

	return nil
}

func (r *roleRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	query := `DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`

	result, err := r.db.ExecContext(ctx, query, userID, roleID)
	if err != nil {
		return errors.DatabaseError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.DatabaseError(err)
	}

	if rowsAffected == 0 {
		return errors.NotFound("user role assignment not found")
	}

	return nil
}

func (r *roleRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*entities.Role, error) {
	query := `
		SELECT r.id, r.name, r.description, r.created_at, r.updated_at
		FROM roles r
		INNER JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY r.name`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.DatabaseError(err)
	}
	defer rows.Close()

	var roles []*entities.Role
	for rows.Next() {
		role := &entities.Role{}
		err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt)
		if err != nil {
			return nil, errors.DatabaseError(err)
		}
		roles = append(roles, role)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.DatabaseError(err)
	}

	return roles, nil
}
