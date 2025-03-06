package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"payment-gateway/internal/models"
	"payment-gateway/internal/repository"
	"time"
)

type TransactionRepo struct {
	db *sql.DB
}

func NewTransactionRepo(db *sql.DB) repository.Transaction {
	return &TransactionRepo{
		db: db,
	}
}

func (r *TransactionRepo) Create(ctx context.Context, transaction *models.Transaction) error {

	query := ` INSERT INTO transactions (amount, type, status, gateway_id, country_id, user_id, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING id
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		transaction.Amount,
		transaction.Type,
		transaction.Status,
		transaction.GatewayID,
		transaction.CountryID,
		transaction.UserID,
		time.Now(),
	).Scan(&transaction.ID)

	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

func (r *TransactionRepo) GetByID(ctx context.Context, id int) (*models.Transaction, error) {
	query := `
		SELECT id, amount, type, status, user_id, gateway_id, country_id, created_at 
		FROM transactions 
		WHERE id = $1
	`

	var transaction models.Transaction
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&transaction.ID,
		&transaction.Amount,
		&transaction.Type,
		&transaction.Status,
		&transaction.UserID,
		&transaction.GatewayID,
		&transaction.CountryID,
		&transaction.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("transaction with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return &transaction, nil
}

func (r *TransactionRepo) UpdateStatus(ctx context.Context, id int, status string) error {
	query := `UPDATE transactions SET status = $1 WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("transaction with ID %d not found", id)
	}

	return nil
}
