package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"payment-gateway/internal/models"
	"payment-gateway/internal/repository"
)

type GatewayRepo struct {
	db *sql.DB
}

func NewGatewayRepo(db *sql.DB) repository.Gateway {
	return &GatewayRepo{
		db: db,
	}
}

func (r *GatewayRepo) FindByID(ctx context.Context, id int) (*models.Gateway, error) {
	query := `SELECT id, name, dataformatsupported, created_at, updated_at 
              FROM gateways WHERE id = $1`

	var gateway models.Gateway
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&gateway.ID,
		&gateway.Name,
		&gateway.DataFormatSupported,
		&gateway.CreatedAt,
		&gateway.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("country with ID %d not found", id)
	}

	if err != nil {
		return nil, fmt.Errorf("error querying country: %w", err)
	}

	return &gateway, nil
}
