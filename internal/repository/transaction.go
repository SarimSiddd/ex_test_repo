package repository

import (
	"context"
	"payment-gateway/internal/models"
)

type Transaction interface {
	Create(ctx context.Context, transaction *models.Transaction) error
	UpdateStatus(ctx context.Context, transactionID int, status string) error
	GetByID(ctx context.Context, transactionID int) (*models.Transaction, error)
}
