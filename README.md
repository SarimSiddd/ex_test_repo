# Payment Gateway Integration System

## Project Overview

The Payment Gateway Integration System is a robust and scalable solution designed to facilitate payment processing through multiple third-party payment gateways. The system intelligently selects appropriate payment gateways based on user country and configured gateway priorities, with built-in support for failover mechanisms and resilience strategies.

Key features include:
- Support for multiple payment gateways (PayPal, Stripe, Adyen, and SOAP-based gateways)
- Country-specific gateway prioritization
- Multiple data format support (JSON and XML/SOAP)
- Asynchronous transaction status updates via callbacks
- Fault tolerance with circuit breakers and retry mechanisms
- Secure handling of sensitive payment data

## Architecture

The system follows a clean architecture approach with clear separation of concerns:

```
payment-gateway/
├── cmd/                  # Application entry points
├── db/                   # Database migrations and helpers
├── internal/             # Internal packages
│   ├── api/              # API handlers and routing
│   ├── config/           # Configuration loading and structures
│   ├── gateway/          # Gateway client implementations
│   ├── kafka/            # Kafka integration for event publishing
│   ├── models/           # Data models
│   ├── repository/       # Data access layer
│   │   └── postgres/     # PostgreSQL implementations
│   └── services/         # Business logic
└── docker-compose.yml    # Docker configuration
```

## Key Components

### API Layer
- **Handlers**: Process HTTP requests and responses for deposit, withdrawal, and callbacks
- **Router**: Defines API routes and middleware

### Service Layer
- **TransactionProcessor**: Handles the core logic for processing deposits and withdrawals
- **CallbackProcessor**: Processes gateway callbacks to update transaction status
- **GatewaySelector**: Selects the appropriate gateway based on user's country and configured priorities
- **DataFormatService**: Handles encoding/decoding of different data formats (JSON, XML)
- **FaultTolerance**: Provides circuit breaker and retry mechanisms

### Repository Layer
- **Transaction**: CRUD operations for transactions
- **Gateway**: CRUD operations for payment gateways
- **Country**: CRUD operations for countries
- **User**: CRUD operations for users

### Gateway Layer
- **HTTPClient**: Sends transaction requests to payment gateways

### Configuration
- **GatewayConfig**: Loads and provides access to gateway configuration

## API Endpoints

### `/deposit`
- **Method**: POST
- **Description**: Process deposit transactions
- **Request Format**:
  ```json
  {
    "amount": 100.00,
    "user_id": 1,
    "currency": "EUR"
  }
  ```
- **Response Format**:
  ```json
  {
    "status_code": 200,
    "message": "Deposit initiated successfully",
    "data": {
      "id": 1,
      "amount": 100.00,
      "type": "deposit",
      "status": "PROCESSING",
      "gateway_id": 1,
      "user_id": 1,
      "created_at": "2025-03-07T15:42:00Z"
    }
  }
  ```

### `/withdrawal`
- **Method**: POST
- **Description**: Process withdrawal transactions
- **Request Format**:
  ```json
  {
    "amount": 50.00,
    "user_id": 1,
    "currency": "EUR"
  }
  ```
- **Response Format**:
  ```json
  {
    "status_code": 200,
    "message": "Withdrawal initiated successfully",
    "data": {
      "id": 2,
      "amount": 50.00,
      "type": "withdrawal",
      "status": "PROCESSING",
      "gateway_id": 1,
      "user_id": 1,
      "created_at": "2025-03-07T15:42:00Z"
    }
  }
  ```

### `/api/callbacks/{gateway}`
- **Method**: POST
- **Description**: Endpoint for payment gateways to send transaction status updates
- **Path Parameter**: `gateway` - The name of the gateway sending the callback
- **Request Format** (JSON example):
  ```json
  {
    "transaction_id": 1,
    "status": "COMPLETED"
  }
  ```
- **Request Format** (XML example):
  ```xml
  <callback>
    <transaction_id>1</transaction_id>
    <status>COMPLETED</status>
  </callback>
  ```

## Data Flow

