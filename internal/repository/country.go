package repository

import (
	"context"
	"payment-gateway/internal/models"
)

type Country interface {
	FindByID(ctx context.Context, id int) (*models.Country, error)
}
