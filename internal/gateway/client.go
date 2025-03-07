package gateway

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"payment-gateway/internal/config"
	"strconv"
	"strings"
	"time"
)

type HTTPClient struct {
}

func NewHTTPClient() *HTTPClient {
	return &HTTPClient{}
}

func (c *HTTPClient) SendTransaction(
	ctx context.Context,
	transactionType string,
	payload []byte,
	transactionID int,
	gatewayDetails config.GatewayDetails,
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
