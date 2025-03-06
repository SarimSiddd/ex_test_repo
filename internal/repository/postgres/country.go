package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"payment-gateway/internal/models"
	"payment-gateway/internal/repository"
)

type CountryRepo struct {
	db *sql.DB
}

func NewCountryRepo(db *sql.DB) repository.Country {
	return &CountryRepo{
		db: db,
	}
}

func (r *CountryRepo) FindByID(ctx context.Context, id int) (*models.Country, error) {
	query := `SELECT id, name, code, currency, created_at, updated_at 
              FROM countries WHERE id = $1`

	var country models.Country
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&country.ID,
		&country.Name,
		&country.Code,
		&country.Currency,
		&country.CreatedAt,
		&country.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("country with ID %d not found", id)
	}

	if err != nil {
		return nil, fmt.Errorf("error querying country: %w", err)
	}

	return &country, nil
}
