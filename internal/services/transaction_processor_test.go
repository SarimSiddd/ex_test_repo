package services_test

import (
	"context"
	"fmt"
	"payment-gateway/internal/config"
	"payment-gateway/internal/models"
	"payment-gateway/internal/services"
	"testing"
)

// Mock implementation of the GatewayConfigProvider
type mockGatewayConfigProvider struct {
	mockGetGatewayDetails func(gatewayName string) (config.GatewayDetails, bool)
}

func (m *mockGatewayConfigProvider) GetGatewayDetails(gatewayName string) (config.GatewayDetails, bool) {
	return m.mockGetGatewayDetails(gatewayName)
}

// Mock implementation of the GatewaySelectorProvider
type mockGatewaySelectorProvider struct {
	mockSelectGatewayForUser func(ctx context.Context, userID int) (*models.Gateway, error)
}

func (m *mockGatewaySelectorProvider) SelectGatewayForUser(ctx context.Context, userID int) (*models.Gateway, error) {
	return m.mockSelectGatewayForUser(ctx, userID)
}

// Mock implementation of the Transaction repository
type mockTransactionRepo struct {
	mockCreate       func(ctx context.Context, transaction *models.Transaction) error
	mockUpdateStatus func(ctx context.Context, transactionID int, status string) error
	mockGetByID      func(ctx context.Context, transactionID int) (*models.Transaction, error)
}

func (m *mockTransactionRepo) Create(ctx context.Context, transaction *models.Transaction) error {
	return m.mockCreate(ctx, transaction)
}

func (m *mockTransactionRepo) UpdateStatus(ctx context.Context, transactionID int, status string) error {
	return m.mockUpdateStatus(ctx, transactionID, status)
}

func (m *mockTransactionRepo) GetByID(ctx context.Context, transactionID int) (*models.Transaction, error) {
	return m.mockGetByID(ctx, transactionID)
}

// Mock implementation of the Client
type mockClient struct {
	mockSendTransaction func(ctx context.Context, transactionType string, payload []byte, transactionID int, gatewayDetails config.GatewayDetails) error
}

func (m *mockClient) SendTransaction(ctx context.Context, transactionType string, payload []byte, transactionID int, gatewayDetails config.GatewayDetails) error {
	return m.mockSendTransaction(ctx, transactionType, payload, transactionID, gatewayDetails)
}

func TestProcessDeposit(t *testing.T) {
	runTransactionTest(t, "deposit", func(processor *services.TransactionProcessor, ctx context.Context, userID int, amount float64, currency string) (*models.Transaction, error) {
		return processor.ProcessDeposit(ctx, userID, amount, currency)
	})
}

func TestProcessWithdrawal(t *testing.T) {
	runTransactionTest(t, "withdrawal", func(processor *services.TransactionProcessor, ctx context.Context, userID int, amount float64, currency string) (*models.Transaction, error) {
		return processor.ProcessWithdrawal(ctx, userID, amount, currency)
	})
}

// runTransactionTest is a helper function to test transaction processing
func runTransactionTest(
	t *testing.T,
	transactionType string,
	processFunc func(processor *services.TransactionProcessor, ctx context.Context, userID int, amount float64, currency string) (*models.Transaction, error),
) {
	// Test data
	userID := 123
	amount := 100.0
	currency := "USD"
	gatewayID := 1
	gatewayName := "stripe"
	transactionID := 456
	dataFormat := "application/json"

	// Create mock gateway details
	gatewayDetails := config.GatewayDetails{
		BaseURL: "https://api.stripe.com",
		Endpoints: config.GatewayEndpoints{
			Deposit:    "/v1/deposits",
			Withdrawal: "/v1/withdrawals",
		},
		CallbackURL: "https://myapp.com/callbacks",
		Headers: map[string]string{
			"Authorization": "Bearer sk_test_123",
		},
		Timeout: 30,
		Retry: config.GatewayRetry{
			MaxAttempts:   3,
			BackoffFactor: 2.0,
		},
	}

	// Setup mocks
	mockGatewayConfig := &mockGatewayConfigProvider{
		mockGetGatewayDetails: func(name string) (config.GatewayDetails, bool) {
			if name == gatewayName {
				return gatewayDetails, true
			}
			return config.GatewayDetails{}, false
		},
	}

	mockGatewaySelector := &mockGatewaySelectorProvider{
		mockSelectGatewayForUser: func(ctx context.Context, uid int) (*models.Gateway, error) {
			if uid == userID {
				return &models.Gateway{
					ID:                  gatewayID,
					Name:                gatewayName,
					DataFormatSupported: dataFormat,
				}, nil
			}
			return nil, fmt.Errorf("gateway not found for user")
		},
	}

	mockTransactionRepo := &mockTransactionRepo{
		mockCreate: func(ctx context.Context, transaction *models.Transaction) error {
			// Set the ID for the transaction
			transaction.ID = transactionID
			return nil
		},
		mockUpdateStatus: func(ctx context.Context, tid int, status string) error {
			if tid != transactionID {
				return fmt.Errorf("transaction not found")
			}
			return nil
		},
		mockGetByID: func(ctx context.Context, tid int) (*models.Transaction, error) {
			if tid != transactionID {
				return nil, fmt.Errorf("transaction not found")
			}
			return &models.Transaction{
				ID:        transactionID,
				Amount:    amount,
				Type:      transactionType,
				Status:    "PROCESSING",
				GatewayID: gatewayID,
				UserID:    userID,
			}, nil
		},
	}

	mockClient := &mockClient{
		mockSendTransaction: func(ctx context.Context, tType string, payload []byte, tid int, gDetails config.GatewayDetails) error {
			if tType != transactionType {
				return fmt.Errorf("unexpected transaction type: %s", tType)
			}
			if tid != transactionID {
				return fmt.Errorf("unexpected transaction ID: %d", tid)
			}
			return nil
		},
	}

	// Create the processor with mocks
	processor := services.NewTransactionProcessor(
		mockGatewayConfig,
		mockGatewaySelector,
		mockTransactionRepo,
		mockClient,
	)

	// Process the transaction
	ctx := context.Background()
	transaction, err := processFunc(processor, ctx, userID, amount, currency)

	// Assertions
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if transaction == nil {
		t.Fatal("Expected transaction to be returned, got nil")
	}

	if transaction.ID != transactionID {
		t.Errorf("Expected transaction ID %d, got %d", transactionID, transaction.ID)
	}

	if transaction.Amount != amount {
		t.Errorf("Expected amount %.2f, got %.2f", amount, transaction.Amount)
	}

	if transaction.Type != transactionType {
		t.Errorf("Expected transaction type %s, got %s", transactionType, transaction.Type)
	}

	if transaction.Status != "PENDING" {
		t.Errorf("Expected status PENDING, got %s", transaction.Status)
	}

	if transaction.GatewayID != gatewayID {
		t.Errorf("Expected gateway ID %d, got %d", gatewayID, transaction.GatewayID)
	}

	if transaction.UserID != userID {
		t.Errorf("Expected user ID %d, got %d", userID, transaction.UserID)
	}
}
