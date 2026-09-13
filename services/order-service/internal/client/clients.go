package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var ErrOutOfStock = errors.New("out of stock")
var ErrSKUNotFound = errors.New("sku not found")
var ErrPaymentDeclined = errors.New("payment declined")

// httpClient is shared by both clients with an explicit timeout — the
// stdlib default client has NO timeout, which is a classic way for one
// slow downstream to hang an entire service.
var httpClient = &http.Client{Timeout: 5 * time.Second}

type InventoryClient struct {
	baseURL string
}

func NewInventoryClient(baseURL string) *InventoryClient {
	return &InventoryClient{baseURL: baseURL}
}

func (c *InventoryClient) Reserve(sku string, qty int) error {
	body, _ := json.Marshal(map[string]any{"sku": sku, "qty": qty})

	resp, err := httpClient.Post(c.baseURL+"/inventory/reserve", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("calling inventory-service: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusConflict:
		return ErrOutOfStock
	case http.StatusNotFound:
		return ErrSKUNotFound
	default:
		return fmt.Errorf("inventory-service returned status %d", resp.StatusCode)
	}
}

type PaymentClient struct {
	baseURL string
}

func NewPaymentClient(baseURL string) *PaymentClient {
	return &PaymentClient{baseURL: baseURL}
}

func (c *PaymentClient) Charge(orderID string, amount float64) (string, error) {
	body, _ := json.Marshal(map[string]any{"order_id": orderID, "amount": amount})

	resp, err := httpClient.Post(c.baseURL+"/payments", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("calling payment-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusPaymentRequired {
		return "", ErrPaymentDeclined
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("payment-service returned status %d", resp.StatusCode)
	}

	var parsed struct {
		TransactionID string `json:"transaction_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", fmt.Errorf("decoding payment-service response: %w", err)
	}
	return parsed.TransactionID, nil
}
