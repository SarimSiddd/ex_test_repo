package services

import (
	"context"
	"fmt"
	"payment-gateway/internal/config"
	"payment-gateway/internal/models"
	"payment-gateway/internal/repository"
	"time"
)

type TransactionProcessor struct {
	gatewayConfig   *config.GatewayConfig
	gatewaySelector *GatewaySelector
	transactionRepo repository.Transaction
}

func NewTransactionProcessor(
	gatewayConfig *config.GatewayConfig,
	gatewaySelector *GatewaySelector,
	transactionRepo repository.Transaction,
) *TransactionProcessor {
	return &TransactionProcessor{
		gatewayConfig:   gatewayConfig,
		gatewaySelector: gatewaySelector,
		transactionRepo: transactionRepo,
	}
}

func (p *TransactionProcessor) processTransaction(
	ctx context.Context,
	userID int,
	amount float64,
	currency string,
	transactionType string,
) (*models.Transaction, error) {
	gateway, err := p.gatewaySelector.SelectGatewayForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to select gateway: %w", err)
	}

	transaction := &models.Transaction{
		Amount:    amount,
		Type:      transactionType,
		Status:    "PENDING",
		GatewayID: gateway.ID,
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	if err := p.transactionRepo.Create(ctx, transaction); err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	gatewayDetails, exists := p.gatewayConfig.Gateways[gateway.Name]
	if !exists {
		return nil, fmt.Errorf("gateway %s not found in configuration", gateway.Name)
	}
}
