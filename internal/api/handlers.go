package api

import (
	"context"
	"encoding/json"
	"net/http"
	"payment-gateway/internal/models"
)

type TransactionProcessorInterface interface {
	ProcessDeposit(ctx context.Context, userID int, amount float64, currency string) (*models.Transaction, error)
	ProcessWithdrawal(ctx context.Context, userID int, amount float64, currency string) (*models.Transaction, error)
}
type TransactionHandler struct {
	transactionProcessor TransactionProcessorInterface
}

func NewTransactionHandler(transactionProcessor TransactionProcessorInterface) *TransactionHandler {
	return &TransactionHandler{
		transactionProcessor: transactionProcessor,
	}
}

// DepositHandler handles deposit requests (feel free to update how user is passed to the request)
// Sample Request (POST /deposit):
//
//	{
//	    "amount": 100.00,
//	    "user_id": 1,
//	    "currency": "EUR"
//	}
func (h *TransactionHandler) DepositHandler(w http.ResponseWriter, r *http.Request) {
	var req models.TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	transaction, err := h.transactionProcessor.ProcessDeposit(r.Context(), req.UserID, req.Amount, req.Currency)
	if err != nil {
		http.Error(w, "Failed to process deposit: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := models.APIResponse{
		StatusCode: http.StatusOK,
		Message:    "Deposit initiated successfully",
		Data:       transaction,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// WithdrawalHandler handles withdrawal requests (feel free to update how user is passed to the request)
// Sample Request (POST /deposit):
//
//	{
//	    "amount": 100.00,
//	    "user_id": 1,
//	    "currency": "EUR"
//	}
func (h *TransactionHandler) WithdrawalHandler(w http.ResponseWriter, r *http.Request) {
	var req models.TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	transaction, err := h.transactionProcessor.ProcessWithdrawal(r.Context(), req.UserID, req.Amount, req.Currency)
	if err != nil {
		http.Error(w, "Failed to process withdrawal: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := models.APIResponse{
		StatusCode: http.StatusOK,
		Message:    "Withdrawal initiated successfully",
		Data:       transaction,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
