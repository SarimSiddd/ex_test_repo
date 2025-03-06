package repository

import (
	"context"
	"payment-gateway/internal/models"
)

type Gateway interface {
	FindByID(ctx context.Context, id int) (*models.Gateway, error)
}
