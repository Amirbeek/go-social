package store

import (
	"context"
	"database/sql"
	"errors"
)

type RolesStore struct {
	db *sql.DB
}
type Role struct {
	ID          int64  `db:"id"`
	Name        string `db:"name"`
	Level       int    `db:"level"`
	Description string `db:"description"`
}

func (r *RolesStore) GetByName(ctx context.Context, roleName string) (*Role, error) {
	q := "SELECT id, name, level, description FROM roles WHERE name = $1"

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	var role Role

	err := r.db.QueryRowContext(ctx, q, roleName).Scan(
		&role.ID,
		&role.Name,
		&role.Level,
		&role.Description,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &role, nil
}
