package services

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"payment-gateway/internal/models"
)

// decodes the incoming request based on content type
func DecodeRequest(r *http.Request, request *models.TransactionRequest) error {
	contentType := r.Header.Get("Content-Type")

	switch contentType {
	case "application/json":
		return json.NewDecoder(r.Body).Decode(request)
	case "text/xml":
		return xml.NewDecoder(r.Body).Decode(request)
	case "application/xml":
		return xml.NewDecoder(r.Body).Decode(request)
	default:
		return fmt.Errorf("unsupported content type")
	}
}

func EncodePayload(data interface{}, dataFormat string) ([]byte, error) {
	switch dataFormat {
	case "application/json":
		return json.Marshal(data)
	case "text/xml", "application/xml":
		return xml.Marshal(data)
	default:
		return nil, fmt.Errorf("unsupported data format: %s", dataFormat)
	}
}

func PrepareTransactionPayload(
	transaction *models.Transaction,
	currency string,
	dataFormat string,
) ([]byte, error) {

	if dataFormat == "application/json" {
		payloadData := map[string]interface{}{
			"transaction_id": transaction.ID,
			"amount":         transaction.Amount,
			"currency":       currency,
			"type":           transaction.Type,
		}
		return EncodePayload(payloadData, dataFormat)
	}

	if dataFormat == "text/xml" || dataFormat == "application/xml" {
		type XMLPayload struct {
			TransactionID int     `xml:"transaction_id"`
			Amount        float64 `xml:"amount"`
			Currency      string  `xml:"currency"`
			Type          string  `xml:"type"`
		}

		xmlPayload := XMLPayload{
			TransactionID: transaction.ID,
			Amount:        transaction.Amount,
			Currency:      currency,
			Type:          transaction.Type,
		}

		return EncodePayload(xmlPayload, dataFormat)
	}

	return nil, fmt.Errorf("unsupported data format: %s", dataFormat)
}
