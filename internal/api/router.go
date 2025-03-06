package api

import (
	"github.com/gorilla/mux"
)

func SetupRouter(transactionHandler *TransactionHandler, callbackHandler *CallbackHandler) *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/deposit", transactionHandler.DepositHandler).Methods("POST")
	router.HandleFunc("/withdrawal", transactionHandler.WithdrawalHandler).Methods("POST")

	router.HandleFunc("/api/callbacks/paypal", callbackHandler.HandlePayPalCallback).Methods("POST")
	router.HandleFunc("/api/callbacks/stripe", callbackHandler.HandleStripeCallback).Methods("POST")
	router.HandleFunc("/api/callbacks/adyen", callbackHandler.HandleAdyenCallback).Methods("POST")
	router.HandleFunc("/api/callbacks/soap-gateway", callbackHandler.HandleSoapGatewayCallback).Methods("POST")

	return router
}
