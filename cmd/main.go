package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"payment-gateway/db"
	"payment-gateway/internal/api"
	"payment-gateway/internal/config"
	"payment-gateway/internal/gateway"
	"payment-gateway/internal/repository/postgres"
	"payment-gateway/internal/services"

	"github.com/gorilla/mux"
)

func main() {

	// Initialize the database connection
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")

	dbURL := "postgres://" + dbUser + ":" + dbPassword + "@" + dbHost + ":" + dbPort + "/" + dbName + "?sslmode=disable"

	db.InitializeDB(dbURL)
	database, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}
	defer database.Close()

	// Set up the HTTP server and routes
	gatewayConfigPath := os.Getenv("GATEWAY_CONFIG_PATH")
	if gatewayConfigPath == "" {
		gatewayConfigPath = "internal/config/gateway_config.yaml"
	}

	gatewayConfig, err := config.LoadGatewayConfig(gatewayConfigPath)
	if err != nil {
		log.Fatalf("Failed to load gateway configuration: %v", err)
	}

	// Initialize repositories
	router := initializeRepositories(database, gatewayConfig)

	// Start the server on port 8080
	log.Println("Starting server on port 8080...")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}

}

func initializeRepositories(database *sql.DB, gatewayConfig *config.GatewayConfig) *mux.Router {

	transactionRepo := postgres.NewTransactionRepo(database)
	gatewayRepo := postgres.NewGatewayRepo(database)
	countryRepo := postgres.NewCountryRepo(database)
	userRepo := postgres.NewUserRepo(database)

	gatewaySelector := services.NewGatewaySelector(gatewayConfig, countryRepo, gatewayRepo, userRepo)

	// Initialize the gateway client
	gatewayClient := &gateway.HTTPClient{}

	transactionProcessor := services.NewTransactionProcessor(
		gatewayConfig,
		gatewaySelector,
		transactionRepo,
		gatewayClient,
	)

	transactionHandler := api.NewTransactionHandler(
		transactionProcessor,
	)

	callbackProcessor := services.NewCallbackProcessor(
		transactionRepo,
		gatewayRepo,
	)

	callbackHandler := api.NewCallbackHandler(callbackProcessor)

	router := api.SetupRouter(transactionHandler, callbackHandler)

	return router
}