1. **Transaction Initiation**:
   - User submits a deposit or withdrawal request
   - System selects appropriate gateway based on user's country
   - Transaction is created with "PENDING" status
   - Request is sent to the selected payment gateway
   - Transaction status is updated to "PROCESSING"
   - Event is published to Kafka for tracking

2. **Transaction Completion**:
   - Gateway processes the transaction and sends a callback
   - Callback processor updates the transaction status
   - Event is published to Kafka with the updated status

## Fault Tolerance and Resilience

The system implements several fault tolerance mechanisms:

1. **Circuit Breaker**:
   - Prevents cascading failures when a service is down
   - Implemented for Kafka publishing operations
   - Configured with a 5-second interval and 3-second timeout

2. **Retry Mechanism**:
   - Automatically retries failed operations with exponential backoff
   - Configurable max attempts per gateway

3. **Gateway Failover**:
   - Country configuration includes multiple gateways with priority levels
   - System can fall back to lower-priority gateways if needed

## Configuration

### Gateway Configuration (YAML)

The system uses a YAML configuration file to define gateway settings and country-specific priorities:

```yaml
gateways:
  paypal:
    base_url: "https://api.paypal.com"
    endpoints:
      deposit: "/v1/payments/payment"
      withdrawal: "/v1/payments/payouts"
    callback_url: "/api/callbacks/paypal"
    headers:
      Content-Type: "application/json"
      Accept: "application/json"
    timeout: 15  # Seconds
    retry:  
      max_attempts: 3
      backoff_factor: 2  # Exponential backoff factor

countries:
  US:  # United States
    gateways:
      paypal: 10
      stripe: 8
      adyen: 5
```

## Database Schema

The system uses PostgreSQL with the following tables:

1. **gateways**:
   - `id`: Serial primary key
   - `name`: Gateway name (unique)
   - `data_format_supported`: Supported data format
   - `created_at`, `updated_at`: Timestamps

2. **countries**:
   - `id`: Serial primary key
   - `name`: Country name (unique)
   - `code`: 2-character country code (unique)
   - `currency`: 3-character currency code
   - `created_at`, `updated_at`: Timestamps

3. **gateway_countries**:
   - `gateway_id`: Foreign key to gateways
   - `country_id`: Foreign key to countries
   - Primary key: (gateway_id, country_id)

4. **transactions**:
   - `id`: Serial primary key
   - `amount`: Transaction amount
   - `type`: Transaction type (deposit/withdrawal)
   - `status`: Transaction status
   - `created_at`: Timestamp
   - `gateway_id`: Foreign key to gateways
   - `country_id`: Foreign key to countries
   - `user_id`: Foreign key to users

5. **users**:
   - `id`: Serial primary key
   - `username`: User's username (unique)
   - `email`: User's email (unique)
   - `password`: User's password (hashed)
   - `country_id`: Foreign key to countries
   - `created_at`, `updated_at`: Timestamps

## Deployment

The system is containerized using Docker and can be deployed using Docker Compose:

```bash
docker-compose up -d
```

This will start:
- PostgreSQL on port 5432
- Kafka on ports 9092 and 9093
- Redis on port 6379
- Application on port 8080

## Current Limitations

The following features are not yet implemented:

- Dynamic loading of gateway config file
- Validate gateway existence in database from config file
- Fallback based on gateway config
- Security configuration based on gateway
- Kafka as a processor when requests are received
- Clients that act on Kafka's event notification
- Using SQL queries directly instead of Database function
- Expose endpoints via Swagger URL

## Future Improvements

1. **Dynamic Configuration**:
   - Implement dynamic loading of gateway configuration
   - Add support for runtime configuration updates

2. **Enhanced Failover**:
   - Implement more sophisticated failover strategies
   - Add support for gateway health checks

3. **Security Enhancements**:
   - Implement gateway-specific security configurations
   - Add support for more secure encryption algorithms

4. **Monitoring and Observability**:
   - Add metrics collection
   - Implement distributed tracing
   - Create dashboards for system monitoring

5. **Performance Optimization**:
   - Implement caching for frequently accessed data
   - Optimize database queries
   - Add connection pooling

6. **Additional Features**:
   - Support for more payment gateways
   - Implement webhook notifications
   - Add support for recurring payments
