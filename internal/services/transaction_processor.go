package services

import (
	"context"
	"encoding/json"
	"fmt"
	"payment-gateway/internal/config"
	"payment-gateway/internal/kafka"
	"payment-gateway/internal/models"
	"payment-gateway/internal/repository"
	"strconv"
	"time"
)

// Client defines the interface for gateway communication
type Client interface {
	SendTransaction(ctx context.Context, transactionType string, payload []byte, transactionID int, gatewayDetails config.GatewayDetails) error
}

// GatewayConfigProvider defines the interface for accessing gateway configuration
type GatewayConfigProvider interface {
	// GetGatewayDetails returns the gateway details for a given gateway name
	GetGatewayDetails(gatewayName string) (config.GatewayDetails, bool)
}

// GatewaySelectorProvider defines the interface for selecting gateways
type GatewaySelectorProvider interface {
	// SelectGatewayForUser selects a gateway for a given user
	SelectGatewayForUser(ctx context.Context, userID int) (*models.Gateway, error)
}

type TransactionProcessor struct {
	gatewayConfig   GatewayConfigProvider
	gatewaySelector GatewaySelectorProvider
	transactionRepo repository.Transaction
	gatewayClient   Client
}

func NewTransactionProcessor(
	gatewayConfig GatewayConfigProvider,
	gatewaySelector GatewaySelectorProvider,
	transactionRepo repository.Transaction,
	gatewayClient Client,
) *TransactionProcessor {
	return &TransactionProcessor{
		gatewayConfig:   gatewayConfig,
		gatewaySelector: gatewaySelector,
		transactionRepo: transactionRepo,
		gatewayClient:   gatewayClient,
	}
}

func (p *TransactionProcessor) ProcessDeposit(ctx context.Context, userID int, amount float64, currency string) (*models.Transaction, error) {
	return p.processTransaction(ctx, userID, amount, currency, "deposit")
}

func (p *TransactionProcessor) ProcessWithdrawal(ctx context.Context, userID int, amount float64, currency string) (*models.Transaction, error) {
	return p.processTransaction(ctx, userID, amount, currency, "withdrawal")
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

	gatewayDetails, exists := p.gatewayConfig.GetGatewayDetails(gateway.Name)
	if !exists {
		return nil, fmt.Errorf("gateway %s not found in configuration", gateway.Name)
	}

	dataFormat := gateway.DataFormatSupported
	payload, err := PrepareTransactionPayload(transaction, currency, dataFormat)
	if err != nil {
		p.transactionRepo.UpdateStatus(ctx, transaction.ID, "FAILED")
		return nil, fmt.Errorf("failed to prepare payload: %w", err)
	}

	err = RetryOperation(func() error {
		return p.gatewayClient.SendTransaction(ctx, transactionType, payload, transaction.ID, gatewayDetails)
	}, gatewayDetails.Retry.MaxAttempts)

	if err != nil {
		p.transactionRepo.UpdateStatus(ctx, transaction.ID, "FAILED")
		return nil, fmt.Errorf("failed to send request to gateway: %w", err)
	}

	if err := p.transactionRepo.UpdateStatus(ctx, transaction.ID, "PROCESSING"); err != nil {
		return nil, fmt.Errorf("failed to update transaction status: %w", err)
	}

	err = PublishWithCircuitBreaker(func() error {
		return p.publishTransactionEvent(ctx, transaction, dataFormat)
	})

	if err != nil {
		fmt.Printf("failed to publish transaction event: %v\n", err)
	}

	return transaction, nil
}

func (p *TransactionProcessor) publishTransactionEvent(
	ctx context.Context,
	transaction *models.Transaction,
	dataFormat string,
) error {
	message := map[string]interface{}{
		"transaction_id": transaction.ID,
		"status":         transaction.Status,
		"type":           transaction.Type,
		"amount":         transaction.Amount,
		"gateway_id":     transaction.GatewayID,
		"user_id":        transaction.UserID,
		"timestamp":      time.Now().Unix(),
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return kafka.PublishTransaction(ctx, strconv.Itoa(transaction.ID), messageBytes, dataFormat)
}
