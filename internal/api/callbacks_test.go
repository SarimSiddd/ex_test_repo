package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"payment-gateway/internal/api"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Mock implementation of the CallbackProcessor
type mockCallbackProcessor struct {
	mockProcessCallback func(ctx context.Context, gatewayName string, callbackData []byte) error
}

func (m *mockCallbackProcessor) ProcessCallback(ctx context.Context, gatewayName string, callbackData []byte) error {
	return m.mockProcessCallback(ctx, gatewayName, callbackData)
}

func TestHandlePayPalCallback(t *testing.T) {
	runCallbackTest(t, "paypal", func(handler *api.CallbackHandler, req *http.Request, rr *httptest.ResponseRecorder) {
		handler.HandlePayPalCallback(rr, req)
	})
}

func TestHandleStripeCallback(t *testing.T) {
	runCallbackTest(t, "stripe", func(handler *api.CallbackHandler, req *http.Request, rr *httptest.ResponseRecorder) {
		handler.HandleStripeCallback(rr, req)
	})
}

func TestHandleAdyenCallback(t *testing.T) {
	runCallbackTest(t, "ayden", func(handler *api.CallbackHandler, req *http.Request, rr *httptest.ResponseRecorder) {
		handler.HandleAdyenCallback(rr, req)
	})
}

func TestHandleSoapGatewayCallback(t *testing.T) {
	runCallbackTest(t, "soap-gateway", func(handler *api.CallbackHandler, req *http.Request, rr *httptest.ResponseRecorder) {
		handler.HandleSoapGatewayCallback(rr, req)
	})
}

// Helper function to run tests for all callback handlers
func runCallbackTest(t *testing.T, expectedGatewayName string, handlerFunc func(handler *api.CallbackHandler, req *http.Request, rr *httptest.ResponseRecorder)) {
	tests := []struct {
		name           string
		requestBody    string
		setupMock      func(m *mockCallbackProcessor)
		expectedStatus int
		checkResponse  func(t *testing.T, response *httptest.ResponseRecorder)
	}{
		{
			name:        "Success",
			requestBody: `{"transaction_id": 123, "status": "COMPLETED"}`,
			setupMock: func(m *mockCallbackProcessor) {
				m.mockProcessCallback = func(ctx context.Context, gatewayName string, callbackData []byte) error {
					// Verify the gateway name is correctly passed
					require.Equal(t, expectedGatewayName, gatewayName)

					// Verify callback data is correctly passed
					var data map[string]interface{}
					err := json.Unmarshal(callbackData, &data)
					require.NoError(t, err)
					require.Equal(t, float64(123), data["transaction_id"])
					require.Equal(t, "COMPLETED", data["status"])

					return nil
				}
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response *httptest.ResponseRecorder) {
				var resp map[string]string
				err := json.NewDecoder(response.Body).Decode(&resp)
				require.NoError(t, err)
				require.Equal(t, "success", resp["status"])
				require.Equal(t, "Callback processed successfully", resp["message"])
			},
		},
		{
			name:        "Processor Error",
			requestBody: `{"transaction_id": 123, "status": "COMPLETED"}`,
			setupMock: func(m *mockCallbackProcessor) {
				m.mockProcessCallback = func(ctx context.Context, gatewayName string, callbackData []byte) error {
					return errors.New("processor error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, response *httptest.ResponseRecorder) {
				require.Contains(t, response.Body.String(), "Failed to process callback: processor error")
			},
		},
		{
			name:        "Empty Body",
			requestBody: ``,
			setupMock: func(m *mockCallbackProcessor) {
				m.mockProcessCallback = func(ctx context.Context, gatewayName string, callbackData []byte) error {
					t.Fatalf("ProcessCallback should not be called with empty body")
					return nil
				}
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, response *httptest.ResponseRecorder) {
				require.Contains(t, response.Body.String(), "Failed to read request body")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock
			mockProcessor := &mockCallbackProcessor{}
			tt.setupMock(mockProcessor)

			// Create handler with mock
			handler := api.NewCallbackHandler(mockProcessor)

			// Create request
			var req *http.Request
			var err error

			if tt.requestBody == "" {
				// For empty body test, use a request that will fail when reading body
				req, err = http.NewRequest("POST", "/callback", errorReader{})
			} else {
				req, err = http.NewRequest("POST", "/callback", strings.NewReader(tt.requestBody))
			}

			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handlerFunc(handler, req, rr)

			// Check status code
			require.Equal(t, tt.expectedStatus, rr.Code)

			// Additional response checks
			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
		})
	}
}

// errorReader is a custom io.Reader that always returns an error
type errorReader struct{}

func (e errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("forced read error")
}
