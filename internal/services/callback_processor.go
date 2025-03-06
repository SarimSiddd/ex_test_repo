package services

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"payment-gateway/internal/kafka"
	"payment-gateway/internal/models"
	"payment-gateway/internal/repository"
	"strconv"
	"time"
)

type CallbackProcessor struct {
	transactionRepo repository.Transaction
	gatewayRepo     repository.Gateway
}

func NewCallbackProcessor(
	transactionRepo repository.Transaction,
	gatewayRepo repository.Gateway,
) *CallbackProcessor {
	return &CallbackProcessor{
		transactionRepo: transactionRepo,
		gatewayRepo:     gatewayRepo,
	}
}

func (p *CallbackProcessor) ProcessCallback(ctx context.Context, gatewayName string, callbackData []byte) error {
	transactionID, status, err := p.parseCallbackData(ctx, gatewayName, callbackData)
	if err != nil {
		return fmt.Errorf("failed to parse callback data: %w", err)
	}

	if err := p.transactionRepo.UpdateStatus(ctx, transactionID, status); err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	transaction, err := p.transactionRepo.GetByID(ctx, transactionID)
	if err != nil {
		return fmt.Errorf("failed to get transaction: %w", err)
	}

	gateway, err := p.gatewayRepo.FindByID(ctx, transaction.GatewayID)
	if err != nil {
		return fmt.Errorf("failed to find gateway: %w", err)
	}

	err = PublishWithCircuitBreaker(func() error {
		return p.publishTransactionEvent(ctx, transaction, gateway.DataFormatSupported)
	})

	if err != nil {
		fmt.Printf("failed to publish transaction update event: %v\n", err)
	}

	return nil
}

func (p *CallbackProcessor) parseCallbackData(ctx context.Context, gatewayName string, callbackData []byte) (int, string, error) {
	gateway, err := p.gatewayRepo.FindByName(ctx, gatewayName)
	if err != nil {
		return 0, "", fmt.Errorf("failed to find gateway: %w", err)
	}

	var transactionID int
	var status string

	switch gateway.DataFormatSupported {
	case "application/json":
		var callbackJSON map[string]interface{}
		if err := json.Unmarshal(callbackData, &callbackJSON); err != nil {
			return 0, "", fmt.Errorf("failed to parse JSON callback data: %w", err)
		}

		if txID, ok := callbackJSON["transaction_id"].(float64); ok {
			transactionID = int(txID)
		} else {
			return 0, "", fmt.Errorf("invalid transaction ID in callback")
		}

		if txStatus, ok := callbackJSON["status"].(string); ok {
			status = txStatus
		} else {
			return 0, "", fmt.Errorf("invalid status in callback")
		}

	case "text/xml", "application/xml":
		type XMLCallback struct {
			TransactionID int    `xml:"transaction_id"`
			Status        string `xml:"status"`
		}

		var callbackXML XMLCallback
		if err := xml.Unmarshal(callbackData, &callbackXML); err != nil {
			return 0, "", fmt.Errorf("failed to parse XML callback data: %w", err)
		}

		transactionID = callbackXML.TransactionID
		status = callbackXML.Status

	default:
		return 0, "", fmt.Errorf("unsupported data format: %s", gateway.DataFormatSupported)
	}

	return transactionID, status, nil
}

func (p *CallbackProcessor) publishTransactionEvent(
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
		"event_type":     "callback_processed",
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return kafka.PublishTransaction(ctx, strconv.Itoa(transaction.ID), messageBytes, dataFormat)
}
