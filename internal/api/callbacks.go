package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// CallbackProcessorInterface defines the contract for callback processing
type CallbackProcessorInterface interface {
	ProcessCallback(ctx context.Context, gatewayName string, callbackData []byte) error
}

type CallbackHandler struct {
	callbackProcessor CallbackProcessorInterface
}

func NewCallbackHandler(callbackProcessor CallbackProcessorInterface) *CallbackHandler {
	return &CallbackHandler{
		callbackProcessor: callbackProcessor,
	}
}

func (h *CallbackHandler) HandlePayPalCallback(w http.ResponseWriter, r *http.Request) {
	h.handleCallback(w, r, "paypal")
}

func (h *CallbackHandler) HandleStripeCallback(w http.ResponseWriter, r *http.Request) {
	h.handleCallback(w, r, "stripe")
}

func (h *CallbackHandler) HandleAdyenCallback(w http.ResponseWriter, r *http.Request) {
	h.handleCallback(w, r, "ayden")
}

func (h *CallbackHandler) HandleSoapGatewayCallback(w http.ResponseWriter, r *http.Request) {
	h.handleCallback(w, r, "soap-gateway")
}

func (h *CallbackHandler) handleCallback(w http.ResponseWriter, r *http.Request, gatewayName string) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	err = h.callbackProcessor.ProcessCallback(r.Context(), gatewayName, body)
	if err != nil {
		fmt.Printf("Error processing callback from %s: %v\n", gatewayName, err)

		// Return an error response
		http.Error(w, "Failed to process callback: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return a success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]string{
		"status":  "success",
		"message": "Callback processed successfully",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("Error encoding response: %v\n", err)
	}
}
