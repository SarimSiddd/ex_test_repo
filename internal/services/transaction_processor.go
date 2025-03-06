package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"payment-gateway/internal/config"
	"payment-gateway/internal/kafka"
	"payment-gateway/internal/models"
	"payment-gateway/internal/repository"
	"strconv"
	"strings"
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

	gatewayDetails, exists := p.gatewayConfig.Gateways[gateway.Name]
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
		return p.sendToGateway(ctx, gatewayDetails, transactionType, payload, transaction.ID)
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

// This can be moved somewhere else
func (p *TransactionProcessor) sendToGateway(
	ctx context.Context,
	gatewayDetails config.GatewayDetails,
	transactionType string,
	payload []byte,
	transactionID int,
) error {
	var endpoint string
	if transactionType == "deposit" {
		endpoint = gatewayDetails.Endpoints.Deposit
	} else {
		endpoint = gatewayDetails.Endpoints.Withdrawal
	}

	url := gatewayDetails.BaseURL + endpoint

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range gatewayDetails.Headers {
		req.Header.Set(key, value)
	}

	req.Header.Set("X-Transaction-ID", strconv.Itoa(transactionID))

	callbackURL := gatewayDetails.CallbackURL
	if callbackURL != "" {
		if !strings.HasPrefix(callbackURL, "http") {
			baseURL := "https://api-domain.com" // Can read from config
			callbackURL = baseURL + callbackURL
		}
		req.Header.Set("X-Callback-URL", callbackURL)
	}

	client := &http.Client{
		Timeout: time.Duration(gatewayDetails.Timeout) * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("gateway returned non-success status: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}
