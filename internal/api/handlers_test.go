package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"payment-gateway/internal/api"
	"payment-gateway/internal/models"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type mockTransactionProcessor struct {
	mockProcessDeposit    func(ctx context.Context, userID int, amount float64, currency string) (*models.Transaction, error)
	mockProcessWithdrawal func(ctx context.Context, userID int, amount float64, currency string) (*models.Transaction, error)
}

func (m *mockTransactionProcessor) ProcessDeposit(ctx context.Context, userID int, amount float64, currency string) (*models.Transaction, error) {
	return m.mockProcessDeposit(ctx, userID, amount, currency)
}

func (m *mockTransactionProcessor) ProcessWithdrawal(ctx context.Context, userID int, amount float64, currency string) (*models.Transaction, error) {
	return m.mockProcessWithdrawal(ctx, userID, amount, currency)
}

func TestDepositHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		setupMock      func(m *mockTransactionProcessor)
		expectedStatus int
		expectedBody   string
		checkResponse  func(t *testing.T, response *httptest.ResponseRecorder)
	}{
		{
			name:        "Success",
			requestBody: `{"amount": 100.00, "user_id": 1, "currency": "EUR"}`,
			setupMock: func(m *mockTransactionProcessor) {
				m.mockProcessDeposit = func(ctx context.Context, userID int, amount float64, currency string) (*models.Transaction, error) {
					return &models.Transaction{
						ID:        123,
						Amount:    100.00,
						Type:      "deposit",
						Status:    "PROCESSING",
						UserID:    1,
						CreatedAt: time.Now(),
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response *httptest.ResponseRecorder) {
				var resp models.APIResponse
				err := json.NewDecoder(response.Body).Decode(&resp)
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, resp.StatusCode)
				require.Equal(t, "Deposit initiated successfully", resp.Message)

				// When unmarshaling JSON, the Data field becomes map[string]interface{}, not *models.Transaction
				t.Logf("resp.Data type: %T, value: %+v", resp.Data, resp.Data)

				// Access the data as a map
				dataMap, ok := resp.Data.(map[string]interface{})
				require.True(t, ok, "resp.Data should be a map[string]interface{}")

				// Check the transaction fields in the map
				require.Equal(t, float64(123), dataMap["ID"], "Transaction ID should be 123")
				require.Equal(t, float64(100.00), dataMap["Amount"], "Transaction Amount should be 100.00")
				require.Equal(t, "deposit", dataMap["Type"], "Transaction Type should be 'deposit'")
				require.Equal(t, "PROCESSING", dataMap["Status"], "Transaction Status should be 'PROCESSING'")
				require.Equal(t, float64(1), dataMap["UserID"], "Transaction UserID should be 1")
			},
		},
		{
			name:           "Invalid JSON",
			requestBody:    `{"amount": "invalid", "user_id": 1, "currency": "EUR"}`,
			setupMock:      func(m *mockTransactionProcessor) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, response *httptest.ResponseRecorder) {
				require.Contains(t, response.Body.String(), "Invalid request body")
			},
		},
		{
			name:        "Service Error",
			requestBody: `{"amount": 100.00, "user_id": 1, "currency": "EUR"}`,
			setupMock: func(m *mockTransactionProcessor) {
				m.mockProcessDeposit = func(ctx context.Context, userID int, amount float64, currency string) (*models.Transaction, error) {
					return nil, errors.New("service error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, response *httptest.ResponseRecorder) {
				require.Contains(t, response.Body.String(), "Failed to process deposit")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock
			mockProcessor := &mockTransactionProcessor{}
			tt.setupMock(mockProcessor)

			// Create handler with mock
			handler := api.NewTransactionHandler(mockProcessor)

			// Create request
			req, err := http.NewRequest("POST", "/deposit", strings.NewReader(tt.requestBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.DepositHandler(rr, req)

			// Check status code
			require.Equal(t, tt.expectedStatus, rr.Code)

			// Additional response checks
			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
		})
	}
}

func TestWithdrawalHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		setupMock      func(m *mockTransactionProcessor)
		expectedStatus int
		checkResponse  func(t *testing.T, response *httptest.ResponseRecorder)
	}{
		{
			name:        "Success",
			requestBody: `{"amount": 100.00, "user_id": 1, "currency": "EUR"}`,
			setupMock: func(m *mockTransactionProcessor) {
				m.mockProcessWithdrawal = func(ctx context.Context, userID int, amount float64, currency string) (*models.Transaction, error) {
					return &models.Transaction{
						ID:        456,
						Amount:    100.00,
						Type:      "withdrawal",
						Status:    "PROCESSING",
						UserID:    1,
						CreatedAt: time.Now(),
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response *httptest.ResponseRecorder) {
				var resp models.APIResponse
				err := json.NewDecoder(response.Body).Decode(&resp)
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, resp.StatusCode)
				require.Equal(t, "Withdrawal initiated successfully", resp.Message)

				// When unmarshaling JSON, the Data field becomes map[string]interface{}, not *models.Transaction
				t.Logf("resp.Data type: %T, value: %+v", resp.Data, resp.Data)

				// Access the data as a map
				dataMap, ok := resp.Data.(map[string]interface{})
				require.True(t, ok, "resp.Data should be a map[string]interface{}")

				// Check the transaction fields in the map
				require.Equal(t, float64(456), dataMap["ID"], "Transaction ID should be 456")
				require.Equal(t, float64(100.00), dataMap["Amount"], "Transaction Amount should be 100.00")
				require.Equal(t, "withdrawal", dataMap["Type"], "Transaction Type should be 'withdrawal'")
				require.Equal(t, "PROCESSING", dataMap["Status"], "Transaction Status should be 'PROCESSING'")
				require.Equal(t, float64(1), dataMap["UserID"], "Transaction UserID should be 1")
			},
		},
		{
			name:           "Invalid JSON",
			requestBody:    `{"amount": "invalid", "user_id": 1, "currency": "EUR"}`,
			setupMock:      func(m *mockTransactionProcessor) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, response *httptest.ResponseRecorder) {
				require.Contains(t, response.Body.String(), "Invalid request body")
			},
		},
		{
			name:        "Service Error",
			requestBody: `{"amount": 100.00, "user_id": 1, "currency": "EUR"}`,
			setupMock: func(m *mockTransactionProcessor) {
				m.mockProcessWithdrawal = func(ctx context.Context, userID int, amount float64, currency string) (*models.Transaction, error) {
					return nil, errors.New("service error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, response *httptest.ResponseRecorder) {
				require.Contains(t, response.Body.String(), "Failed to process withdrawal")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock
			mockProcessor := &mockTransactionProcessor{}
			tt.setupMock(mockProcessor)

			// Create handler with mock
			handler := api.NewTransactionHandler(mockProcessor)

			// Create request
			req, err := http.NewRequest("POST", "/withdrawal", strings.NewReader(tt.requestBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.WithdrawalHandler(rr, req)

			// Check status code
			require.Equal(t, tt.expectedStatus, rr.Code)

			// Additional response checks
			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
		})
	}
}
