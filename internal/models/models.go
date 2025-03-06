package models

import "time"

type User struct {
	ID        int       `json:"id" xml:"id"`
	Username  string    `json:"username" xml:"username"`
	Email     string    `json:"email" xml:"email"`
	CountryID int       `json:"country_id" xml:"country_id"`
	CreatedAt time.Time `json:"created_at" xml:"created_at"`
	UpdatedAt time.Time `json:"updated_at" xml:"updated_at"`
}

type Gateway struct {
	ID                  int       `json:"id" xml:"id"`
	Name                string    `json:"name" xml:"name"`
	DataFormatSupported string    `json:"data_format_supported" xml:"data_format_supported"`
	CreatedAt           time.Time `json:"created_at" xml:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" xml:"updated_at"`
}

type Transaction struct {
	ID        int
	Amount    float64
	Type      string
	Status    string
	GatewayID int
	CountryID int
	UserID    int
	CreatedAt time.Time
}

type Country struct {
	ID        int       `json:"id" xml:"id"`
	Name      string    `json:"name" xml:"name"`
	Code      string    `json:"code" xml:"code"`
	Currency  string    `json:"currency" xml:"currency"`
	CreatedAt time.Time `json:"created_at" xml:"created_at"`
	UpdatedAt time.Time `json:"updated_at" xml:"updated_at"`
}

// a standard request structure for the transactions
type TransactionRequest struct {
	Amount   float64 `json:"amount"`
	UserID   int     `json:"user_id"`
	Currency string  `json:"currency"`
}

// a standard response structure for the APIs
type APIResponse struct {
	StatusCode int         `json:"status_code" xml:"status_code"`
	Message    string      `json:"message" xml:"message"`
	Data       interface{} `json:"data,omitempty" xml:"data,omitempty"`
}
