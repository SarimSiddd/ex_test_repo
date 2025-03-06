package repository

import (
	"context"
	"payment-gateway/internal/models"
)

type User interface {
	FindByID(ctx context.Context, id int) (*models.User, error)
}
