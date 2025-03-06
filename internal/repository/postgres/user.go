package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"payment-gateway/internal/models"
	"payment-gateway/internal/repository"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) repository.User {
	return &UserRepo{
		db: db,
	}
}

func (r *UserRepo) FindByID(ctx context.Context, id int) (*models.User, error) {
	query := `SELECT id, username, email, country_id FROM users WHERE id = $1`

	var user models.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CountryID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user with ID %d not found", id)
	}

	if err != nil {
		return nil, fmt.Errorf("error querying user: %w", err)
	}

	return &user, nil
}
